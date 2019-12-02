package main

import (
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
)

func LoyaltyRoutes(s *Server) {
	private := s.router.Group("/")
	private.Use(HydrateUserMiddleware(s))
	private.POST("/update-points", updatePoints(s))
	private.GET("/points/:cID", pointsForCustomer(s))
}

const buyPointsPerPound = 1
const discountPointsPerPound = 100

var CustomerMap = map[string]*Customer{
	"000001": &Customer{"000001", 0},
	"000002": &Customer{"000002", 1000},
}

var ProductPointsMultiplier = map[string]float64{
	"0001": 2.0,
	"0003": 1.5,
}

type GetPointsResponse struct {
	Customer       *Customer
	PointsPerPound int
}

func pointsForCustomer(s *Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		cID := c.Param("ID")
		if cID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"Message": "missing customer field"})
			return
		}
		customer, ok := CustomerMap[cID]
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"Message": "customer with id: " + cID + " not found"})
			return
		}
		c.JSON(http.StatusOK, GetPointsResponse{customer, discountPointsPerPound})
	}
}

type UpdatePointsRequest struct {
	CustomerID          string
	Cart                map[string]*ProductOrder
	ApplyDiscountPoints int
}

type UpdatePointsResponse struct {
	CustomerID        string
	PointsBeforeOrder int
	PointsAfterOrder  int
	Discount          float64
}

func updatePoints(s *Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req UpdatePointsRequest
		c.BindJSON(&req)
		customer, ok := CustomerMap[req.CustomerID]
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"Message": "customer with id: " + req.CustomerID + " not found"})
			return
		}
		user := c.MustGet("user").(*User)
		prices := FetchProductPrices(s.config.priceEndpoint, user.Token)
		if prices == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"Message": "unable to reach price server"})
			return
		}
		resp := UpdatePointsResponse{req.CustomerID, customer.Points, customer.Points, 0}
		for _, p := range req.Cart {
			prod, ok := prices[p.ID]
			if !ok {
				continue
			}
			mult := 1.0
			// check if there's a multiplier on for the current product, use it if so
			if m, ok := ProductPointsMultiplier[p.ID]; ok {
				mult = m
			}
			// We only count full points, so half points get dropped.
			// Formula is: QTY x Multiplier x PointsPerPoundWhenBuying x ItemPrice
			floatingPoints := float64(p.Quantity) * mult * float64(buyPointsPerPound) * prod.Price
			// math.Trunc() will do that to a float, int() converts to int type
			resp.PointsAfterOrder += int(math.Trunc(floatingPoints))
		}
		if req.ApplyDiscountPoints > 0 {
			if req.ApplyDiscountPoints > customer.Points {
				c.JSON(http.StatusBadRequest, gin.H{"Message": "customer does not have enough points to fulfill request"})
				return
			}
			resp.Discount = float64(req.ApplyDiscountPoints) / discountPointsPerPound
			resp.PointsAfterOrder -= req.ApplyDiscountPoints
		}
		// save state to customer points
		customer.Points = resp.PointsAfterOrder
		c.JSON(http.StatusOK, resp)
	}
}
