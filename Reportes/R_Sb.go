package Reportes

import (
	"Proyecto/Environment"
	"Proyecto/Partitions"
	"Proyecto/Utils"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func GenerarReporteSB(diskPath, path string, id string) string {
	// Abrir el archivo binario del disco montado
	file, err := Utils.AbrirArchivo(diskPath)
	if err != nil {
		return fmt.Sprintf("Error: No se pudo abrir el archivo en la ruta: %s", diskPath)
	}
	defer file.Close()

	// Leer el superbloque de la partición montada
	var tempSuperblock Partitions.Superblock
	superblockStart := 0 // Aquí debes calcular la posición inicial
	if err := Utils.LeerArchivo(file, &tempSuperblock, int64(superblockStart)); err != nil {
		return "Error: No se pudo leer el superbloque desde el archivo"
	}

	// Generar el archivo .dot del Superblock
	reportPath := path
	if err := GenerateSBReport(id, reportPath); err != nil {
		return fmt.Sprintf("Error al generar el reporte SB: %v", err)
	}

	// Renderizar el archivo .dot a .jpg usando Graphviz
	dotFile := strings.TrimSuffix(reportPath, filepath.Ext(reportPath)) + ".dot"
	outputJpg := reportPath
	cmd := exec.Command("dot", "-Tjpg", dotFile, "-o", outputJpg)
	err = cmd.Run()
	if err != nil {
		return fmt.Sprintf("Error al renderizar el archivo .dot a imagen: %v", err)
	}

	return fmt.Sprintf("Imagen generada exitosamente en: %s", outputJpg)
}

func GenerateSBReport(id string, outputPath string) error {
	// Buscar la partición montada por ID
	var mountedPartition Environment.MountedPartition
	var partitionFound bool

	for _, partitions := range Environment.GetMountedPartitions() {
		for _, partition := range partitions {
			if partition.MountID == id {
				mountedPartition = partition
				partitionFound = true
				break
			}
		}
		if partitionFound {
			break
		}
	}

	if !partitionFound {
		return fmt.Errorf("partición no encontrada")
	}

	// Abrir el archivo binario de la partición
	file, err := Utils.AbrirArchivo(mountedPartition.MountPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Leer el MBR
	var TempMBR Partitions.MBR
	if err := Utils.LeerArchivo(file, &TempMBR, 0); err != nil {
		return err
	}

	// Buscar la partición dentro del MBR
	var index int = -1
	for i := 0; i < 4; i++ {
		if TempMBR.MbrPartitions[i].PartSize != 0 {
			if strings.Contains(string(TempMBR.MbrPartitions[i].PartID[:]), id) {
				index = i
				break
			}
		}
	}

	if index == -1 {
		return fmt.Errorf("partición no encontrada ")
	}

	// Leer el Superblock
	var sb Partitions.Superblock
	sbStart := TempMBR.MbrPartitions[index].PartStart
	if err := Utils.LeerArchivo(file, &sb, int64(sbStart)); err != nil {
		return err
	}

	// Generar el reporte en formato DOT
	return GenerateDOTFile(sb, outputPath)
}

func GenerateDOTFile(sb Partitions.Superblock, outputPath string) error {
	currentDate := time.Now().Format("02/01/2006")
	currentDate2 := time.Now().Format("02/01/2006")
	copy(sb.S_mtime[:], currentDate)
	copy(sb.S_umtime[:], currentDate2)

	reportsDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(reportsDir, os.ModePerm); err != nil {
		return fmt.Errorf("error al crear la carpeta de reportes: %v", err)
	}

	dotFilePath := strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + ".dot"
	dotFile, err := os.Create(dotFilePath)
	if err != nil {
		return fmt.Errorf("error al crear el archivo .dot de reporte: %v", err)
	}
	defer dotFile.Close()

	graphContent := fmt.Sprintf(`
digraph G {
    node [fillcolor=lightyellow style=filled]
    rankdir=LR;
    subgraph cluster_SB {
        color=lightblue fillcolor=aliceblue label="Superblock Report" style=filled
        sb [label=<<table border="0" cellborder="1" cellspacing="0" cellpadding="4">
            <tr><td colspan="2" bgcolor="lightblue"><b>Superblock Information</b></td></tr>
            <tr><td><b>S_filesystem_type</b></td><td>%d (EXT2)</td></tr>
            <tr><td><b>S_inodes_count</b></td><td>%d</td></tr>
            <tr><td><b>S_blocks_count</b></td><td>%d</td></tr>
            <tr><td><b>S_free_blocks_count</b></td><td>%d</td></tr>
            <tr><td><b>S_free_inodes_count</b></td><td>%d</td></tr>
            <tr><td><b>S_mtime</b></td><td>%s</td></tr>
            <tr><td><b>S_umtime</b></td><td>%s</td></tr>
            <tr><td><b>S_mnt_count</b></td><td>%d</td></tr>
            <tr><td><b>S_magic</b></td><td>0x%X</td></tr>
            <tr><td><b>S_inode_size</b></td><td>%d bytes</td></tr>
            <tr><td><b>S_block_size</b></td><td>%d bytes</td></tr>
            <tr><td><b>S_bm_inode_start</b></td><td>%d</td></tr>
            <tr><td><b>S_bm_block_start</b></td><td>%d</td></tr>
            <tr><td><b>S_inode_start</b></td><td>%d</td></tr>
            <tr><td><b>S_block_start</b></td><td>%d</td></tr>
            <tr><td><b>S_fist_ino</b></td><td>%d</td></tr> <!-- Nuevo campo -->
            <tr><td><b>S_first_blo</b></td><td>%d</td></tr> <!-- Nuevo campo -->
        </table>> shape=plaintext]
    }
}
`, sb.S_filesystem_type, sb.S_inodes_count, sb.S_blocks_count, sb.S_free_blocks_count,
		sb.S_free_inodes_count, currentDate, currentDate2, sb.S_mnt_count,
		sb.S_magic, sb.S_inode_size, sb.S_block_size, sb.S_bm_inode_start,
		sb.S_bm_block_start, sb.S_inode_start, sb.S_block_start, sb.S_fist_ino, sb.S_first_blo)

	_, err = dotFile.WriteString(graphContent)
	if err != nil {
		return fmt.Errorf("error al escribir en el archivo .dot: %v", err)
	}

	return nil
}
