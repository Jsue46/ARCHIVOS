package Environment

import (
	"Proyecto/Partitions"
	"Proyecto/Utils"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

func Mkdisk(size int, fit string, unit string, path string) string {
	var output strings.Builder

	output.WriteString("╔═════════════════════ INICIO MKDISK ═════════════════════════╗\n")
	output.WriteString(fmt.Sprintf("  Size: %d\n  Fit: %s\n  Unit: %s\n  Path: %s\n", size, fit, unit, path))

	// Validaciones
	if fit != "bf" && fit != "wf" && fit != "ff" {
		return "  Error: Fit debe ser 'bf', 'wf' o 'ff'"
	}
	if size <= 0 {
		return "  Error: Size debe ser mayor a 0"
	}
	if unit != "k" && unit != "m" {
		return "  Error: Las unidades válidas son 'k' o 'm'"
	}

	// Crear directorios
	if err := os.MkdirAll(path[:strings.LastIndex(path, "/")], os.ModePerm); err != nil {
		return fmt.Sprintf("  Error al crear directorios: %s", err.Error())
	}

	// Crear archivo
	if err := Utils.CrearArchivo(path); err != nil {
		return fmt.Sprintf("  Error al crear archivo: %s", err.Error())
	}

	// Convertir tamaño a bytes
	sizeInBytes := size * 1024
	if unit == "m" {
		sizeInBytes *= 1024
	}

	// Abrir archivo
	file, err := Utils.AbrirArchivo(path)
	if err != nil {
		return fmt.Sprintf("  Error al abrir archivo: %s", err.Error())
	}
	defer file.Close()

	// Escribir ceros
	zeroBlock := make([]byte, sizeInBytes)
	if _, err := file.Write(zeroBlock); err != nil {
		return fmt.Sprintf("  Error al escribir en el archivo: %s", err.Error())
	}

	// Crear MBR
	var nuevoMBR Partitions.MBR
	nuevoMBR.MbrTamanio = int32(sizeInBytes)
	nuevoMBR.MbrDskSignature = rand.Int31()
	copy(nuevoMBR.DskFit[:], fit)

	formattedDate := time.Now().Format("02/01/2006")
	copy(nuevoMBR.MbrFechaCreacion[:], formattedDate)

	// Escribir MBR
	if err := Utils.EscribirArchivo(file, nuevoMBR, 0); err != nil {
		return fmt.Sprintf("  Error al escribir el MBR: %s", err.Error())
	}

	// Leer MBR para verificación
	var tempMBR Partitions.MBR
	if err := Utils.LeerArchivo(file, &tempMBR, 0); err != nil {
		return fmt.Sprintf("  Error al leer el MBR: %s", err.Error())
	}

	// Generar salida
	output.WriteString("\n  MBR creado exitosamente:\n")
	output.WriteString(Partitions.ImprimirMBR(tempMBR))
	output.WriteString("╚═════════════════════ FIN MKDISK ════════════════════════════╝\n")

	return output.String()
}
