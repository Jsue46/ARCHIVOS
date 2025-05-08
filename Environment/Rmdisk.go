package Environment

import (
	"fmt"
	"os"
	"strings"
)

// Rmdisk elimina un archivo de disco en la ruta especificada
func Rmdisk(path string) string {
	var output strings.Builder
	output.WriteString("╔═════════════════════ INICIO RMDISK  ═════════════════════════╗\n")

	// Verificar si el archivo existe
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Sprintf("  Error: El archivo no existe en la ruta especificada: %s\n", path)
	}

	// Intentar eliminar el archivo
	err := os.Remove(path)
	if err != nil {
		return fmt.Sprintf("  Error: No se pudo eliminar el disco en la ruta %s: %v\n", path, err)
	}

	// Mensajes de éxito
	output.WriteString(fmt.Sprintf("  Disco en la ruta %s\n", path))
	output.WriteString("    Ha sido eliminado correctamente.\n")
	output.WriteString("╚═════════════════════   FIN RMDISK   ═════════════════════════╝")

	return output.String()
}
