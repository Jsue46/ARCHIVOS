package Partitions

import (
	"fmt"
	"strings"
)

// Definición de la estructura del MBR
type MBR struct {
	MbrTamanio       int32
	MbrFechaCreacion [19]byte
	MbrDskSignature  int32
	DskFit           [1]byte
	MbrPartitions    [4]Partition
}

// Función para imprimir un MBR
func ImprimirMBR(mbr MBR) string {
	var output strings.Builder

	output.WriteString(fmt.Sprintf("  Signature: %d", mbr.MbrDskSignature))
	output.WriteString(fmt.Sprintf("  Fecha Creación: %s", string(mbr.MbrFechaCreacion[:])))
	output.WriteString(fmt.Sprintf("  Tamaño: %d bytes", mbr.MbrTamanio))
	output.WriteString(fmt.Sprintf("  Fit: %s \n", string(mbr.DskFit[:])))

	return output.String()
}

// Definición de la estructura de una partición
type Partition struct {
	PartStatus      [1]byte
	PartType        [1]byte
	PartFit         [1]byte
	PartStart       int32
	PartSize        int32
	PartName        [16]byte
	PartCorrelative int32
	PartID          [4]byte
}

// Función para imprimir una partición
func ImprimirPartition(data Partition) string {
	var output strings.Builder

	output.WriteString(fmt.Sprintf("  Name: %s", string(data.PartName[:])))
	output.WriteString(fmt.Sprintf("  Type: %s", string(data.PartType[:])))
	output.WriteString(fmt.Sprintf("  Start: %d", data.PartStart))
	output.WriteString(fmt.Sprintf("  Size: %d", data.PartSize))
	output.WriteString(fmt.Sprintf("  Status: %s", string(data.PartStatus[:])))
	output.WriteString(fmt.Sprintf("  Id: %s\n", string(data.PartID[:])))

	return output.String()
}

// Definición de la estructura de un EBR
type EBR struct {
	PartMount byte
	PartFit   byte
	PartStart int32
	PartSize  int32
	PartNext  int32
	PartName  [16]byte
}

// Función para imprimir un EBR
func ImprimirEBR(ebr EBR) string {
	var output strings.Builder

	output.WriteString(fmt.Sprintf("  Name: %s", string(ebr.PartName[:])))
	output.WriteString(fmt.Sprintf("  Fit: %c", ebr.PartFit))
	output.WriteString(fmt.Sprintf("  Start: %d", ebr.PartStart))
	output.WriteString(fmt.Sprintf("  Size: %d", ebr.PartSize))
	output.WriteString(fmt.Sprintf("  Next: %d", ebr.PartNext))
	output.WriteString(fmt.Sprintf("  Mount: %c\n", ebr.PartMount))

	return output.String()
}

// Definición de la estructura del superbloque
type Superblock struct {
	S_filesystem_type   int32
	S_inodes_count      int32
	S_blocks_count      int32
	S_free_blocks_count int32
	S_free_inodes_count int32
	S_mtime             [17]byte
	S_umtime            [17]byte
	S_mnt_count         int32
	S_magic             int32
	S_inode_size        int32
	S_block_size        int32
	S_fist_ino          int32
	S_first_blo         int32
	S_bm_inode_start    int32
	S_bm_block_start    int32
	S_inode_start       int32
	S_block_start       int32
}

// Función para imprimir un superbloque
func PrintSuperblock(sb Superblock) string {
	var output strings.Builder
	output.WriteString(" ═════════════════════  SUPERBLOCK   ══════════════════════════ \n")
	output.WriteString(fmt.Sprintf(" S_filesystem_type: %d\n", sb.S_filesystem_type))
	output.WriteString(fmt.Sprintf(" S_inodes_count: %d\n", sb.S_inodes_count))
	output.WriteString(fmt.Sprintf(" S_blocks_count: %d\n", sb.S_blocks_count))
	output.WriteString(fmt.Sprintf(" S_free_blocks_count: %d\n", sb.S_free_blocks_count))
	output.WriteString(fmt.Sprintf(" S_free_inodes_count: %d\n", sb.S_free_inodes_count))
	output.WriteString(fmt.Sprintf(" S_mtime: %s\n", string(sb.S_mtime[:])))
	output.WriteString(fmt.Sprintf(" S_umtime: %s\n", string(sb.S_umtime[:])))
	output.WriteString(fmt.Sprintf(" S_mnt_count: %d\n", sb.S_mnt_count))
	output.WriteString(fmt.Sprintf(" S_magic: 0x%X\n", sb.S_magic))
	output.WriteString(fmt.Sprintf(" S_inode_size: %d\n", sb.S_inode_size))
	output.WriteString(fmt.Sprintf(" S_block_size: %d\n", sb.S_block_size))
	output.WriteString(fmt.Sprintf(" S_fist_ino: %d\n", sb.S_fist_ino))
	output.WriteString(fmt.Sprintf(" S_first_blo: %d\n", sb.S_first_blo))
	output.WriteString(fmt.Sprintf(" S_bm_inode_start: %d\n", sb.S_bm_inode_start))
	output.WriteString(fmt.Sprintf(" S_bm_block_start: %d\n", sb.S_bm_block_start))
	output.WriteString(fmt.Sprintf(" S_inode_start: %d\n", sb.S_inode_start))
	output.WriteString(fmt.Sprintf(" S_block_start: %d\n", sb.S_block_start))
	output.WriteString(" ══════════════════════════════════════════════════════════════ \n")
	return output.String()
}

