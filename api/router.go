package api

import (
	"net/http"
	"parte3/internal/sales"

	//"Ejercicio_Final-Taller_Go/internal/sales"
	"Ejercicio_Final-Taller_Go/internal/user"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// InitRoutes registers all user and sales CRUD endpoints on the given Gin engine.
// It initializes the storage, service, and handler for both users and sales,
// then binds each HTTP method and path to the appropriate handler function.
func InitRoutes(e *gin.Engine, userAPIURL string) { // Modificamos la firma para recibir userAPIURL
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Inicialización de la lógica de usuarios (sin cambios)
	userStorage := user.NewLocalStorage()
	userService := user.NewService(userStorage, logger)
	userHandler := handler{
		userService: userService,
		logger:      logger,
	}

	e.POST("/users", userHandler.handleCreate)
	e.GET("/users/:id", userHandler.handleRead)
	e.PATCH("/users/:id", userHandler.handleUpdate)
	e.DELETE("/users/:id", userHandler.handleDelete)

	e.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// Inicialización de la lógica de ventas
	salesStorage := sales.NewLocalStorage()
	salesService := sales.NewService(salesStorage, logger, userAPIURL) // Usamos la userAPIURL recibida
	salesHandler := NewSalesHandler(salesService, logger)

	e.POST("/sales", salesHandler.handleCreateSale)
	// Ruta para actualizar el estado de una venta
	e.PATCH("/sales/:id", salesHandler.PatchSaleHandler(salesService))
	e.GET("/sales", salesHandler.handlerGetSale)

}
