package Environment

import (
	"Proyecto/Partitions"
	"Proyecto/Utils"
	"bytes"
	"fmt"
	"strings"
)

func Mount(path string, name string) string {
	var output strings.Builder
	output.WriteString("╔══════════════════════ INICIO MOUNT  ══════════════════════════╗\n")
	file, err := Utils.AbrirArchivo(path)
	if err != nil {
		return fmt.Sprintf("  Error: No se pudo abrir el archivo en la ruta: %s\n╚══════════════════════   FIN MOUNT   ═════════════════════════╝", path)
	}
	defer file.Close()

	var TempMBR Partitions.MBR
	if err := Utils.LeerArchivo(file, &TempMBR, 0); err != nil {
		return "Error: No se pudo leer el MBR desde el archivo\n╚══════════════════════   FIN MOUNT   ═════════════════════════╝"
	}

	output.WriteString(fmt.Sprintf("  Buscando partición con nombre: '%s'\n", name))

	partitionFound := false
	var partition Partitions.Partition
	var partitionIndex int

	// Convertir el nombre a comparar a un arreglo de bytes de longitud fija
	nameBytes := [16]byte{}
	copy(nameBytes[:], []byte(name))

	for i := 0; i < 4; i++ {
		if TempMBR.MbrPartitions[i].PartType[0] == 'p' && bytes.Equal(TempMBR.MbrPartitions[i].PartName[:], nameBytes[:]) {
			partition = TempMBR.MbrPartitions[i]
			partitionIndex = i
			partitionFound = true
			break
		}
	}

	if !partitionFound {
		return "Error: Partición no encontrada o no es una partición primaria\n╚══════════════════════   FIN MOUNT   ═════════════════════════╝"
	}

	// Verificar si la partición ya está montada en el MBR
	if partition.PartStatus[0] == '1' {
		return "Error: La partición ya está montada\n╚══════════════════════   FIN MOUNT   ═════════════════════════╝"
	}

	output.WriteString(fmt.Sprintf("  Partición encontrada: '%s' en posición %d\n", strings.TrimSpace(string(partition.PartName[:])), partitionIndex+1))

	// Verificar si la partición ya está en la lista de particiones montadas
	for _, mountedDiskPartitions := range mountedPartitions {
		for _, mp := range mountedDiskPartitions {
			if mp.MountPath == path && mp.MountName == name {
				output.WriteString(fmt.Sprintf("  AVISO: La partición '%s' ya está montada con ID: %s\n", name, mp.MountID))
				output.WriteString("╚══════════════════════   FIN MOUNT   ═════════════════════════╝\n")
				return output.String()
			}
		}
	}

	// Generar el ID de la partición utilizando la función `generateDiskID`
	diskID := generateDiskID(path)

	// Crear el ID de la partición utilizando el último par de dígitos de un carnet
	carnet := "202201336"
	lastTwoDigits := carnet[len(carnet)-2:]

	// Determinar la letra para el ID
	var letter byte
	if _, exists := mountedPartitions[diskID]; !exists || len(mountedPartitions[diskID]) == 0 {
		// Si es el primer disco o el disco no tiene particiones montadas
		if len(mountedPartitions) == 0 {
			letter = 'A'
		} else {
			// Buscar la última letra usada
			lastDiskID := getLastDiskID()
			if len(mountedPartitions[lastDiskID]) > 0 {
				lastLetter := mountedPartitions[lastDiskID][0].MountID[len(mountedPartitions[lastDiskID][0].MountID)-1]
				letter = lastLetter + 1
			} else {
				letter = 'A'
			}
		}
	} else {
		// Usar la misma letra que las otras particiones del disco
		letter = mountedPartitions[diskID][0].MountID[len(mountedPartitions[diskID][0].MountID)-1]
	}

	partitionID := fmt.Sprintf("%s%d%c", lastTwoDigits, partitionIndex+1, letter)

	// Actualizar el status de la partición en el MBR
	partition.PartStatus[0] = '1'
	copy(partition.PartID[:], partitionID)
	TempMBR.MbrPartitions[partitionIndex] = partition

	// Añadir la partición a la lista de particiones montadas
	newMountedPartition := MountedPartition{
		MountPath:   path,
		MountName:   name,
		MountID:     partitionID,
		MountStatus: '1',
		LoggedIn:    false,
		MountStart:  partition.PartStart,
	}

	// Usar la función auxiliar para actualizar si ya existe
	if !actualizarParticionMontada(path, name, newMountedPartition) {
		// Si no existe, añadirla
		mountedPartitions[diskID] = append(mountedPartitions[diskID], newMountedPartition)
	}

	// Escribir el MBR actualizado en el archivo
	if err := Utils.EscribirArchivo(file, TempMBR, 0); err != nil {
		return "Error: No se pudo sobrescribir el MBR en el archivo\n╚══════════════════════   FIN MOUNT   ═════════════════════════╝"
	}

	// Imprimir el mensaje confirmando que la partición ha sido montada, junto con su ID.
	output.WriteString(fmt.Sprintf("  Partición montada con ID: %s\n", partitionID))

	output.WriteString("\n  MBR actualizado:\n")
	output.WriteString(Partitions.ImprimirMBR(TempMBR))
	output.WriteString("\n\n  Particiones Montadas:\n")
	output.WriteString(ImprimirMountedPartitions())

	output.WriteString("╚══════════════════════   FIN MOUNT   ═════════════════════════╝\n")
	return output.String()
}

func getLastDiskID() string {
	var lastDiskID string
	for diskID := range mountedPartitions {
		lastDiskID = diskID
	}
	return lastDiskID
}

func generateDiskID(path string) string {
	return strings.ToLower(path)
}

// Función para buscar y actualizar una partición existente
func actualizarParticionMontada(path string, name string, newPartition MountedPartition) bool {
	for diskID, partitions := range mountedPartitions {
		for i, partition := range partitions {
			if partition.MountPath == path && partition.MountName == name {
				// Actualizar la partición existente
				mountedPartitions[diskID][i] = newPartition
				return true
			}
		}
	}
	return false
}
