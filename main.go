package main

import (
	"Ejercicio_Final-Taller_Go/api"
	"fmt"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	api.InitRoutes(r)

	if err := r.Run(":8080"); err != nil {
		panic(fmt.Errorf("error trying to start server: %v", err))
	}
}
