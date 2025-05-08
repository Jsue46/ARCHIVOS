package Reportes

import (
	"Proyecto/Environment"
	"Proyecto/Partitions"
	"Proyecto/Utils"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GenerarReporteArbol genera un reporte visual del árbol de inodos y bloques
func GenerarReporteArbol(diskPath, outputPath, id string) string {
	fmt.Printf("🔍 Depuración: Iniciando generación de reporte de árbol para ID: %s\n", id)

	// Abrir el archivo binario del disco montado
	file, err := Utils.AbrirArchivo(diskPath)
	if err != nil {
		fmt.Printf("❌ Error al abrir el archivo: %v\n", err)
		return fmt.Sprintf("Error: No se pudo abrir el archivo en la ruta: %s", diskPath)
	}
	defer file.Close()
	fmt.Println("✅ Archivo abierto correctamente.")

	// Obtener la partición montada
	var partitionStart int64
	partitionFound := false
	for _, partitions := range Environment.GetMountedPartitions() {
		for _, partition := range partitions {
			if partition.MountID == id {
				partitionStart = int64(partition.MountStart)
				partitionFound = true
				fmt.Printf("✅ Partición encontrada: MountStart = %d\n", partitionStart)
				break
			}
		}
		if partitionFound {
			break
		}
	}

	if !partitionFound {
		fmt.Printf("❌ Error: No se encontró la partición con ID: %s\n", id)
		return fmt.Sprintf("Error: No se encontró la partición con ID: %s", id)
	}

	// Leer el superbloque
	var superblock Partitions.Superblock
	if err := Utils.LeerArchivo(file, &superblock, partitionStart); err != nil {
		fmt.Printf("❌ Error al leer el superbloque: %v\n", err)
		return fmt.Sprintf("Error al leer superbloque: %v", err)
	}
	fmt.Printf("✅ Superbloque leído correctamente: S_filesystem_type = %d\n", superblock.S_filesystem_type)

	// Validar que es un sistema de archivos válido
	if superblock.S_filesystem_type == 0 {
		fmt.Println("❌ Error: El sistema de archivos no está formateado.")
		return "Error: El sistema de archivos no está formateado"
	}

	// Crear el archivo DOT para el árbol
	dotContent := &strings.Builder{}
	dotContent.WriteString("digraph G {\n")
	dotContent.WriteString("  rankdir=\"LR\";\n")
	dotContent.WriteString("  node [shape=record, fontname=\"Arial\", fontsize=10];\n\n")

	// Procesar el inodo raíz (generalmente inodo 0)
	fmt.Println("🔍 Depuración: Procesando inodo raíz (Inodo 0).")
	if err := processInodeForTree(0, file, superblock, dotContent); err != nil {
		fmt.Printf("❌ Error al procesar inodo raíz: %v\n", err)
		return fmt.Sprintf("Error al procesar inodo raíz: %v", err)
	}

	dotContent.WriteString("}\n")

	// Depuración: Imprimir el contenido del archivo DOT
	fmt.Println("🔍 Depuración: Contenido del archivo DOT generado:")
	fmt.Println(dotContent.String())

	// Crear directorio si no existe
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		fmt.Printf("❌ Error al crear directorio: %v\n", err)
		return fmt.Sprintf("Error al crear directorio: %v", err)
	}
	fmt.Println("✅ Directorio creado correctamente.")

	// Guardar contenido DOT
	dotFile := strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + ".dot"
	if err := os.WriteFile(dotFile, []byte(dotContent.String()), 0644); err != nil {
		fmt.Printf("❌ Error al guardar el archivo DOT: %v\n", err)
		return fmt.Sprintf("Error al guardar el archivo DOT: %v", err)
	}
	fmt.Printf("✅ Archivo DOT guardado correctamente: %s\n", dotFile)

	// Convertir el archivo DOT en imagen con Graphviz
	cmd := exec.Command("dot", "-Tjpg", dotFile, "-o", outputPath)
	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Error al convertir el archivo DOT a imagen: %v\n", err)
		return fmt.Sprintf("Error al convertir el archivo DOT a imagen: %v", err)
	}
	fmt.Printf("✅ Imagen generada correctamente: %s\n", outputPath)

	return fmt.Sprintf("Reporte de árbol generado exitosamente en: %s", outputPath)
}

