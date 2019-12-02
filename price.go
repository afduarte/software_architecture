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
	private.POST("/calculate", calculateCart(s))

	manager := s.router.Group("/manager")
	manager.Use(HydrateUserMiddleware(s))
	manager.Use(RequiresPermissionMiddleware(ManagerRole))
	manager.PUT("/set-price/:ID", setPrice(s))
}

var ProductMap = map[string]*Product{
	"0001": &Product{"0001", "Gadget", 45.50},
	"0002": &Product{"0002", "Widget 1.0", 5.45},
	"0003": &Product{"0003", "Widget 2.0", 7.45},
	"9999": &Product{"9999", "Delivery", 5.0},
}

var DiscountList = []Discount{
	&PercentDiscount{"0001", 20.0},
	&AnyXForY{[]string{"0002", "0003"}, 3, 2},
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
			c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "Message": "price field missing"})
			return
		}
		price, err = strconv.ParseFloat(priceStr, 64)
		if err != nil || price < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "Message": "price value must be a decimal number and bigger than 0"})
			return
		}
		product, ok := ProductMap[id]
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "Message": "product with ID " + id + " not found"})
			return
		}
		product.Price = price

		c.JSON(http.StatusOK, product)
	}
}

type CartValueResponse struct {
	Total           float64
	Discount        float64
	DiscountReasons []string
}

func calculateCart(s *Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		var cart map[string]*ProductOrder
		c.BindJSON(&cart)
		res := CartValueResponse{0, 0, make([]string, 0)}
		// calculate total
		for _, p := range cart {
			// get product in map
			prod, ok := ProductMap[p.ID]
			if !ok {
				c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "Message": "product with ID " + p.ID + " not found"})
				return
			}
			// add price to total if it exists
			res.Total += prod.Price

		}
		// apply any discounts
		for _, d := range DiscountList {
			// apply discounts to cart
			discount, reason := d.Discount(cart, ProductMap)
			if discount != 0.0 && reason != "" {
				res.Discount += discount
				res.DiscountReasons = append(res.DiscountReasons, reason)
			}
		}
		c.JSON(http.StatusOK, res)
	}
}
