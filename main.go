package main

import (
	"Proyecto/Analyzer"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type CommandRequest struct {
	Command string `json:"command"`
}

func executeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var request CommandRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Error al decodificar JSON", http.StatusBadRequest)
		return
	}

	var results []string
	lines := strings.Split(request.Command, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		comando, parametros := Analyzer.GetEntrada(line)
		result := Analyzer.AnalizadorComandos(comando, parametros)
		// Limpia cualquier carácter especial que pueda interferir con JSON
		result = strings.ReplaceAll(result, "\r", "")
		results = append(results, result)
	}

	// Usar json.Marshal en lugar de json.NewEncoder para tener más control
	response := map[string]string{"output": strings.Join(results, "\n")}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		fmt.Println("Error al codificar JSON:", err)
		http.Error(w, "Error al codificar JSON", http.StatusInternalServerError)
		return
	}

	fmt.Println("Respuesta JSON:", string(jsonResponse))

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

/*
	func executeHandler(w http.ResponseWriter, r *http.Request) {
		var request CommandRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http.Error(w, "Error al decodificar JSON", http.StatusBadRequest)
			return
		}

		// Respuesta simple para pruebas
		response := map[string]string{"output": "Comando recibido: " + request.Command}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
*/
func main() {
	router := mux.NewRouter()
	router.HandleFunc("/execute", executeHandler).Methods("POST")

	cors := handlers.CORS(
		handlers.AllowedOrigins([]string{"http://localhost:3000"}), // Específica el origen de tu frontend
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
		handlers.ExposedHeaders([]string{"Content-Length"}),
		handlers.AllowCredentials(),
	)

	fmt.Println("SERVIDOR CORRIENDO EN http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", cors(router)))
}
