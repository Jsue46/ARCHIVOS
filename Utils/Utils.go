package Utils

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
)

// CrearArchivo crea un archivo vacío en la ruta especificada.
func CrearArchivo(name string) error {
	// Asegura que el directorio exista
	dir := filepath.Dir(name)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("error al crear el directorio: %v", err)
	}

	// Crea el archivo si no existe
	if _, err := os.Stat(name); os.IsNotExist(err) {
		file, err := os.Create(name)
		if err != nil {
			return fmt.Errorf("error al crear el archivo: %v", err)
		}
		defer file.Close()
	}

	return nil
}

// AbrirArchivo abre un archivo en modo lectura y escritura.
func AbrirArchivo(name string) (*os.File, error) {
	file, err := os.OpenFile(name, os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("error al abrir el archivo: %v", err)
	}
	return file, nil
}

// EscribirArchivo escribe un objeto en un archivo binario en la posición especificada.
func EscribirArchivo(file *os.File, data interface{}, position int64) error {
	if _, err := file.Seek(position, 0); err != nil {
		return fmt.Errorf("error al buscar la posición en el archivo: %v", err)
	}

	if err := binary.Write(file, binary.LittleEndian, data); err != nil {
		return fmt.Errorf("error al escribir el objeto en el archivo: %v", err)
	}

	return nil
}

// LeerArchivo lee un objeto de un archivo binario en la posición especificada.
func LeerArchivo(file *os.File, data interface{}, position int64) error {
	if _, err := file.Seek(position, 0); err != nil {
		return fmt.Errorf("error al buscar la posición en el archivo: %v", err)
	}

	if err := binary.Read(file, binary.LittleEndian, data); err != nil {
		return fmt.Errorf("error al leer el objeto del archivo: %v", err)
	}

	return nil
}
