package main

import (
	"fmt"
	"log"
	"os"

	"parte3/api"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Definir la URL base de la API de usuarios.
	// Se asume que tu API de usuarios corre en http://localhost:8080
	userAPIURL := os.Getenv("USER_API_URL")
	if userAPIURL == "" {
		userAPIURL = "http://localhost:8080" // URL por defecto para la API de usuarios
		log.Printf("Warning: USER_API_URL not set, using default: %s", userAPIURL)
	}

	api.InitRoutes(r, userAPIURL) // Pasar la userAPIURL

	port := os.Getenv("SALES_API_PORT")
	if port == "" {
		port = "8080" // Puerto por defecto para la API de ventas
	}
	addr := fmt.Sprintf(":%s", port)

	log.Printf("Starting sales API server at %s", addr)
	if err := r.Run(addr); err != nil {
		panic(fmt.Errorf("error trying to start sales API server: %v", err))
	}
}
