package main

import "github.com/gin-gonic/gin"

type Server struct {
	router  *gin.Engine
	service string
}

type Order struct {
	ID       string
	Products []Product
	Discount []Discount
}

type Product struct {
	ID       string
	Quantity int
	Price    float32
}

type Discount struct {
	Amount float32
	Reason string
}
