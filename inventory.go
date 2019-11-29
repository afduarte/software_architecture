package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func InventoryRoutes(s *Server) {
	private := s.router.Group("/")
	private.Use(HydrateUserMiddleware(s))
	private.GET("/inventory", getInventory(s))

	manager := s.router.Group("/manager")
	manager.Use(HydrateUserMiddleware(s))
	manager.Use(RequiresPermissionMiddleware(ManagerRole))
	manager.GET("/inventory", getInventory(s))
}

var InventoryMap = map[string]*InventoryStock{
	"0001": &InventoryStock{"0001", 5, 2},
}

func getInventory(s *Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, InventoryMap)
	}
}
