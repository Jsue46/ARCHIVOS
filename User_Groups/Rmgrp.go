package User_Groups

import (
	"Proyecto/Environment"
	"Proyecto/Partitions"
	"Proyecto/Utils"
	"encoding/binary"
	"fmt"
	"strings"
)

func RMGRP(name string) string {
	var output strings.Builder
	output.WriteString("╔══════════════════════ INICIO RMGRP  ═════════════════════════╗ \n")
	output.WriteString(fmt.Sprintf("  Nombre del grupo a eliminar: %s\n", name))

	// Verificar si el usuario actual es root
	if !IsRootLoggedIn() {
		return "❌ Error: Solo el usuario root puede eliminar grupos."
	}

	// Obtener la partición montada actualmente
	mountedPartition := Environment.GetCurrentMountedPartition()
	if mountedPartition == nil {
		return "❌ Error: No hay ninguna partición montada."
	}

	// Abrir el archivo del sistema de archivos
	file, err := Utils.AbrirArchivo(mountedPartition.MountPath)
	if err != nil {
		return fmt.Sprintf("❌ Error al abrir el archivo: %v", err)
	}
	defer file.Close()

	// Leer el Superblock de la partición
	var tempSuperblock Partitions.Superblock
	if err := Utils.LeerArchivo(file, &tempSuperblock, int64(mountedPartition.MountStart)); err != nil {
		return fmt.Sprintf("❌ Error al leer el Superblock: %v", err)
	}

	// Buscar el archivo "users.txt"
	indexInode, log := InitSearch("/users.txt", file, tempSuperblock)
	output.WriteString(log) // Agregar el log de InitSearch al log principal
	if indexInode == -1 {
		return "❌ Error: No se encontró el archivo users.txt."
	}

	// Leer el inodo del archivo "users.txt"
	var crrInode Partitions.Inode
	if err := Utils.LeerArchivo(file, &crrInode, int64(tempSuperblock.S_inode_start+indexInode*int32(binary.Size(Partitions.Inode{})))); err != nil {
		return fmt.Sprintf("❌ Error al leer el inodo de users.txt: %v", err)
	}

	blockSize := binary.Size(Partitions.Fileblock{})
	foundGroup := false

	// Recorrer todos los bloques directos del inodo
	for i := 0; i < 12; i++ {
		block := crrInode.I_block[i]
		if block == -1 {
			continue
		}

		var fileBlock Partitions.Fileblock
		blockOffset := int64(tempSuperblock.S_block_start + block*int32(blockSize))

		if err := Utils.LeerArchivo(file, &fileBlock, blockOffset); err != nil {
			return fmt.Sprintf("❌ Error al leer bloque de archivo: %v", err)
		}

		// Imprimir contenido completo para depuración
		output.WriteString(" ═══════════════  CONTENIDO DEL BLOQUE   ══════════════════════ \n")
		content := string(fileBlock.B_content[:])
		output.WriteString(content + "\n")
		output.WriteString(" ═══════════════  FIN CONTENIDO DEL BLOQUE   ══════════════════════ \n")

		// Depuración: contar y mostrar todos los grupos encontrados en este bloque
		groupCount := 0
		for _, line := range strings.Split(content, "\n") {
			if strings.Contains(line, ",G,") {
				output.WriteString(fmt.Sprintf("Grupo encontrado en bloque: %s\n", line))
				groupCount++
			}
		}
		output.WriteString(fmt.Sprintf("Total de grupos encontrados en este bloque: %d\n", groupCount))

		// Normalizar el contenido para manejar diferentes formatos
		content = strings.ReplaceAll(content, ", ", ",")
		content = strings.ReplaceAll(content, " ,", ",")

		// Buscar el grupo por nombre usando un enfoque simple pero efectivo
		searchPattern := fmt.Sprintf("G,%s", name)
		if strings.Contains(content, searchPattern) {
			output.WriteString(fmt.Sprintf("✅ Patrón '%s' encontrado en el bloque. Procesando...\n", searchPattern))

			// Procesar el contenido línea por línea para encontrar y modificar la línea específica
			modified := false
			lines := strings.Split(content, "\n")

			for j, line := range lines {
				// Normalizar la línea
				line = strings.ReplaceAll(line, ", ", ",")
				line = strings.ReplaceAll(line, " ,", ",")

				if strings.Contains(line, searchPattern) {
					output.WriteString(fmt.Sprintf("✅ Línea encontrada: '%s'\n", line))
					// Extraer el ID del grupo
					parts := strings.Split(strings.TrimSpace(line), ",")
					if len(parts) >= 3 {
						// Marcar como eliminado cambiando el ID a 0
						lines[j] = "0,G," + name
						modified = true
						foundGroup = true
					}
				}
			}

			if modified {
				// Reconstruir el contenido del bloque
				newContent := strings.Join(lines, "\n")
				fileBlock.B_content = [64]byte{}
				copy(fileBlock.B_content[:], newContent)

				// Escribir el bloque actualizado en el archivo
				if err := Utils.EscribirArchivo(file, fileBlock, blockOffset); err != nil {
					return fmt.Sprintf("❌ Error al escribir el bloque actualizado: %v", err)
				}
			}
		} else {
			// Intento alternativo usando otra forma de buscar
			lines := strings.Split(content, "\n")
			modified := false

			for j, line := range lines {
				// Quitar espacios en blanco al inicio y final de la línea
				trimmedLine := strings.TrimSpace(line)
				if trimmedLine == "" {
					continue
				}

				parts := strings.Split(trimmedLine, ",")
				// Imprimir información de depuración para cada línea analizada
				if len(parts) >= 3 {
					output.WriteString(fmt.Sprintf("Analizando: ID=%s, Tipo=%s, Nombre=%s\n",
						strings.TrimSpace(parts[0]),
						strings.TrimSpace(parts[1]),
						strings.TrimSpace(parts[2])))

					// Verificar si es un grupo y coincide con el nombre buscado
					// Más flexible con espacios y formato
					if strings.Contains(strings.ToLower(strings.TrimSpace(parts[1])), "g") &&
						strings.EqualFold(strings.TrimSpace(parts[2]), name) {
						output.WriteString("✅ Grupo encontrado, marcando como eliminado en su bloque.\n")
						// Marcar como eliminado cambiando el ID a 0
						lines[j] = "0,G," + name
						modified = true
						foundGroup = true
					}
				}
			}

			if modified {
				// Reconstruir el contenido del bloque
				newContent := strings.Join(lines, "\n")
				fileBlock.B_content = [64]byte{}
				copy(fileBlock.B_content[:], newContent)

				// Escribir el bloque actualizado en el archivo
				if err := Utils.EscribirArchivo(file, fileBlock, blockOffset); err != nil {
					return fmt.Sprintf("❌ Error al escribir el bloque actualizado: %v", err)
				}
			}
		}
	}

	if foundGroup {
		output.WriteString("✅ Grupo eliminado correctamente. \n")
	} else {
		output.WriteString("❌ Error: El grupo no existe en el sistema.\n")
	}

	output.WriteString("╚══════════════════════   FIN RMGRP   ═════════════════════════╝ \n")
	return output.String()
}
