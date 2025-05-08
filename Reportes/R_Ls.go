package Reportes

import (
	"Proyecto/Environment"
	"Proyecto/Partitions"
	"Proyecto/User_Groups"
	"Proyecto/Utils"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

func GenerarReporteLS(pathFileLs string, outputPath string, id string) string {
	var output strings.Builder
	output.WriteString("╔═════════════════════ GENERAR REPORTE ════════════════════════╗\n")

	if pathFileLs == "" {
		pathFileLs = "/"
	}

	// Verificar que hay una sesión activa
	if !User_Groups.IsUserLoggedIn() {
		output.WriteString("Error: No hay una sesión activa. Use 'login' primero.\n")
		output.WriteString("╚═══════════════════════ FIN GENERAR REPORTE ═══════════════════╝\n")
		return output.String()
	}

	// Obtener la partición montada por ID
	mountedPartition, found := Environment.GetMountedPartitionByID(id)
	if !found {
		output.WriteString(fmt.Sprintf("Error: No se encontró la partición con ID %s montada\n", id))
		output.WriteString("╚═══════════════════════ FIN GENERAR REPORTE ═══════════════════╝\n")
		return output.String()
	}

	// Abrir el archivo
	file, err := Utils.AbrirArchivo(mountedPartition.MountPath)
	if err != nil {
		output.WriteString(fmt.Sprintf("Error al abrir el archivo: %v\n", err))
		output.WriteString("╚═══════════════════════ FIN GENERAR REPORTE ═══════════════════╝\n")
		return output.String()
	}
	defer file.Close()

	// Leer superbloque
	var superblock Partitions.Superblock
	if err := Utils.LeerArchivo(file, &superblock, int64(mountedPartition.MountStart)); err != nil {
		output.WriteString(fmt.Sprintf("Error al leer el superbloque: %v\n", err))
		output.WriteString("╚═══════════════════════ FIN GENERAR REPORTE ═══════════════════╝\n")
		return output.String()
	}

	// Buscar directorio (el inodo raíz siempre es 0 para "/")
	var inodeIndex int32
	var searchLog string
	if pathFileLs == "/" {
		inodeIndex = 0 // El inodo raíz siempre es 0 en EXT2
		output.WriteString("DEBUG: Usando inodo raíz (0) para la ruta '/'\n")
	} else {
		inodeIndex, searchLog = User_Groups.InitSearch(pathFileLs, file, superblock)
		output.WriteString(fmt.Sprintf("DEBUG: Resultado de búsqueda para '%s': inodo=%d\n", pathFileLs, inodeIndex))
		output.WriteString("DEBUG: Log de búsqueda:\n" + searchLog + "\n")

		if inodeIndex == -1 {
			output.WriteString(fmt.Sprintf("Error al buscar directorio: ruta '%s' no encontrada\n", pathFileLs))
			output.WriteString("╚═══════════════════════ FIN GENERAR REPORTE ═══════════════════╝\n")
			return output.String()
		}
	}

	// Leer inodo del directorio
	var dirInode Partitions.Inode
	inodePos := superblock.S_inode_start + inodeIndex*int32(binary.Size(Partitions.Inode{}))
	if err := Utils.LeerArchivo(file, &dirInode, int64(inodePos)); err != nil {
		output.WriteString(fmt.Sprintf("Error al leer inodo en posición %d: %v\n", inodePos, err))
		output.WriteString("╚═══════════════════════ FIN GENERAR REPORTE ═══════════════════╝\n")
		return output.String()
	}

	// Mostrar información detallada del inodo para depuración
	output.WriteString(fmt.Sprintf("DEBUG: Información del inodo %d:\n", inodeIndex))
	output.WriteString(fmt.Sprintf("  - Tipo: '%c' (0=directorio, 1=archivo)\n", dirInode.I_type[0]))
	output.WriteString(fmt.Sprintf("  - UID: %d\n", dirInode.I_uid))
	output.WriteString(fmt.Sprintf("  - GID: %d\n", dirInode.I_gid))
	output.WriteString(fmt.Sprintf("  - Tamaño: %d bytes\n", dirInode.I_size))
	output.WriteString(fmt.Sprintf("  - Fecha creación: %s\n", string(dirInode.I_ctime[:])))
	output.WriteString(fmt.Sprintf("  - Permisos: %s\n", string(dirInode.I_perm[:])))

	// Mostrar los primeros bloques asignados
	output.WriteString("  - Bloques asignados: [")
	for i := 0; i < 5 && i < len(dirInode.I_block); i++ {
		if i > 0 {
			output.WriteString(", ")
		}
		output.WriteString(fmt.Sprintf("%d", dirInode.I_block[i]))
	}
	output.WriteString("...]\n")

	// Verificar si es un directorio (debe tener I_type[0] = '0')
	if dirInode.I_type[0] != '0' {
		output.WriteString(fmt.Sprintf("Error al generar contenido LS: la ruta '%s' no es un directorio (tipo=%c)\n",
			pathFileLs, dirInode.I_type[0]))
		output.WriteString("╚═══════════════════════ FIN GENERAR REPORTE ═══════════════════╝\n")
		return output.String()
	}

	// Si llegamos aquí, es un directorio válido, continuar con la generación del reporte...
	// [resto del código para generar el reporte]

	output.WriteString("╚═══════════════════════ FIN GENERAR REPORTE ═══════════════════╝\n")
	return output.String()
}

func generateLSDotContent(partition Environment.MountedPartition, dirPath string) (string, error) {
	file, err := Utils.AbrirArchivo(partition.MountPath)
	if err != nil {
		return "", fmt.Errorf("no se pudo abrir el disco: %v", err)
	}
	defer file.Close()

	// Leer superbloque
	var superblock Partitions.Superblock
	if err := Utils.LeerArchivo(file, &superblock, int64(partition.MountStart)); err != nil {
		return "", fmt.Errorf("no se pudo leer el superbloque: %v", err)
	}

	// Buscar directorio (el inodo raíz siempre es 0 para "/")
	var inodeIndex int32
	var searchLog string
	if dirPath == "/" {
		inodeIndex = 0 // El inodo raíz siempre es 0 en EXT2
	} else {
		inodeIndex, searchLog = User_Groups.InitSearch(dirPath, file, superblock)
		if inodeIndex == -1 {
			return "", fmt.Errorf("error al buscar directorio: %s", searchLog)
		}
	}

	// Leer inodo del directorio
	var dirInode Partitions.Inode
	inodePos := superblock.S_inode_start + inodeIndex*int32(binary.Size(Partitions.Inode{}))
	if err := Utils.LeerArchivo(file, &dirInode, int64(inodePos)); err != nil {
		return "", fmt.Errorf("error al leer inodo en posición %d: %v", inodePos, err)
	}

	// Verificar si es un directorio, considerando casos especiales
	if dirPath == "/" {
		// Para la raíz, asumimos que es directorio pero verificamos
		if dirInode.I_type[0] != '0' {
			// Si no está marcado como directorio, lo marcamos pero advertimos
			fmt.Println("Advertencia: El inodo raíz no estaba marcado como directorio")
		}
	} else if dirInode.I_type[0] != '0' {
		// Para directorios no raíz, verificamos estrictamente
		return "", fmt.Errorf("la ruta '%s' no es un directorio", dirPath)
	}

	// Imprimir información de depuración
	fmt.Printf("DEBUG: Directorio '%s', Inodo %d, Tipo: %c\n",
		dirPath, inodeIndex, dirInode.I_type[0])

	// Resto del código...

	// Generar gráfico DOT
	var dot strings.Builder
	dot.WriteString("digraph G {\n")
	dot.WriteString("  rankdir=\"LR\";\n")
	dot.WriteString("  node [shape=plaintext];\n")
	dot.WriteString("  graph [fontname=\"Arial\", fontsize=10];\n")
	dot.WriteString("  edge [fontname=\"Arial\", fontsize=8];\n\n")

	// Crear tabla principal
	dot.WriteString("  ls_table [label=<\n")
	dot.WriteString("    <table border='0' cellborder='1' cellspacing='0'>\n")
	dot.WriteString("      <tr><td colspan='6' bgcolor='#e0e0e0'><b>Contenido de: " + dirPath + "</b></td></tr>\n")
	dot.WriteString("      <tr><td><b>Permisos</b></td><td><b>Owner</b></td><td><b>Grupo</b></td><td><b>Tamaño</b></td><td><b>Fecha</b></td><td><b>Nombre</b></td></tr>\n")

	// Procesar entradas del directorio
	for i, blockIndex := range dirInode.I_block {
		if blockIndex == -1 || i >= 12 { // Solo bloques directos
			continue
		}

		var folderBlock Partitions.Folderblock
		blockPos := superblock.S_block_start + blockIndex*int32(binary.Size(Partitions.Folderblock{}))
		if err := Utils.LeerArchivo(file, &folderBlock, int64(blockPos)); err != nil {
			continue
		}

		for _, entry := range folderBlock.B_content {
			if entry.B_inodo == -1 {
				continue
			}

			name := strings.TrimRight(string(entry.B_name[:]), "\x00")
			if name == "." || name == ".." {
				continue
			}

			var entryInode Partitions.Inode
			entryInodePos := superblock.S_inode_start + entry.B_inodo*int32(binary.Size(Partitions.Inode{}))
			if err := Utils.LeerArchivo(file, &entryInode, int64(entryInodePos)); err != nil {
				continue
			}

			// Formatear datos
			entryType := "📄"
			if entryInode.I_type[0] == '0' {
				entryType = "📁"
			}

			// Manejo seguro de la fecha
			modTime := "N/A"
			if len(entryInode.I_mtime) >= 4 { // Asegurarse de que haya al menos 4 bytes
				mtimeInt := int64(binary.LittleEndian.Uint32(entryInode.I_mtime[:4]))
				if mtimeInt > 0 {
					modTime = time.Unix(mtimeInt, 0).Format("02/01/2006 ")
				}
			}

			// Añadir fila a la tabla
			dot.WriteString(fmt.Sprintf("      <tr><td>%s</td><td>%d</td><td>%d</td><td>%d</td><td>%s</td><td>%s %s</td></tr>\n",
				string(entryInode.I_perm[:]),
				entryInode.I_uid,
				entryInode.I_gid,
				entryInode.I_size,
				modTime,
				entryType,
				name))
		}
	}

	dot.WriteString("    </table>\n")
	dot.WriteString("  >];\n")
	dot.WriteString("}\n")

	return dot.String(), nil
}
