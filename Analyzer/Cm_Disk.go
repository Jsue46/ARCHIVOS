package Analyzer

import (
	"Proyecto/Environment"
	"fmt"
	"strconv"
	"strings"
)

// fn_mkdisk procesa el comando mkdisk y devuelve string en lugar de imprimir
func fn_mkdisk(parametros string) string {
	paramMap := extraerParametros(parametros)

	var output strings.Builder

	// Lista de parámetros válidos
	validParams := map[string]bool{
		"size": true,
		"fit":  true,
		"unit": true,
		"path": true,
	}

	// Validar si hay parámetros no reconocidos
	for param := range paramMap {
		if !validParams[param] {
			return fmt.Sprintf("Error: El parámetro '%s' no es válido", param)
		}
	}

	// Validar parámetro size
	sizeStr, ok := paramMap["size"]
	if !ok {
		return "Error: El parámetro 'size' es obligatorio"
	}
	size, err := strconv.Atoi(sizeStr)
	if err != nil || size <= 0 {
		return "Error: El valor de 'size' debe ser un número entero mayor que 0"
	}

	// Procesar parámetro fit (opcional)
	fit := strings.ToLower(paramMap["fit"])
	if fit == "" {
		fit = "ff" // Valor por defecto
	} else if fit != "bf" && fit != "ff" && fit != "wf" {
		return "Error: El valor de 'fit' debe ser 'bf', 'ff' o 'wf'"
	}

	// Procesar parámetro unit (opcional)
	unit := strings.ToLower(paramMap["unit"])
	if unit == "" {
		unit = "m" // Valor por defecto
	} else if unit != "k" && unit != "m" {
		return "Error: El valor de 'unit' debe ser 'k' o 'm'"
	}

	// Validar parámetro path (obligatorio)
	path := paramMap["path"]
	if path == "" {
		return "Error: El parámetro 'path' es obligatorio"
	}

	// Llamar a la función Mkdisk y capturar su salida
	output.WriteString(Environment.Mkdisk(size, fit, unit, path))

	return output.String()
}

// fn_rmdisk procesa el comando rmdisk.
func fn_rmdisk(parametros string) string {
	paramMap := extraerParametros(parametros)
	var output strings.Builder

	path := strings.ToLower(paramMap["path"])
	if path == "" {
		return "Error: El parámetro 'path' es obligatorio"
	}

	// Llamar a la función Rmdisk del paquete Environment
	output.WriteString(Environment.Rmdisk(path))
	return output.String()
}

// fn_fdisk procesa el comando fdisk.
func fn_fdisk(parametros string) string {
	paramMap := extraerParametros(parametros)
	var output strings.Builder

	// Validar y procesar parámetros
	unit := strings.ToLower(paramMap["unit"])
	if unit == "" {
		unit = "k" // Valor por defecto
	} else if unit != "b" && unit != "k" && unit != "m" {
		return "Error: La unidad debe ser 'b', 'k' o 'm'"
	}

	fit := strings.ToLower(paramMap["fit"])
	if fit == "" {
		fit = "wf" // Valor por defecto
	} else if fit != "bf" && fit != "ff" && fit != "wf" {
		return "Error: El ajuste debe ser 'bf', 'ff' o 'wf'"
	}

	partType := strings.ToLower(paramMap["type"])
	if partType == "" {
		partType = "p" // Valor por defecto
	} else if partType != "p" && partType != "e" && partType != "l" {
		return "Error: El tipo de partición debe ser 'p', 'e' o 'l'"
	}

	size, err := strconv.Atoi(paramMap["size"])
	if err != nil || size <= 0 {
		return "Error: El valor de 'size' debe ser un número entero mayor que 0"
	}

	name := strings.ToLower(paramMap["name"])
	if name == "" {
		return "Error: El nombre de la partición es obligatorio"
	}

	path := strings.ToLower(paramMap["path"])
	if path == "" {
		return "Error: El parámetro 'path' es obligatorio"
	}

	// Convertir el tamaño a bytes
	sizeInBytes := size
	switch unit {
	case "k":
		sizeInBytes *= 1024
	case "m":
		sizeInBytes *= 1024 * 1024
	}

	// Llamar a la función FDISK con los parámetros procesados
	output.WriteString(Environment.Fdisk(sizeInBytes, path, name, unit, partType, fit))
	return output.String()
}

// fn_mount procesa el comando mount.
func fn_mount(parametros string) string {
	var output strings.Builder
	paramMap := extraerParametros(parametros)

	path := strings.ToLower(paramMap["path"])
	name := strings.ToLower(paramMap["name"])

	if path == "" || name == "" {
		return "Error: Path y Name son obligatorios"
	}

	// Llamar a la función Mount con los parámetros procesados
	output.WriteString(Environment.Mount(path, name))
	return output.String()
}

// fn_mounted procesa el comando mounted.
func fn_mounted(_ string) string {
	var output strings.Builder
	mountedPartitions := Environment.GetMountedPartitions()

	// Verificar si hay particiones montadas
	if len(mountedPartitions) == 0 {
		return "No hay Particiones Montadas."
	}

	// Mostrar los IDs de las particiones montadas
	output.WriteString("╔═════════════════ PARTICIONES MONTADAS ═══════════════════════╗\n")
	for disk, partitions := range mountedPartitions {
		output.WriteString(fmt.Sprintf("  Disco: %s\n", disk))
		for _, partition := range partitions {
			output.WriteString(fmt.Sprintf("    - ID: %s\n", partition.MountID))
		}
	}
	output.WriteString("╚═════════════════════ FIN PARTICIONES ════════════════════════╝")
	return output.String()
}
