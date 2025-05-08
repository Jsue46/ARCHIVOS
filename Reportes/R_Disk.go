package Reportes

import (
	"Proyecto/Partitions"
	"Proyecto/Utils"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func GenerarReporteDisk(diskPath, reportPath string) string {
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
					break
				}
				ebrs = append(ebrs, tempEBR)
				ebrPosition = tempEBR.PartNext
			}
		}
	}

	// Calcular el tamaño total del disco
	totalDiskSize := TempMBR.MbrTamanio

	// Generar el archivo .dot del DISK
	if err := GenerateDiskReport(TempMBR, ebrs, reportPath, file, totalDiskSize); err != nil {
		output.WriteString(fmt.Sprintf("Error al generar el reporte DISK: %v", err))
	} else {
		output.WriteString(fmt.Sprintf("Reporte DISK generado exitosamente en: %s", reportPath))
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

// Función para generar el reporte DISK en formato .dot con colores
func GenerateDiskReport(mbr Partitions.MBR, ebrs []Partitions.EBR, outputPath string, file *os.File, totalDiskSize int32) error {
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

	// Calcular tamaño real del MBR
	mbrSize := int32(binary.Size(mbr))

	// Iniciar el contenido del archivo en formato Graphviz (.dot) con estilos
	graphContent := `digraph G {
        node [shape=none];
        graph [splines=false];
        rankdir=LR;
    `

	// Título del disco con estilo
	graphContent += `
        subgraph cluster_disk {
            label="Disco1.dsk";
            style="rounded,filled";
            fillcolor="#bdc3c7";
            color=black;
            fontcolor=black;
            fontsize=20;
    `

	// Iniciar tabla para las particiones con estilos
	graphContent += `
            disk [label=<
                <TABLE BORDER="0" CELLBORDER="1" CELLSPACING="0" CELLPADDING="6">
                    <TR>
    `

	// 1. MBR con color azul
	mbrPercentage := float64(mbrSize) / float64(totalDiskSize) * 100
	graphContent += fmt.Sprintf(`
                        <TD COLSPAN="1" BORDER="1" BGCOLOR="#3498db">
                            <FONT COLOR="white" POINT-SIZE="14"><B>MBR</B></FONT><BR/>
                            <FONT POINT-SIZE="12">%d bytes</FONT><BR/>
                            <FONT POINT-SIZE="12">%.2f%%</FONT>
                        </TD>
    `, mbrSize, mbrPercentage)

	// Variables para el espacio utilizado
	var usedSpace int32 = mbrSize
	var extendedSpace int32 = 0

	// Procesar las particiones primarias y extendidas
	for i := 0; i < 4; i++ {
		partition := mbr.MbrPartitions[i]
		if partition.PartSize > 0 {
			percentage := float64(partition.PartSize) / float64(totalDiskSize) * 100
			partitionName := strings.TrimRight(string(partition.PartName[:]), "\x00")

			if string(partition.PartType[:]) == "p" { // Partición primaria (verde)
				graphContent += fmt.Sprintf(`
                            <TD COLSPAN="1" BORDER="1" BGCOLOR="#2ecc71">
                                <FONT COLOR="white" POINT-SIZE="14"><B>Primaria</B></FONT><BR/>
                                <FONT POINT-SIZE="12">%s</FONT><BR/>
                                <FONT POINT-SIZE="12">%d bytes</FONT><BR/>
                                <FONT POINT-SIZE="12">%.2f%%</FONT>
                            </TD>
                `, partitionName, partition.PartSize, percentage)
				usedSpace += partition.PartSize
			} else if string(partition.PartType[:]) == "e" { // Partición extendida (naranja)
				extendedSpace = partition.PartSize
				graphContent += `
                            <TD COLSPAN="1" BORDER="1" BGCOLOR="#f39c12">
                                <TABLE BORDER="0" CELLBORDER="1" CELLSPACING="0" CELLPADDING="6">
                                    <TR>
                                        <TD COLSPAN="3" BGCOLOR="#f39c12">
                                            <FONT COLOR="white" POINT-SIZE="14"><B>Extendida</B></FONT><BR/>
                                            ` + fmt.Sprintf(`<FONT POINT-SIZE="12">%d bytes</FONT><BR/><FONT POINT-SIZE="12">%.2f%%</FONT>`, extendedSpace, percentage) + `
                                        </TD>
                                    </TR>
                                    <TR>
                `

				// Procesar las particiones lógicas dentro de la extendida
				var ebrSpace int32 = 0
				for _, ebr := range ebrs {
					if ebr.PartSize > 0 {
						logicalPercentage := float64(ebr.PartSize) / float64(totalDiskSize) * 100
						ebrSize := int32(binary.Size(ebr))
						ebrPercentage := float64(ebrSize) / float64(totalDiskSize) * 100

						// EBR (naranja oscuro)
						graphContent += fmt.Sprintf(`
                                            <TD COLSPAN="1" BORDER="1" BGCOLOR="#e67e22">
                                                <FONT COLOR="white" POINT-SIZE="14"><B>EBR</B></FONT><BR/>
                                                <FONT POINT-SIZE="12">%d bytes</FONT><BR/>
                                                <FONT POINT-SIZE="12">%.2f%%</FONT>
                                            </TD>
                        `, ebrSize, ebrPercentage)

						// Partición lógica (rojo)
						graphContent += fmt.Sprintf(`
                                            <TD COLSPAN="1" BORDER="1" BGCOLOR="#e74c3c">
                                                <FONT COLOR="white" POINT-SIZE="14"><B>Lógica</B></FONT><BR/>
                                                <FONT POINT-SIZE="12">%s</FONT><BR/>
                                                <FONT POINT-SIZE="12">%d bytes</FONT><BR/>
                                                <FONT POINT-SIZE="12">%.2f%%</FONT>
                                            </TD>
                        `, strings.TrimRight(string(ebr.PartName[:]), "\x00"), ebr.PartSize, logicalPercentage)

						ebrSpace += ebr.PartSize + ebrSize
					}
				}

				// Espacio libre dentro de la extendida (gris claro)
				freeExtended := extendedSpace - ebrSpace
				if freeExtended > 0 {
					freeExtendedPercentage := float64(freeExtended) / float64(totalDiskSize) * 100
					graphContent += fmt.Sprintf(`
                                            <TD COLSPAN="1" BORDER="1" BGCOLOR="#ecf0f1">
                                                <FONT POINT-SIZE="14"><B>Libre</B></FONT><BR/>
                                                <FONT POINT-SIZE="12">%d bytes</FONT><BR/>
                                                <FONT POINT-SIZE="12">%.2f%%</FONT>
                                            </TD>
                    `, freeExtended, freeExtendedPercentage)
				}

				graphContent += `
                                    </TR>
                                </TABLE>
                            </TD>
                `
				usedSpace += extendedSpace
			}
		}
	}

	// Espacio libre fuera de las particiones (gris claro)
	freeSpace := totalDiskSize - usedSpace
	if freeSpace > 0 {
		freePercentage := float64(freeSpace) / float64(totalDiskSize) * 100
		graphContent += fmt.Sprintf(`
                            <TD COLSPAN="1" BORDER="1" BGCOLOR="#ecf0f1">
                                <FONT POINT-SIZE="14"><B>Libre</B></FONT><BR/>
                                <FONT POINT-SIZE="12">%d bytes</FONT><BR/>
                                <FONT POINT-SIZE="12">%.2f%%</FONT>
                            </TD>
        `, freeSpace, freePercentage)
	}

	graphContent += `
                    </TR>
                </TABLE>
            >];
        }
    }
    `

	// Escribir el contenido en el archivo
	_, err = dotFile.WriteString(graphContent)
	if err != nil {
		return fmt.Errorf("error al escribir en el archivo .dot: %v", err)
	}

	// Convertir el archivo DOT a PDF
	pdfPath := strings.TrimSuffix(dotFilePath, ".dot") + ".pdf"
	cmd := exec.Command("dot", "-Tpdf", dotFilePath, "-o", pdfPath)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error al convertir DOT a PDF: %v", err)
	}

	return nil
}
