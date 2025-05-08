package Reportes

import (
	"Proyecto/Partitions"
	"Proyecto/Utils"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func GenerarReporteMBR(diskPath, reportPath string) string {
	var output strings.Builder
	// Abrir el archivo binario del disco montado
	file, err := Utils.AbrirArchivo(diskPath)
	if err != nil {
		return fmt.Sprintf("Error: No se pudo abrir el archivo en la ruta: %s", diskPath)
	}
	defer file.Close()

	// Leer el objeto MBR desde el archivo binario
	var TempMBR Partitions.MBR
	if err := Utils.LeerArchivo(file, &TempMBR, 0); err != nil {
		return "Error: No se pudo leer el MBR desde el archivo"
	}

	// Leer y procesar los EBRs si hay particiones extendidas
	var ebrs []Partitions.EBR
	for i := 0; i < 4; i++ {
		if string(TempMBR.MbrPartitions[i].PartType[:]) == "e" { // Partición extendida
			ebrPosition := TempMBR.MbrPartitions[i].PartStart
			for ebrPosition != -1 {
				var tempEBR Partitions.EBR
				if err := Utils.LeerArchivo(file, &tempEBR, int64(ebrPosition)); err != nil {
					output.WriteString("Error: No se pudo leer el EBR desde el archivo")
					break
				}
				ebrs = append(ebrs, tempEBR)
				ebrPosition = tempEBR.PartNext
			}
		}
	}

	// Generar el archivo .dot del MBR con EBRs
	if err := GenerateMBRReport(TempMBR, ebrs, reportPath, file); err != nil {
		output.WriteString(fmt.Sprintf("Error al generar el reporte MBR: %v", err))
	} else {
		output.WriteString(fmt.Sprintf("Reporte MBR generado exitosamente en: %s", reportPath))
		// Renderizar el archivo .dot a .jpg usando Graphviz
		dotFile := strings.TrimSuffix(reportPath, filepath.Ext(reportPath)) + ".dot"
		outputJpg := reportPath
		cmd := exec.Command("dot", "-Tjpg", dotFile, "-o", outputJpg)
		err = cmd.Run()
		if err != nil {
			output.WriteString(fmt.Sprintf("Error al renderizar el archivo .dot a imagen: %v", err))
		} else {
			output.WriteString(fmt.Sprintf("Imagen generada exitosamente en: %s", outputJpg))
		}
	}
	return output.String()
}

// GenerateMBRReport genera un reporte del MBR y los EBRs
func GenerateMBRReport(mbr Partitions.MBR, ebrs []Partitions.EBR, outputPath string, file *os.File) error {
	reportsDir := filepath.Dir(outputPath)
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

	// Iniciar el contenido del archivo en formato Graphviz
	graphContent := "digraph G {\n"
	graphContent += "\tnode [fillcolor=lightyellow style=filled]\n"
	graphContent += "\trankdir=LR;\n"

	// Crear una sola tabla para el MBR y las particiones
	graphContent += "\tsubgraph cluster_MBR {\n"
	graphContent += "\t\tcolor=lightblue fillcolor=aliceblue label=\"MBR Report\" style=filled\n"
	graphContent += "\t\tmbr [label=<<table border=\"0\" cellborder=\"1\" cellspacing=\"0\" cellpadding=\"4\">\n"

	// Encabezado de la tabla
	graphContent += "\t\t\t<tr><td colspan=\"2\" bgcolor=\"cadetblue\"><b>MBR Information</b></td></tr>\n"
	graphContent += fmt.Sprintf("\t\t\t<tr><td><b>Tamaño</b></td><td>%d</td></tr>\n", mbr.MbrTamanio)
	graphContent += fmt.Sprintf("\t\t\t<tr><td><b>Fecha Creación</b></td><td>%s</td></tr>\n", strings.TrimRight(string(mbr.MbrFechaCreacion[:]), "\x00"))
	graphContent += fmt.Sprintf("\t\t\t<tr><td><b>Disk Signature</b></td><td>%d</td></tr>\n", mbr.MbrDskSignature)

	// Iterar sobre las 4 particiones del MBR
	for i := 0; i < 4; i++ {
		partition := mbr.MbrPartitions[i]
		if partition.PartSize > 0 {
			partitionName := strings.TrimRight(string(partition.PartName[:]), "\x00")

			// Determinar el color del encabezado
			var headerColor string
			switch string(partition.PartType[:]) {
			case "p":
				headerColor = "lightgreen"
			case "e":
				headerColor = "lightblue"
			case "l":
				headerColor = "lightyellow"
			default:
				headerColor = "white"
			}

			graphContent += fmt.Sprintf("\t\t\t<tr><td colspan=\"2\" bgcolor=\"%s\"><b>Partición %d</b></td></tr>\n", headerColor, i+1)
			graphContent += fmt.Sprintf("\t\t\t<tr><td><b>Status</b></td><td>%s</td></tr>\n", string(partition.PartStatus[:]))
			graphContent += fmt.Sprintf("\t\t\t<tr><td><b>Type</b></td><td>%s</td></tr>\n", string(partition.PartType[:]))
			graphContent += fmt.Sprintf("\t\t\t<tr><td><b>Fit</b></td><td>%s</td></tr>\n", string(partition.PartFit[:]))
			graphContent += fmt.Sprintf("\t\t\t<tr><td><b>Start</b></td><td>%d</td></tr>\n", partition.PartStart)
			graphContent += fmt.Sprintf("\t\t\t<tr><td><b>Size</b></td><td>%d</td></tr>\n", partition.PartSize)
			graphContent += fmt.Sprintf("\t\t\t<tr><td><b>Name</b></td><td>%s</td></tr>\n", partitionName)

			// Manejar particiones extendidas y sus EBRs
			if string(partition.PartType[:]) == "e" {
				graphContent += fmt.Sprintf("\t\t\t<tr><td colspan=\"2\" bgcolor=\"lightpink\"><b>EBRs de la Partición Extendida %d</b></td></tr>\n", i+1)

				// Leer los EBRs en la partición extendida
				ebrPosition := partition.PartStart
				for {
					var ebr Partitions.EBR
					err := Utils.LeerArchivo(file, &ebr, int64(ebrPosition))
					if err != nil {
						break
					}

					ebrName := strings.TrimRight(string(ebr.PartName[:]), "\x00")
					graphContent += fmt.Sprintf("\t\t\t<tr><td><b>EBR Start</b></td><td>%d</td></tr>\n", ebr.PartStart)
					graphContent += fmt.Sprintf("\t\t\t<tr><td><b>EBR Size</b></td><td>%d</td></tr>\n", ebr.PartSize)
					graphContent += fmt.Sprintf("\t\t\t<tr><td><b>EBR Next</b></td><td>%d</td></tr>\n", ebr.PartNext)
					graphContent += fmt.Sprintf("\t\t\t<tr><td><b>EBR Name</b></td><td>%s</td></tr>\n", ebrName)

					// Verificar si hay más EBRs
					if ebr.PartNext == -1 {
						break
					}
					ebrPosition = ebr.PartNext
				}
			}
		}
	}

	graphContent += "\t\t</table>> shape=plaintext]\n"
	graphContent += "\t}\n"
	graphContent += "}\n"

	// Escribir el contenido en el archivo .dot
	_, err = dotFile.WriteString(graphContent)
	if err != nil {
		return fmt.Errorf("error al escribir en el archivo .dot: %v", err)
	}

	return nil
}
