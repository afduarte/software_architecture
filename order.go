package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func OrderRoutes(s *Server) {
	private := s.router.Group("/")
	private.Use(HydrateUserMiddleware(s))
	private.POST("/new", buyOrder(s))
}

type BuyOrderResponse struct {
	Order    *Order
	Message  string
	Warnings []string
	Errors   []string `json:",omitempty"`
}

type BuyOrderRequest struct {
	Cart            map[string]*ProductOrder
	CustomerID      string
	UsePoints       int
	DeliveryAddress string
}

var OrdersMap = make(map[string]*Order)

func buyOrder(s *Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := c.MustGet("user").(*User)
		errors := make([]string, 0)
		warnings := make([]string, 0)
		stock := FetchInventory(s.config.inventoryEndpoint, user.Token)
		if stock == nil {
			errors := append(errors, "unable to reach inventory server")
			c.JSON(http.StatusServiceUnavailable, BuyOrderResponse{nil, "unable to fulfill order", warnings, errors})
			return
		}
		var orderReq BuyOrderRequest
		c.BindJSON(&orderReq)
		for _, p := range orderReq.Cart {
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
			c.JSON(http.StatusBadRequest, BuyOrderResponse{nil, "unable to fulfill order", warnings, errors})
			return
		}
		err := SendDecrementRequest(s.config.inventoryEndpoint, user.Token, orderReq.Cart)
		if err != nil {
			errors := append(errors, err.Error())
			c.JSON(http.StatusServiceUnavailable, BuyOrderResponse{nil, "unable to fulfill order", warnings, errors})
			return
		}

		if orderReq.DeliveryAddress != "" {
			orderReq.Cart["9999"] = &ProductOrder{"9999", 1}
		}

		cartResp, cartErr := SendCalculateCartRequest(s.config.priceEndpoint, user.Token, orderReq.Cart)
		if cartErr != nil {
			errors := append(errors, cartErr.Error())
			c.JSON(http.StatusServiceUnavailable, BuyOrderResponse{nil, "unable to fulfill order", warnings, errors})
			return
		}

		loyaltyResp, loyaltyErr := SendUpdatePointsRequest(s.config.loyaltyEndpoint, user.Token, orderReq.CustomerID, orderReq.Cart, orderReq.UsePoints)
		if loyaltyErr != nil {
			errors := append(errors, loyaltyErr.Error())
			c.JSON(http.StatusServiceUnavailable, BuyOrderResponse{nil, "unable to fulfill order", warnings, errors})
			return
		}

		if loyaltyResp.Discount > 0 {
			cartResp.Discount += loyaltyResp.Discount
			cartResp.DiscountReasons = append(cartResp.DiscountReasons, fmt.Sprintf("%.2f off for using %d loyalty points", loyaltyResp.Discount, orderReq.UsePoints))
		}

		id := uuid.Must(uuid.NewRandom())
		order := &Order{id.String(), user.Name, orderReq.CustomerID, orderReq.DeliveryAddress, "processed", time.Now(), orderReq.Cart, cartResp.Total, cartResp.Discount, cartResp.DiscountReasons}
		OrdersMap[id.String()] = order

		c.JSON(http.StatusOK, BuyOrderResponse{order, "order processed successfully", warnings, errors})
	}
}
