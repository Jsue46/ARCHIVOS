package User_Groups

import (
	"Proyecto/Environment"
	"Proyecto/Partitions"
	"Proyecto/Utils"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
)

type User struct {
	ID          int
	GID         int
	Name        string
	Pass        string
	Group       string
	PartitionID string
}

// Variable para almacenar el usuario actual
var loggedInUser User
var isLoggedIn bool

// Función para verificar si el usuario actual es root
func IsRootLoggedIn() bool {
	return isLoggedIn && loggedInUser.Name == "root"
}

// Función para establecer el usuario actual
func SetCurrentUser(user User) {
	loggedInUser = user
	isLoggedIn = true
}

// Función para limpiar el usuario actual
func ClearCurrentUser() {
	loggedInUser = User{}
	isLoggedIn = false
}

// Función para obtener el usuario actual
func GetCurrentUser() User {
	return loggedInUser
}

// Función para verificar si hay un usuario logueado
func IsUserLoggedIn() bool {
	return isLoggedIn
}

// Funcion para Iniciar Sesión
func Login(user string, pass string, id string) string {
	var output strings.Builder
	output.WriteString("╔═════════════════════ INICIAR SESION ═════════════════════════╗\n")
	output.WriteString(fmt.Sprintf("  USUARIO: %s\n", user))
	output.WriteString(fmt.Sprintf("  CONTRASEÑA: %s\n", pass))
	output.WriteString(fmt.Sprintf("  ID: %s\n", id))

	// Verificar si ya hay una sesión activa globalmente
	if isLoggedIn {
		return "       ⚠️ Error: Ya hay Usuario Activo   "
	}

	mountedPartitions := Environment.GetMountedPartitions()
	var filepath string
	var partitionFound bool
	var partitionID string

	// Buscar la partición montada
	for _, partitions := range mountedPartitions {
		for _, partition := range partitions {
			if partition.MountID == id {
				filepath = partition.MountPath
				partitionID = partition.MountID // Guardar el PartitionID
				partitionFound = true
				break
			}
		}
		if partitionFound {
			break
		}
	}

	if !partitionFound {
		return " ⚠️ Error: Partición No Fue Encontrada   "
	}

	// Abrir el archivo del sistema de archivos binario
	file, err := Utils.AbrirArchivo(filepath)
	if err != nil {
		return fmt.Sprintf(" ⚠️ Error: No se pudo Abrir el Archivo: %s\n", err)
	}
	defer file.Close()

	// Leer el MBR
	var TempMBR Partitions.MBR
	if err := Utils.LeerArchivo(file, &TempMBR, 0); err != nil {
		return fmt.Sprintf(" ⚠️ Error: No se pudo Leer el MBR: %s    \n", err)
	}

	// Buscar la partición dentro del MBR
	var index int = -1
	for i := 0; i < 4; i++ {
		if TempMBR.MbrPartitions[i].PartSize != 0 {
			if strings.Contains(string(TempMBR.MbrPartitions[i].PartID[:]), id) {
				if TempMBR.MbrPartitions[i].PartStatus[0] == '1' {
					index = i
				} else {
					return " ⚠️ Error: La Partición No Está Montada  "
				}
				break
			}
		}
	}

	if index == -1 {
		return " ⚠️ Error: No se Encontró La Partición    "
	}

	// Leer el Superblock
	var tempSuperblock Partitions.Superblock
	if err := Utils.LeerArchivo(file, &tempSuperblock, int64(TempMBR.MbrPartitions[index].PartStart)); err != nil {
		return fmt.Sprintf(" ⚠️ Error: No se Pudo Leer el SuperBlock: %s \n", err)
	}

	// Buscar el archivo users.txt
	indexInode, log := InitSearch("/users.txt", file, tempSuperblock)
	output.WriteString(log) // Agregar el log de InitSearch al log principal
	if indexInode == -1 {
		return " ⚠️ Error: Archivo '/users.txt' No Encontrado  "
	}

	// Leer el Inodo del archivo "users.txt"
	var crrInode Partitions.Inode
	if err := Utils.LeerArchivo(file, &crrInode, int64(tempSuperblock.S_inode_start+indexInode*int32(binary.Size(Partitions.Inode{})))); err != nil {
		return " ⚠️ Error: No se Pudo Leer el Inodo      "
	}

	// Obtener el contenido del archivo users.txt
	data, log := GetInodeFileData(crrInode, file, tempSuperblock)
	output.WriteString(log) // Agregar el log de GetInodeFileData al log principal
	lines := strings.Split(data, "\n")

	// Verificar credenciales
	for _, line := range lines {
		words := strings.Split(line, ",")
		if len(words) == 5 && words[0] != "0" { // Ignorar usuarios eliminados
			if words[3] == user && words[4] == pass {
				// Crear estructura User con datos reales
				newUser := User{
					ID:          atoi(words[0]),
					GID:         atoi(words[0]),
					Name:        words[3],
					Pass:        words[4],
					Group:       words[2],
					PartitionID: partitionID,
				}

				SetCurrentUser(newUser)
				Environment.ParticionConInicioSesion(id)

				output.WriteString(fmt.Sprintf(" ═══════════  Usuario '%s' Logueado con Exito  ═════════════════ \n", user))
				return output.String()
			}
		}
	}

	output.WriteString(" ⚠️ Error: Usuario o Contraseña Incorrectos \n")
	return output.String()
}

// Función auxiliar para convertir string a int
func atoi(s string) int {
	i, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0
	}
	return i
}
