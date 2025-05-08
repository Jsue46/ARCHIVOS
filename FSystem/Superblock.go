package FSystem

import (
	"Proyecto/Partitions"
	"Proyecto/Utils"
	"fmt"
	"os"
	"strings"
)

// Función auxiliar para marcar los inodos y bloques usados
func markUsedInodesAndBlocks(newSuperblock Partitions.Superblock, file *os.File) error {
	if err := Utils.EscribirArchivo(file, byte(1), int64(newSuperblock.S_bm_inode_start)); err != nil {
		return err
	}
	if err := Utils.EscribirArchivo(file, byte(1), int64(newSuperblock.S_bm_inode_start+1)); err != nil {
		return err
	}
	if err := Utils.EscribirArchivo(file, byte(1), int64(newSuperblock.S_bm_block_start)); err != nil {
		return err
	}
	if err := Utils.EscribirArchivo(file, byte(1), int64(newSuperblock.S_bm_block_start+1)); err != nil {
		return err
	}
	return nil
}

// Función para imprimir el Superblock

func PrintSuperblock(sb Partitions.Superblock) string {
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
