package Reportes

import (
	"Proyecto/Environment"
	"Proyecto/Partitions"
	"Proyecto/Utils"
	"fmt"
	"os"
	"strings"
)

func GenerarReporteBitmapInodos(outputPath string, id string) string {
	// Verificar si se proporcionó una ruta específica para el reporte
	if outputPath == "" {
		return "Error: No se especificó la ruta del archivo de reporte."
	}

	// Obtener la partición montada con el ID correspondiente
	var filepath string
	var partitionFound bool
	for _, partitions := range Environment.GetMountedPartitions() {
		for _, partition := range partitions {
			if partition.MountID == id {
				filepath = partition.MountPath
				partitionFound = true
				break
			}
		}
		if partitionFound {
			break
		}
	}

	if !partitionFound {
		return fmt.Sprintf("Error: No se encontró la partición con ID: %s", id)
	}

	// Abrir el archivo binario del disco montado
	file, err := Utils.AbrirArchivo(filepath)
	if err != nil {
		return fmt.Sprintf("Error: No se pudo abrir el archivo en la ruta: %s", filepath)
	}
	defer file.Close()

	// Leer el MBR de la partición
	var TempMBR Partitions.MBR
	if err := Utils.LeerArchivo(file, &TempMBR, 0); err != nil {
		return "Error: No se pudo leer el MBR desde el archivo"
	}

	// Buscar la partición dentro del MBR
	var index int = -1
	for i := 0; i < 4; i++ {
		if TempMBR.MbrPartitions[i].PartSize != 0 {
			if strings.Contains(string(TempMBR.MbrPartitions[i].PartID[:]), id) {
				if TempMBR.MbrPartitions[i].PartStatus[0] == '1' {
					index = i
				} else {
					return "Error: La partición no está montada"
				}
				break
			}
		}
	}

	if index == -1 {
		return "Error: No se encontró la partición"
	}

	// Leer el superbloque de la partición montada
	var superblock Partitions.Superblock
	superblockStart := TempMBR.MbrPartitions[index].PartStart
	if err := Utils.LeerArchivo(file, &superblock, int64(superblockStart)); err != nil {
		return "Error: No se pudo leer el superbloque desde el archivo"
	}

	// Leer el bitmap de inodos
	inodeBitmapSize := superblock.S_inodes_count
	var bitmap []byte = make([]byte, inodeBitmapSize)
	if err := Utils.LeerArchivo(file, &bitmap, int64(superblock.S_bm_inode_start)); err != nil {
		return "Error: No se pudo leer el bitmap de inodos"
	}

	// Crear el archivo de reporte
	reportFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Sprintf("Error: No se pudo crear el archivo de reporte en: %s", outputPath)
	}
	defer reportFile.Close()

	// Escribir los bits en el archivo, mostrando 20 bits por línea
	var output strings.Builder
	for i := 0; i < len(bitmap); i++ {
		output.WriteString(fmt.Sprintf("%d", bitmap[i]))
		if (i+1)%20 == 0 {
			output.WriteString("\n")
		} else {
			output.WriteString(" ")
		}
	}

	// Escribir el contenido del bitmap en el archivo
	if _, err := reportFile.WriteString(output.String()); err != nil {
		return "Error: No se pudo escribir en el archivo de reporte"
	}

	return fmt.Sprintf("Reporte bitmap de inodos generado exitosamente en: %s", outputPath)
}

func GenerateBmInodeReport(file *os.File, superblock Partitions.Superblock, path string) error {
	// Abrir el archivo de salida para escribir el reporte
	outputFile, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("error al crear el archivo de reporte: %v", err)
	}
	defer outputFile.Close()

	// Calcular el número total de inodos
	totalInodes := superblock.S_inodes_count

	// Leer el bitmap de inodos
	bitmapInodes := make([]byte, totalInodes)
	if _, err := file.ReadAt(bitmapInodes, int64(superblock.S_bm_inode_start)); err != nil {
		return fmt.Errorf("error al leer el bitmap de inodos: %v", err)
	}

	// Escribir el reporte en el archivo
	var output strings.Builder
	output.WriteString(" ═════════════════════ BITMAP INODES ═════════════════════════ \n")
	for i := int32(0); i < totalInodes; i++ {
		// Convertir cada byte en bits
		bit := bitmapInodes[i]

		// Escribir cada bit en el output
		if bit == 0 {
			output.WriteString("0")
		} else {
			output.WriteString("1")
		}

		// Insertar un salto de línea después de cada 20 bits
		if (i+1)%20 == 0 {
			output.WriteString("\n")
		}
	}

	// Asegurar que el último salto de línea sea agregado si no se completó un múltiplo de 20
	if totalInodes%20 != 0 {
		output.WriteString("\n")
	}

	// Escribir el resultado en el archivo de salida
	if _, err := outputFile.WriteString(output.String()); err != nil {
		return fmt.Errorf("error al escribir en el archivo de reporte: %v", err)
	}

	return nil
}
