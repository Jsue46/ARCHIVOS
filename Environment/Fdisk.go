package Environment

import (
	"Proyecto/Partitions"
	"Proyecto/Utils"
	"encoding/binary"
	"fmt"
	"strings"
)

func Fdisk(size int, path string, name string, unit string, type_ string, fit string) string {

	var output strings.Builder
	output.WriteString("╔═════════════════════ INICIO  FDISK  ═════════════════════════╗\n")
	output.WriteString(fmt.Sprintf("  Tamaño: %d \n", size))
	output.WriteString(fmt.Sprintf("  Ruta: %s\n", path))
	output.WriteString(fmt.Sprintf("  Nombre: %s\n", name))
	output.WriteString(fmt.Sprintf("  Unidad: %s\n", unit))
	output.WriteString(fmt.Sprintf("  Tipo: %s\n", type_))
	output.WriteString(fmt.Sprintf("  Ajuste: %s\n", fit))

	if fit != "bf" && fit != "ff" && fit != "wf" {
		return "Error: El ajuste debe ser 'bf', 'ff' o 'wf'"
	}
	if size <= 0 {
		return "Error: El tamaño debe ser mayor a 0"
	}
	if unit != "b" && unit != "k" && unit != "m" {
		return "Error: La unidad debe ser 'b', 'k' o 'm'"
	}
	if type_ != "p" && type_ != "e" && type_ != "l" {
		return "Error: El tipo debe ser 'p', 'e' o 'l'"
	}
	if name == "" {
		return "Error: El nombre es obligatorio"
	}
	if path == "" {
		return "Error: La ruta es obligatoria"
	}
	// Abrir el archivo binario en la ruta proporcionada
	file, err := Utils.AbrirArchivo(path)
	if err != nil {
		return fmt.Sprintf("  Error: No se pudo abrir el archivo con la ruta proporcionada: %s", path)
	}
	defer file.Close()

	// Leer el MBR
	var TempMBR Partitions.MBR
	if err := Utils.LeerArchivo(file, &TempMBR, 0); err != nil {
		return "Error: No se pudo leer el MBR del archivo"
	}
	// Imprimir el objeto MBR
	output.WriteString(Partitions.ImprimirMBR(TempMBR))
	output.WriteString("───────────────────────────────────────────────────────────────── \n")

	// Validaciones de las particiones
	var primaryCount, extendedCount, totalPartitions int
	var usedSpace int32 = 0

	for i := 0; i < 4; i++ {
		if TempMBR.MbrPartitions[i].PartSize != 0 {
			totalPartitions++
			usedSpace += TempMBR.MbrPartitions[i].PartSize

			if TempMBR.MbrPartitions[i].PartType[0] == 'p' {
				primaryCount++
			} else if TempMBR.MbrPartitions[i].PartType[0] == 'e' {
				extendedCount++
			}
		}
	}
	// Validar que no se exceda el número máximo de particiones primarias y extendidas
	if totalPartitions >= 4 {
		return "Error: No se pueden crear más de 4 particiones primarias o extendidas en total."
	}
	// Validar que solo haya una partición extendida
	if type_ == "e" && extendedCount > 0 {
		return "Error: Solo se permite una partición extendida por disco."
	}
	// Validar que no se creen particiones extendidas antes de las primarias
	if type_ == "l" && extendedCount == 0 {
		return "Error: No se puede crear una partición lógica sin una partición extendida."
	}
	// Validar que el tamaño de la nueva partición no exceda el tamaño del disco
	if usedSpace+int32(size) > TempMBR.MbrTamanio {
		return "Error: No hay suficiente espacio en el disco para crear esta partición."
	}
	// Determinar la posición de inicio de la nueva partición
	var gap int32 = int32(binary.Size(TempMBR))
	if totalPartitions > 0 {
		gap = TempMBR.MbrPartitions[totalPartitions-1].PartStart + TempMBR.MbrPartitions[totalPartitions-1].PartSize
	}

	// Encontrar una posición vacía para la nueva partición
	for i := 0; i < 4; i++ {
		if TempMBR.MbrPartitions[i].PartSize == 0 {
			if type_ == "p" || type_ == "e" {
				TempMBR.MbrPartitions[i].PartSize = int32(size)
				TempMBR.MbrPartitions[i].PartStart = gap
				copy(TempMBR.MbrPartitions[i].PartName[:], name)
				copy(TempMBR.MbrPartitions[i].PartFit[:], fit)
				copy(TempMBR.MbrPartitions[i].PartStatus[:], "0")
				copy(TempMBR.MbrPartitions[i].PartType[:], type_)
				TempMBR.MbrPartitions[i].PartCorrelative = int32(totalPartitions + 1)

				if type_ == "e" {
					ebrStart := gap
					ebr := Partitions.EBR{
						PartFit:   fit[0],
						PartStart: ebrStart,
						PartSize:  0,
						PartNext:  -1,
					}
					copy(ebr.PartName[:], "")
					Utils.EscribirArchivo(file, ebr, int64(ebrStart))
				}

				break
			}
		}
	}

	// Mostrar las particiones creadas
	output.WriteString("Particiones en el disco:\n")
	for i := 0; i < 4; i++ {
		part := TempMBR.MbrPartitions[i]
		if part.PartSize > 0 {
			// Mostrar información de la partición existente
			output.WriteString(fmt.Sprintf(" Partición %d:\n", i+1))
			output.WriteString(fmt.Sprintf("  Nombre: %s", strings.TrimSpace(string(part.PartName[:]))))
			output.WriteString(fmt.Sprintf(" Tipo: %s", strings.TrimSpace(string(part.PartType[:]))))
			output.WriteString(fmt.Sprintf(" Start: %d", part.PartStart))
			output.WriteString(fmt.Sprintf(" Tamaño: %d", part.PartSize))
			output.WriteString(fmt.Sprintf(" Status: %s", strings.TrimSpace(string(part.PartStatus[:]))))
			output.WriteString(fmt.Sprintf(" Id: %s\n", strings.TrimSpace(string(part.PartID[:]))))
		} else {
			// Mostrar información de una partición vacía con valores predeterminados
			output.WriteString(fmt.Sprintf(" Partición %d:\n", i+1))
			output.WriteString("  Nombre: null")
			output.WriteString(" Tipo: null")
			output.WriteString(" Start: 0")
			output.WriteString(" Tamaño: 0")
			output.WriteString(" Status: null ")
			output.WriteString(" Id: 0\n")
		}
	}

	// Sobrescribir el MBR
	if err := Utils.EscribirArchivo(file, TempMBR, 0); err != nil {
		return "Error: No se pudo escribir el MBR en el archivo"
	}

	var TempMBR2 Partitions.MBR
	// Leer el objeto nuevamente para verificar
	if err := Utils.LeerArchivo(file, &TempMBR2, 0); err != nil {
		return "Error: No se pudo leer el MBR del archivo después de escribirlo"
	}

	// Imprimir el objeto MBR actualizado
	output.WriteString(Partitions.ImprimirMBR(TempMBR2))
	output.WriteString("╚═════════════════════   FIN   FDISK   ═════════════════════════╝\n")

	return output.String()
}
