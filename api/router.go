package api

import (
	"github.com/gin-gonic/gin"
	"github.com/savanrajyaguru/ecommerce-go-inventory-service/api/v1/inventory"
	"github.com/savanrajyaguru/ecommerce-go-inventory-service/api/v1/middleware"
	"github.com/savanrajyaguru/ecommerce-go-inventory-service/repository"
	"github.com/savanrajyaguru/ecommerce-go-inventory-service/services"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

	repo := repository.NewInventoryRepository()
	service := services.NewInventoryService(repo)
	handler := inventory.NewInventoryHandler(service)

	v1 := r.Group("/v1")
	{
		inventory.RegisterRoutes(v1, handler)
	}

	return r
}