// Definición de la estructura de un inodo
type Inode struct {
	I_uid   int32
	I_gid   int32
	I_size  int32
	I_atime [17]byte
	I_ctime [17]byte
	I_mtime [17]byte
	I_block [15]int32
	I_type  [1]byte
	I_perm  [3]byte
}

// Función para imprimir un inodo
func PrintInode(inode Inode) string {
	var output strings.Builder
	output.WriteString(" ════════════════════════  INODO  ═════════════════════════════ \n")
	output.WriteString(fmt.Sprintf(" I_uid: %d\n", inode.I_uid))
	output.WriteString(fmt.Sprintf(" I_gid: %d\n", inode.I_gid))
	output.WriteString(fmt.Sprintf(" I_size: %d\n", inode.I_size))
	output.WriteString(fmt.Sprintf(" I_atime: %s\n", string(inode.I_atime[:])))
	output.WriteString(fmt.Sprintf(" I_ctime: %s\n", string(inode.I_ctime[:])))
	output.WriteString(fmt.Sprintf(" I_mtime: %s\n", string(inode.I_mtime[:])))
	output.WriteString(fmt.Sprintf(" I_type: %s\n", string(inode.I_type[:])))
	output.WriteString(fmt.Sprintf(" I_perm: %s\n", string(inode.I_perm[:])))
	output.WriteString(fmt.Sprintf(" I_block: %v\n", inode.I_block))
	output.WriteString(" ══════════════════════════════════════════════════════════════ \n")
	return output.String()
}

// Definición de la estructura de un bloque de carpetas
type Folderblock struct {
	B_content [4]Content
}

// Función para imprimir un bloque de carpetas
func PrintFolderblock(folderblock Folderblock) string {
	var output strings.Builder
	output.WriteString(" ═════════════════════  FOLDERBLOCK ═══════════════════════════ \n")
	for i, content := range folderblock.B_content {
		output.WriteString(fmt.Sprintf(" Contenido %d: Name: %s, Inodo: %d\n", i, string(content.B_name[:]), content.B_inodo))
	}
	output.WriteString(" ══════════════════════════════════════════════════════════════ \n")
	return output.String()
}

// Definición de la estructura de un bloque de archivos
type Content struct {
	B_name  [12]byte
	B_inodo int32
}

// Definición de la estructura de un bloque de archivos
type Fileblock struct {
	B_content [64]byte
}

// Función para imprimir un bloque de archivos
func PrintFileblock(fileblock Fileblock) string {
	var output strings.Builder
	output.WriteString(" ══════════════════════  FILEBLOCK ════════════════════════════ \n")
	output.WriteString(fmt.Sprintf(" B_content: %s\n", string(fileblock.B_content[:])))
	output.WriteString(" ══════════════════════════════════════════════════════════════ \n")
	return output.String()
}

// Definición de la estructura de un bloque de apuntadores
type Pointerblock struct {
	B_pointers [16]int32
}

// Función para imprimir un bloque de apuntadores
func PrintPointerblock(pointerblock Pointerblock) string {
	var output strings.Builder
	output.WriteString(" ═════════════════════  POINTERBLOCK ══════════════════════════ \n")
	for i, pointer := range pointerblock.B_pointers {
		output.WriteString(fmt.Sprintf(" Pointer %d: %d\n", i, pointer))
	}
	output.WriteString(" ══════════════════════════════════════════════════════════════ \n")
	return output.String()
}
