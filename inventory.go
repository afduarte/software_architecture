package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func InventoryRoutes(s *Server) {
	private := s.router.Group("/")
	private.Use(HydrateUserMiddleware(s))
	private.GET("/", getInventory(s))
	private.POST("/decrement", decrementStock(s))
}

var InventoryMap = map[string]*InventoryStock{
	"0001": &InventoryStock{"0001", 5, 2},
	"0002": &InventoryStock{"0002", 50, 20},
	"0003": &InventoryStock{"0003", 100, 20},
}

func getInventory(s *Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, InventoryMap)
	}
}

// decrementStock is unsafe because stock should be checked before decrementing
func decrementStock(s *Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		var decrements map[string]*ProductOrder
		c.BindJSON(&decrements)
		for _, d := range decrements {
			inv, ok := InventoryMap[d.ID]
			if !ok {
				continue
			}
			inv.Quantity -= d.Quantity
		}
		c.JSON(http.StatusOK, InventoryMap)
	}
}
