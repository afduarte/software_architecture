package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func PriceRoutes(s *Server) {
	private := s.router.Group("/")
	private.Use(HydrateUserMiddleware(s))
	private.GET("/", getProducts(s))

	manager := s.router.Group("/manager")
	manager.Use(HydrateUserMiddleware(s))
	manager.Use(RequiresPermissionMiddleware(ManagerRole))
	manager.PUT("/set-price/:ID", setPrice(s))
}

var ProductMap = map[string]*Product{
	"0001": &Product{"0001", "Gadget", 45.50},
	"0002": &Product{"0002", "Widget", 5.45},
}

func getProducts(s *Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, ProductMap)
	}
}

func setPrice(s *Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error
		var priceStr string
		var price float64
		id := c.Param("ID")
		if priceStr = c.PostForm("price"); priceStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "message": "price field missing"})
			return
		}
		price, err = strconv.ParseFloat(priceStr, 64)
		if err != nil || price < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "message": "price value must be a decimal number and bigger than 0"})
			return
		}
		product, ok := ProductMap[id]
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "message": "product with ID " + id + " not found"})
			return
		}
		product.Price = price

		c.JSON(http.StatusOK, product)
	}
}
