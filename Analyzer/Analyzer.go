package Analyzer

import (
	"regexp"
	"strings"
)

var paramRegex = regexp.MustCompile(`-(\w+)=("[^"]+"|\S+)`)

func GetEntrada(input string) (string, string) {
	parts := strings.Fields(input)
	if len(parts) > 0 {
		comandos := strings.ToLower(parts[0])
		parametros := strings.Join(parts[1:], " ")
		return comandos, parametros
	}
	return "", input
}

func extraerParametros(parametros string) map[string]string {
	matches := paramRegex.FindAllStringSubmatch(parametros, -1)
	paramMap := make(map[string]string)

	for _, match := range matches {
		flagName := strings.ToLower(match[1])
		flagValue := strings.Trim(match[2], "\"")
		if flagName == "path" {
			// Convertir el valor del path a minúsculas
			flagValue = strings.ToLower(flagValue)
		}
		paramMap[flagName] = flagValue
	}

	return paramMap
}

func AnalizadorComandos(comandos string, parametros string) string {
	extraerParametros(parametros)

	// Agregar el comando ejecutado al resultado
	result := "> " + comandos + " " + parametros + "\n"

	switch {
	case strings.Contains(comandos, "mkdisk"):
		return result + fn_mkdisk(parametros)
	case strings.Contains(comandos, "rmdisk"):
		return result + fn_rmdisk(parametros)
	case strings.Contains(comandos, "fdisk"):
		return result + fn_fdisk(parametros)
	case strings.Contains(comandos, "mounted"):
		return result + fn_mounted(parametros)
	case strings.Contains(comandos, "mount"):
		return result + fn_mount(parametros)
	case strings.Contains(comandos, "mkfs"):
		return result + fn_mkfs(parametros)
	case strings.Contains(comandos, "cat"):
		return result + fn_cat(parametros)
	case strings.Contains(comandos, "login"):
		return result + fn_login(parametros)
	case strings.Contains(comandos, "logout"):
		return result + fn_logout(parametros)
	case strings.Contains(comandos, "mkgrp"):
		return result + fn_mkgrp(parametros)
	case strings.Contains(comandos, "rmgrp"):
		return result + fn_rmgrp(parametros)
	case strings.Contains(comandos, "mkusr"):
		return result + fn_mkusr(parametros)
	case strings.Contains(comandos, "rmusr"):
		return result + fn_rmusr(parametros)
	case strings.Contains(comandos, "chgrp"):
		return result + fn_chgrp(parametros)
	case strings.Contains(comandos, "mkfile"):
		return result + fn_mkfile(parametros)
	case strings.Contains(comandos, "mkdir"):
		return result + fn_mkdir(parametros)
	case strings.Contains(comandos, "rep"): //Aun Faltan Agregar ARBOL,LS,FILE
		return result + fn_rep(parametros)
	default:
		return result + "Error: Comando inválido o no encontrado"
	}
}
