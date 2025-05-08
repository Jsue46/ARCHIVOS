package Reportes

import (
	"Proyecto/Environment"
	"Proyecto/Partitions"
	"Proyecto/User_Groups"
	"Proyecto/Utils"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func GenerarReporteBloques(pathFileLs string, outputPath string, id string) string {
	if pathFileLs == "" {
		pathFileLs = "/users.txt"
	}

	// Obtener la partición montada por ID
	mountedPartition, found := Environment.GetMountedPartitionByID(id)
	if !found {
		return fmt.Sprintf("Error: No se encontró la partición con ID %s montada", id)
	}

	// Abrir el archivo binario del disco montado
	file, err := Utils.AbrirArchivo(mountedPartition.MountPath)
	if err != nil {
		return fmt.Sprintf("Error: No se pudo abrir el archivo en la ruta: %s", mountedPartition.MountPath)
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
		if TempMBR.MbrPartitions[i].PartSize != 0 && strings.Contains(string(TempMBR.MbrPartitions[i].PartID[:]), id) {
			if TempMBR.MbrPartitions[i].PartStatus[0] == '1' {
				index = i
			} else {
				return "Error: La partición no está montada"
			}
			break
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

	// Buscar el inodo correspondiente al path para obtener los bloques
	inodeNumber, _ := User_Groups.InitSearch(pathFileLs, file, superblock)
	if inodeNumber == -1 {
		return fmt.Sprintf("Error: No se pudo encontrar el inodo para la ruta: %s", pathFileLs)
	}

	// Leer el inodo específico
	var inode Partitions.Inode
	inodeStart := superblock.S_inode_start + inodeNumber*int32(binary.Size(Partitions.Inode{}))
	if err := Utils.LeerArchivo(file, &inode, int64(inodeStart)); err != nil {
		return fmt.Sprintf("Error al leer el inodo: %v", err)
	}

	// Generar reporte para los bloques asociados al inodo
	for i := 0; i < len(inode.I_block); i++ {
		if inode.I_block[i] == -1 {
			continue
		}

		// Determinar el tipo de bloque y leerlo
		blockNumber := inode.I_block[i]
		blockType := determinarTipoBloque(i)
		fmt.Printf("Tipo de bloque: %s\n", blockType)

		// Leer el bloque de la partición
		var block Partitions.Folderblock // O Fileblock dependiendo del tipo de bloque
		blockStart := superblock.S_block_start + blockNumber*int32(binary.Size(Partitions.Folderblock{}))
		if err := Utils.LeerArchivo(file, &block, int64(blockStart)); err != nil {
			return fmt.Sprintf("Error al leer el bloque: %v", err)
		}

		// Generar reporte del bloque
		if err := GenerateBlockReport(block, outputPath, blockNumber); err != nil {
			return fmt.Sprintf("Error al generar el reporte del bloque: %v", err)
		}
	}

	return fmt.Sprintf("Reporte de bloques generado exitosamente en: %s", outputPath)
}

func determinarTipoBloque(i int) string {
	if i < 12 {
		return "Directo"
	} else if i == 12 {
		return "Indirecto simple"
	} else if i == 13 {
		return "Indirecto doble"
	} else if i == 14 {
		return "Indirecto triple"
	}
	return ""
}

// Función para generar el reporte de un bloque de carpeta en formato .dot
func GenerateBlockReport(block Partitions.Folderblock, outputPath string, blockNumber int32) error {
	// Obtener la ruta del directorio donde se guardará el reporte
	reportsDir := filepath.Dir(outputPath)
	// Crear la carpeta si no existe
	err := os.MkdirAll(reportsDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error al crear la carpeta de reportes: %v", err)
	}

	// Crear el archivo .dot donde se generará el reporte
	dotFilePath := strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + ".dot"
	dotFile, err := os.Create(dotFilePath)
	if err != nil {
		return fmt.Errorf("error al crear el archivo .dot de reporte: %v", err)
	}
	defer dotFile.Close()

	// Escribir contenido inicial para el archivo .dot (definición de grafo)
	dotFile.WriteString("digraph BlockReport {\n")
	dotFile.WriteString(fmt.Sprintf("\tlabel=\"Reporte del Bloque %d\";\n", blockNumber))
	dotFile.WriteString("\tnode [shape=plaintext];\n")

	// Definir tabla para el bloque de carpeta con colores
	tableColor := "#f39c12"
	headerColor := "#2ecc71"
	rowEvenColor := "#ecf0f1"
	rowOddColor := "#bdc3c7"

	// Reportar el contenido del bloque
	dotFile.WriteString(fmt.Sprintf("\tbl%d [label=<\n", blockNumber))
	dotFile.WriteString(fmt.Sprintf("<TABLE BORDER=\"1\" CELLBORDER=\"1\" CELLSPACING=\"0\" BGCOLOR=\"%s\">\n", tableColor))
	dotFile.WriteString(fmt.Sprintf("<TR><TD COLSPAN=\"2\" BGCOLOR=\"%s\"><B>Bloque Carpeta %d</B></TD></TR>\n", headerColor, blockNumber))
	dotFile.WriteString("<TR><TD><B>b_name</B></TD><TD><B>b_inodo</B></TD></TR>\n")

	// Agregar las entradas del bloque con colores alternos para las filas
	for i, content := range block.B_content {
		rowColor := rowEvenColor
		if i%2 != 0 {
			rowColor = rowOddColor
		}

		if content.B_inodo != -1 && cleanString(content.B_name[:]) != "" {
			// Si el bloque tiene contenido válido
			dotFile.WriteString(fmt.Sprintf("<TR BGCOLOR=\"%s\"><TD>%s</TD><TD>%d</TD></TR>\n", rowColor, cleanString(content.B_name[:]), content.B_inodo))
		} else {
			// Si la entrada está vacía
			dotFile.WriteString(fmt.Sprintf("<TR BGCOLOR=\"%s\"><TD>Entrada %d</TD><TD>Vacío</TD></TR>\n", rowColor, i+1))
		}
	}

	dotFile.WriteString("</TABLE>\n")
	dotFile.WriteString(">];\n")

	// Completar el archivo .dot (cerrar grafo)
	dotFile.WriteString("}\n")

	// Generar la imagen a partir del archivo .dot utilizando Graphviz
	cmd := exec.Command("dot", "-Tjpg", dotFilePath, "-o", outputPath)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error al generar el reporte gráfico: %v", err)
	}

	return nil
}
