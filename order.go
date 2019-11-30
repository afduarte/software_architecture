package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func OrderRoutes(s *Server) {
	private := s.router.Group("/")
	private.Use(HydrateUserMiddleware(s))
	private.POST("/", buyOrder(s))
}

type BuyOrderResponse struct {
	Message  string
	Warnings []string
	Errors   []string
}

var OrdersMap = make(map[string]*Order)

func buyOrder(s *Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := c.MustGet("user").(*User)
		stock := FetchInventory(s.config.inventoryEndpoint, user.Token)
		var cart []ProductOrder
		c.BindJSON(&cart)
		errors := make([]string, 0)
		warnings := make([]string, 0)
		for _, p := range cart {
			s, ok := stock[p.ID]
			if !ok {
				errors = append(errors, "product with ID: "+p.ID+" not found")
				continue
			}
			if p.Quantity > s.Quantity {
				errors = append(errors, "insufficient stock to fulfill order for product with ID: "+p.ID)
				continue
			}
			if (s.Quantity - p.Quantity) <= s.LowWarning {
				warnings = append(warnings, "order will take stock for product with ID "+p.ID+" below the warning threshold")
				continue
			}
		}
		if len(errors) > 0 {
			c.JSON(http.StatusBadRequest, BuyOrderResponse{"unable to fulfill order", warnings, errors})
			return
		}
		err := SendDecrementRequest(s.config.inventoryEndpoint, user.Token, &cart)
		if err != nil {
			errors := append(errors, err.Error())
			c.JSON(http.StatusBadRequest, BuyOrderResponse{"unable to fulfill order", warnings, errors})
			return
		}
		id := uuid.Must(uuid.NewRandom())
		discounts := make([]Discount, 0)
		OrdersMap[id.String()] = &Order{id.String(), cart, discounts}
	}
}
