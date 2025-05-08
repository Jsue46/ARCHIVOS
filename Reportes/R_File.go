package Reportes

import (
	"Proyecto/Environment"
	"Proyecto/Partitions"
	"Proyecto/User_Groups"
	"Proyecto/Utils"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func GenerarReporteFile(pathFileLs, path, id string) string {
	// Verificar sesión activa
	if !User_Groups.IsUserLoggedIn() {
		return "Error: No hay una sesión activa. Use 'login' primero."
	}

	// Obtener la partición montada por ID
	var mountedPartition Environment.MountedPartition
	var found bool
	for _, partitions := range Environment.GetMountedPartitions() {
		for _, partition := range partitions {
			if partition.MountID == id {
				mountedPartition = partition
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		return fmt.Sprintf("Error: No se encontró la partición con ID %s montada", id)
	}

	// Abrir el archivo del sistema
	file, err := Utils.AbrirArchivo(mountedPartition.MountPath)
	if err != nil {
		return fmt.Sprintf("Error: No se pudo abrir el disco: %v", err)
	}
	defer file.Close()

	// Leer el superbloque
	var superblock Partitions.Superblock
	if err := Utils.LeerArchivo(file, &superblock, int64(mountedPartition.MountStart)); err != nil {
		return fmt.Sprintf("Error: No se pudo leer el superbloque: %v", err)
	}

	// Buscar el archivo
	inodeIndex, _ := User_Groups.InitSearch(pathFileLs, file, superblock)
	if inodeIndex == -1 {
		return fmt.Sprintf("Error: Archivo '%s' no encontrado", pathFileLs)
	}

	// Leer el inodo
	var fileInode Partitions.Inode
	inodePos := superblock.S_inode_start + inodeIndex*int32(binary.Size(Partitions.Inode{}))
	if err := Utils.LeerArchivo(file, &fileInode, int64(inodePos)); err != nil {
		return fmt.Sprintf("Error al leer el inodo del archivo: %v", err)
	}

	// Verificar que sea un archivo regular
	if fileInode.I_type[0] != '1' {
		return "Error: La ruta especificada no es un archivo regular"
	}

	// Leer contenido
	content, _ := User_Groups.GetInodeFileData(fileInode, file, superblock)
	cleanedContent := cleanFileContent(content)

	// Crear directorio de salida si no existe
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Sprintf("Error al crear directorios de salida: %v", err)
	}

	// Escribir el archivo de salida
	if err := os.WriteFile(path, []byte(cleanedContent), 0644); err != nil {
		return fmt.Sprintf("Error al guardar el archivo: %v", err)
	}

	return fmt.Sprintf("Contenido de '%s' guardado en: %s", pathFileLs, path)
}

func cleanFileContent(content string) string {
	// Eliminar caracteres nulos y no imprimibles
	cleaned := strings.Map(func(r rune) rune {
		if r >= 32 && r <= 126 || r == '\n' || r == '\t' {
			return r
		}
		return -1
	}, content)

	// Eliminar líneas vacías y espacios extras
	lines := strings.Split(cleaned, "\n")
	var result strings.Builder

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			result.WriteString(trimmed)
			result.WriteString("\n")
		}
	}

	return strings.TrimSpace(result.String())
}