func processInodeForTree(inodeIndex int32, file *os.File, sb Partitions.Superblock, dot *strings.Builder) error {
	fmt.Printf("🔍 Depuración: Procesando inodo %d\n", inodeIndex)

	// Validar que el inodo esté dentro del rango válido
	if inodeIndex < 0 || inodeIndex >= sb.S_inodes_count {
		fmt.Printf("❌ Error: Índice de inodo %d fuera de rango\n", inodeIndex)
		return fmt.Errorf("índice de inodo %d fuera de rango", inodeIndex)
	}

	// Calcular posición del inodo
	inodePos := sb.S_inode_start + inodeIndex*int32(binary.Size(Partitions.Inode{}))
	fmt.Printf("🔍 Depuración: Posición del inodo %d = %d\n", inodeIndex, inodePos)

	// Leer el inodo
	var inode Partitions.Inode
	if err := Utils.LeerArchivo(file, &inode, int64(inodePos)); err != nil {
		fmt.Printf("❌ Error al leer inodo %d: %v\n", inodeIndex, err)
		return fmt.Errorf("error al leer inodo %d: %v", inodeIndex, err)
	}
	fmt.Printf("✅ Inodo %d leído correctamente: Tipo = %s, Tamaño = %d\n",
		inodeIndex, string(inode.I_type[:]), inode.I_size)

	// Procesar bloques del inodo
	for _, blockIndex := range inode.I_block {
		if blockIndex == -1 {
			continue
		}
		fmt.Printf("🔍 Depuración: Procesando bloque %d del inodo %d\n", blockIndex, inodeIndex)
	}

	return nil
}

func processDirectoryBlock(blockIndex int32, file *os.File, sb Partitions.Superblock, dot *strings.Builder) error {
	// Validar que el bloque esté dentro del rango válido
	if blockIndex < 0 || blockIndex >= sb.S_blocks_count {
		return fmt.Errorf("índice de bloque %d fuera de rango", blockIndex)
	}

	// Leer el bloque de directorio
	var folderBlock Partitions.Folderblock
	blockPos := sb.S_block_start + blockIndex*int32(binary.Size(Partitions.Folderblock{}))
	if err := Utils.LeerArchivo(file, &folderBlock, int64(blockPos)); err != nil {
		return fmt.Errorf("error al leer bloque directorio %d: %v", blockIndex, err)
	}

	// Crear nodo para el bloque
	dot.WriteString(fmt.Sprintf("  block%d [label=\"Bloque Directorio %d|{", blockIndex, blockIndex))

	for i, content := range folderBlock.B_content {
		if content.B_inodo != -1 {
			name := strings.TrimRight(string(content.B_name[:]), "\x00")
			dot.WriteString(fmt.Sprintf("<f%d> %s (Inodo %d)|", i, name, content.B_inodo))
		}
	}
	dot.WriteString("}\"];\n\n")

	// Procesar los inodos referenciados (excepto . y ..)
	for i, content := range folderBlock.B_content {
		if content.B_inodo != -1 && !(string(content.B_name[:]) == "." || string(content.B_name[:]) == "..") {
			dot.WriteString(fmt.Sprintf("  block%d:f%d -> inode%d;\n", blockIndex, i, content.B_inodo))
			if err := processInodeForTree(content.B_inodo, file, sb, dot); err != nil {
				return err
			}
		}
	}

	return nil
}

func processFileBlock(blockIndex int32, file *os.File, sb Partitions.Superblock, dot *strings.Builder) error {
	// Validar que el bloque esté dentro del rango válido
	if blockIndex < 0 || blockIndex >= sb.S_blocks_count {
		return fmt.Errorf("índice de bloque %d fuera de rango", blockIndex)
	}

	// Leer el bloque de archivo
	var fileBlock Partitions.Fileblock
	blockPos := sb.S_block_start + blockIndex*int32(binary.Size(Partitions.Fileblock{}))
	if err := Utils.LeerArchivo(file, &fileBlock, int64(blockPos)); err != nil {
		return fmt.Errorf("error al leer bloque archivo %d: %v", blockIndex, err)
	}

	// Limpiar y filtrar el contenido del bloque
	content := cleanContent(string(fileBlock.B_content[:]))

	// Crear nodo para el bloque
	dot.WriteString(fmt.Sprintf("  block%d [label=\"Bloque Archivo %d|{%s}\"];\n",
		blockIndex, blockIndex, content))

	return nil
}

func cleanContent(input string) string {
	var cleaned strings.Builder
	for _, r := range input {
		// Filtrar caracteres imprimibles (ASCII 32-126) y eliminar caracteres no deseados
		if r >= 32 && r <= 126 {
			cleaned.WriteRune(r)
		}
	}
	// Eliminar espacios en blanco redundantes y caracteres no deseados
	return strings.TrimSpace(cleaned.String())
}
