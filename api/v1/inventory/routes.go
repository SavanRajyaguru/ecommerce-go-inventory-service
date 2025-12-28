package inventory

import (
	"github.com/gin-gonic/gin"
	"github.com/savanrajyaguru/ecommerce-go-inventory-service/internal/auth"
)

func RegisterRoutes(router *gin.RouterGroup, h *InventoryHandler) {
	inv := router.Group("/inventory")
	inv.Use(auth.AuthMiddleware())
	{
		// Admin Write
		inv.POST("", auth.RoleMiddleware(auth.RoleAdmin), h.CreateInventory)
		inv.PUT("/:productId", auth.RoleMiddleware(auth.RoleAdmin), h.UpdateStock)

		// Public/User Read
		inv.GET("/:productId", auth.RoleMiddleware(auth.RoleAdmin, auth.RoleUser), h.GetInventory)
		inv.GET("/category/:category", auth.RoleMiddleware(auth.RoleAdmin, auth.RoleUser), h.ListByCategory)
	}
}
