package FSystem

import (
	"Proyecto/Partitions"
	"Proyecto/Utils"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

// Función auxiliar para crear el sistema de archivos EXT2
func create_ext2(n int32, partition Partitions.Partition, newSuperblock Partitions.Superblock, date string, file *os.File) string {
	var output strings.Builder
	output.WriteString(" ═════════════════════  CREANDO EXT2 ══════════════════════════ \n")
	output.WriteString(fmt.Sprintf("  INODOS: %d \n", n))

	// Imprimir el Superblock calculado
	output.WriteString(Partitions.PrintSuperblock(newSuperblock))
	output.WriteString(fmt.Sprintf("  Date: %s\n", date))

	// Escribir los bitmaps de inodos y bloques
	for i := int32(0); i < n; i++ {
		if err := Utils.EscribirArchivo(file, byte(0), int64(newSuperblock.S_bm_inode_start+i)); err != nil {
			return "Error al escribir en el bitmap de inodos"
		}
	}

	// Escribir los bitmaps de bloques
	for i := int32(0); i < 3*n; i++ {
		if err := Utils.EscribirArchivo(file, byte(0), int64(newSuperblock.S_bm_block_start+i)); err != nil {
			return "Error al escribir en el bitmap de bloques"
		}
	}

	// Inicializar los inodos y bloques con valores predeterminados
	if err := initInodesAndBlocks(n, newSuperblock, file); err != nil {
		return "Error al inicializar inodos y bloques"
	}

	// Crear la carpeta raíz y el archivo "users.txt"
	if err := createRootAndUsersFile(newSuperblock, date, file); err != nil {
		return "Error al crear la carpeta raíz y el archivo users.txt"
	}

	// Escribir el superbloque actualizado en el archivo
	if err := Utils.EscribirArchivo(file, newSuperblock, int64(partition.PartStart)); err != nil {
		return "Error al escribir el superbloque actualizado en el archivo"
	}

	// Marcar los primeros inodos y bloques como usados
	if err := markUsedInodesAndBlocks(newSuperblock, file); err != nil {
		return "Error al marcar inodos y bloques como usados"
	}

	// Leer e imprimir los inodos después de formatear
	output.WriteString(" ══════════════════  IMPRIMIENDO INODOS  ══════════════════════ \n")
	for i := int32(0); i < n; i++ {
		var inode Partitions.Inode
		offset := int64(newSuperblock.S_inode_start + i*int32(binary.Size(Partitions.Inode{})))
		if err := Utils.LeerArchivo(file, &inode, offset); err != nil {
			return fmt.Sprintf("Error al leer inodo: %v", err)
		}
		Partitions.PrintInode(inode)
	}

	// Leer e imprimir los Folderblocks y Fileblocks
	output.WriteString(" ══════════════  FOLDERBLOCKS Y FILEBLOCKS  ═══════════════════ \n")

	// Imprimir Folderblocks
	for i := int32(0); i < 1; i++ {
		var folderblock Partitions.Folderblock
		offset := int64(newSuperblock.S_block_start + i*int32(binary.Size(Partitions.Folderblock{})))
		if err := Utils.LeerArchivo(file, &folderblock, offset); err != nil {
			return fmt.Sprintf("Error al leer Folderblock: %v", err)
		}
		output.WriteString(Partitions.PrintFolderblock(folderblock))
	}

	// Imprimir Fileblocks
	for i := int32(0); i < 1; i++ {
		var fileblock Partitions.Fileblock
		offset := int64(newSuperblock.S_block_start + int32(binary.Size(Partitions.Folderblock{})) + i*int32(binary.Size(Partitions.Fileblock{})))
		if err := Utils.LeerArchivo(file, &fileblock, offset); err != nil {
			return fmt.Sprintf("Error al leer Fileblock: %v", err)
		}
		output.WriteString(Partitions.PrintFileblock(fileblock))
	}

	// Imprimir el Superblock final
	output.WriteString(Partitions.PrintSuperblock(newSuperblock))
	output.WriteString(" ═══════════════════ FINALIZANDO EXT2  ════════════════════════ \n")
	return output.String()
}

// Función auxiliar para inicializar inodos y bloques
func initInodesAndBlocks(n int32, newSuperblock Partitions.Superblock, file *os.File) error {
	var newInode Partitions.Inode
	for i := int32(0); i < 15; i++ {
		newInode.I_block[i] = -1
	}

	for i := int32(0); i < n; i++ {
		if err := Utils.EscribirArchivo(file, newInode, int64(newSuperblock.S_inode_start+i*int32(binary.Size(Partitions.Inode{})))); err != nil {
			return err
		}
	}

	var newFileblock Partitions.Fileblock
	for i := int32(0); i < 3*n; i++ {
		if err := Utils.EscribirArchivo(file, newFileblock, int64(newSuperblock.S_block_start+i*int32(binary.Size(Partitions.Fileblock{})))); err != nil {
			return err
		}
	}

	return nil
}

// ----------------------------------------------------------------------------------------------- +Agregado
// Función auxiliar para crear la carpeta raíz y el archivo users.txt
func createRootAndUsersFile(newSuperblock Partitions.Superblock, date string, file *os.File) error {
	var Inode0, Inode1 Partitions.Inode

	// Inicializa los inodos con la fecha proporcionada
	initInode(&Inode0, date)
	initInode(&Inode1, date)

	// Asigna los bloques correspondientes a los inodos
	Inode0.I_block[0] = 0
	Inode1.I_block[0] = 1

	// Contenido del archivo users.txt
	data := "  1,G,root\n  1,U,root,root,123\n"
	actualSize := int32(len(data))
	Inode1.I_size = actualSize

	// Crea un bloque de archivo y copia el contenido en él
	var Fileblock1 Partitions.Fileblock
	copy(Fileblock1.B_content[:], data)

	// Crea un bloque de carpeta (raíz) y asigna las entradas iniciales
	var Folderblock0 Partitions.Folderblock
	Folderblock0.B_content[0].B_inodo = 0
	copy(Folderblock0.B_content[0].B_name[:], ".")
	Folderblock0.B_content[1].B_inodo = 0
	copy(Folderblock0.B_content[1].B_name[:], "..")
	Folderblock0.B_content[2].B_inodo = 1
	copy(Folderblock0.B_content[2].B_name[:], "users.txt")

	// Escribe los inodos y bloques en las posiciones correctas en el archivo del sistema
	if err := Utils.EscribirArchivo(file, Inode0, int64(newSuperblock.S_inode_start)); err != nil {
		return err
	}
	if err := Utils.EscribirArchivo(file, Inode1, int64(newSuperblock.S_inode_start+int32(binary.Size(Partitions.Inode{})))); err != nil {
		return err
	}
	if err := Utils.EscribirArchivo(file, Folderblock0, int64(newSuperblock.S_block_start)); err != nil {
		return err
	}
	if err := Utils.EscribirArchivo(file, Fileblock1, int64(newSuperblock.S_block_start+int32(binary.Size(Partitions.Folderblock{})))); err != nil {
		return err
	}

	return nil
}

// SearchInodeByPath busca un archivo o directorio en el sistema de archivos EXT2.
func SearchInodeByPath(steps []string, inode Partitions.Inode, file *os.File, superblock Partitions.Superblock) (int32, string) {
	var output strings.Builder
	output.WriteString(" ══════════════   BUSCANDO INODO POR PATH  ════════════════════ \n")

	index := int32(0)

	// Extrae el primer elemento del path y elimina espacios en blanco
	SearchedName := strings.Replace(steps[0], " ", "", -1)

	output.WriteString(fmt.Sprintf(" ══════════════  SearchedName: %s\n", SearchedName))

	// Iterar sobre los bloques del inodo
	for _, block := range inode.I_block {
		if block != -1 {
			if index < 13 {
				var crrFolderblock Partitions.Folderblock

				// Leer el bloque de carpeta desde el archivo binario
				if err := Utils.LeerArchivo(file, &crrFolderblock, int64(superblock.S_block_start+block*int32(binary.Size(Partitions.Folderblock{})))); err != nil {
					output.WriteString(fmt.Sprintf(" Error al leer el bloque de carpeta: %v\n", err))
					return -1, output.String()
				}

				// Buscar el archivo/directorio dentro del bloque de carpeta
				for _, folder := range crrFolderblock.B_content {
					output.WriteString(fmt.Sprintf(" ═══ Folder ═══ Name: %s, B_inodo: %d\n", string(folder.B_name[:]), folder.B_inodo))

					if strings.Contains(string(folder.B_name[:]), SearchedName) {
						output.WriteString(fmt.Sprintf(" len(steps): %d, steps: %v\n", len(steps), steps))

						if len(steps) == 1 {
							output.WriteString(" ═══ Folder found ═══ \n")
							return folder.B_inodo, output.String()
						} else {
							output.WriteString(" ═══ NextInode ═══ \n")
							var NextInode Partitions.Inode

							// Leer el siguiente inodo desde el archivo binario
							if err := Utils.LeerArchivo(file, &NextInode, int64(superblock.S_inode_start+folder.B_inodo*int32(binary.Size(Partitions.Inode{})))); err != nil {
								output.WriteString(fmt.Sprintf(" Error al leer el siguiente inodo: %v\n", err))
								return -1, output.String()
							}

							// Llamada recursiva para seguir con la búsqueda
							return SearchInodeByPath(steps[1:], NextInode, file, superblock)
						}
					}
				}
			} else {
				output.WriteString(" Manejo de bloques indirectos no implementado\n")
			}
		}
		index++
	}

	output.WriteString(" ════════════════   BUSQUEDA FINALIZADA   ═════════════════════ \n")
	return -1, output.String()
}
