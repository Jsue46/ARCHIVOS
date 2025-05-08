package Analyzer

import (
	"Proyecto/User_Groups"
	"flag"
	"os"
	"strings"
)

//═════════════════════   Administración de Usuarios y Grupo ═════════════════════

// Parametro para Iniciar Sesion
func fn_login(input string) string {
	var output strings.Builder
	fs := flag.NewFlagSet("login", flag.ExitOnError)
	user := fs.String("user", "", "Usuario")
	pass := fs.String("pass", "", "Contraseña")
	id := fs.String("id", "", "Id")

	fs.Parse(os.Args[1:])
	matches := paramRegex.FindAllStringSubmatch(input, -1)
	for _, match := range matches {
		flagName := match[1]
		flagValue := match[2]

		flagValue = strings.Trim(flagValue, "\"")

		switch flagName {
		case "user", "pass", "id":
			fs.Set(flagName, flagValue)
		default:
			output.WriteString(" Error: Flag not found ")
		}
	}

	output.WriteString(User_Groups.Login(*user, *pass, *id))
	return output.String()
}

// Parametro para Cerrar Sesion
func fn_logout(_ string) string {
	var output strings.Builder
	output.WriteString(User_Groups.Logout())
	return output.String()
}

// Parametro para Crear un Grupo
func fn_mkgrp(parametros string) string {
	var output strings.Builder
	params := extraerParametros(parametros)
	if name, ok := params["name"]; ok {

		output.WriteString(User_Groups.MKGRP(name))
	} else {
		output.WriteString(" Error: Falta el parámetro -name")
	}
	return output.String()
}

// Parametro para Eliminar un Grupo
func fn_rmgrp(parametros string) string {
	var output strings.Builder
	params := extraerParametros(parametros)
	if name, ok := params["name"]; ok {

		output.WriteString(User_Groups.RMGRP(name))
	} else {
		output.WriteString(" Error: Falta el parámetro -name ")
	}
	return output.String()
}

// Parametro para Crear un Usuario
func fn_mkusr(parametros string) string {
	var output strings.Builder
	paramMap := extraerParametros(parametros)

	// Validar que existan los parámetros necesarios
	user, userOK := paramMap["user"]
	pass, passOK := paramMap["pass"]
	grp, grpOK := paramMap["grp"]

	if !userOK || !passOK || !grpOK {
		return " Error: Faltan parámetros obligatorios (-user, -pass, -grp) "
	}

	output.WriteString(User_Groups.MKUSR(user, pass, grp))
	return output.String()
}

// Parametro para Eliminar un Usuario
func fn_rmusr(parametros string) string {
	var output strings.Builder
	paramMap := extraerParametros(parametros)

	// Validar que exista el parámetro obligatorio
	user, userOK := paramMap["user"]

	if !userOK {
		return " Error: Falta el parámetro obligatorio (-user)"
	}

	output.WriteString(User_Groups.RMUSR(user))
	return output.String()
}

// Funcion para cambiar el grupo de un usuario
func fn_chgrp(parametros string) string {
	var output strings.Builder
	paramMap := extraerParametros(parametros)

	// Validar que existan los parámetros necesarios
	user, userOK := paramMap["user"]
	grp, grpOK := paramMap["grp"]

	if !userOK || !grpOK {
		return "Error: Faltan parámetros obligatorios (-user, -pass, -grp)"
	}

	output.WriteString(User_Groups.CHGRP(user, grp))
	return output.String()
}
