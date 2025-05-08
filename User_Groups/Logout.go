package User_Groups

import (
	"Proyecto/Environment"
	"strings"
)

// Función para cerrar sesión
func Logout() string {
	var output strings.Builder
	output.WriteString("╔═════════════════════ CERRAR  SESION ═════════════════════════╗\n")
	// Obtener las particiones montadas
	mountedPartitions := Environment.GetMountedPartitions()
	// Variable para verificar si hay una sesión activa
	var sessionFound bool = false

	// Buscar la partición que tiene una sesión activa
	for _, partitions := range mountedPartitions {
		for _, partition := range partitions {
			if partition.LoggedIn {
				// Marcar la partición como deslogueada
				output.WriteString(Environment.ParticionSinInicioSesion(partition.MountID))
				sessionFound = true
				break
			}
		}
		if sessionFound {
			break
		}
	}

	// Si no se encontró una sesión activa, mostrar un error
	if !sessionFound {
		output.WriteString("  Error: No hay una Sesión Activa para Cerrar.\n")
	} else {
		output.WriteString("\t Sesión Finalizada  con éxito.\n")
	}

	output.WriteString("╚═══════════════════   CERRANDO SESION  ═══════════════════════╝ \n")
	return output.String()
}
