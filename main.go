package main

import (
	"flag"
	"net/http"

	"github.com/gin-gonic/gin"
)

var service = flag.String("s", "order", "The type of service to run, can be one of [order, inventory, price, loyalty, auth]")

func main() {
	flag.Parse()
	println(service)
	s := &Server{
		router:  gin.Default(),
		service: *service,
	}
	s.routes()
	s.router.Run() // listen and serve on 0.0.0.0:8080
}

func (s *Server) routes() {
	switch s.service {
	case "order":
		OrderRoutes(s)
	case "inventory":
		InventoryRoutes(s)
	case "price":
		PriceRoutes(s)
	case "loyalty":
		LoyaltyRoutes(s)
	case "auth":
		AuthRoutes(s)
	default:
		panic("type " + s.service + " is not allowed, allowed types: [order, inventory, price, loyalty, auth]")
	}
}

func noOp(s *Server) func(*gin.Context) {
	return func(c *gin.Context) {

		c.String(http.StatusOK, "OK")
	}
}
