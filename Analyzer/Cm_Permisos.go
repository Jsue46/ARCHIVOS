package Analyzer

import (
	"Proyecto/FPermissions"
	"strings"
)

// Archvivos Y permisos
func fn_mkfile(parametros string) string {
	var output strings.Builder
	// Extraer los parámetros en formato map[string]string
	paramsMap := extraerParametros(parametros)

	// Pasar los parámetros a la función Mkfile
	output.WriteString(FPermissions.Mkfile(paramsMap))
	return output.String()
}

// Archivos y permisos
func fn_mkdir(parametros string) string {
	var output strings.Builder
	paramsMap := extraerParametros(parametros)

	//Pasar los parámetros a la función Mkdir
	output.WriteString(FPermissions.Mkdir(paramsMap))
	return output.String()
}
