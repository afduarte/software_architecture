package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/gin-gonic/gin"
)

type Server struct {
	router  *gin.Engine
	service string
	config  *Config
}

type Config struct {
	authEndpoint      string
	inventoryEndpoint string
	loyaltyEndpoint   string
	orderEndpoint     string
	priceEndpoint     string
}

// PermissionRole represents the permission level of a user
type PermissionRole int

const (
	// UserRole < ManagerRole, it's also the 0 value, so any User created without a Role gets User
	UserRole PermissionRole = iota
	// ManagerRole > UserRole
	ManagerRole
)

func (s PermissionRole) String() string {
	return roleToString[s]
}

var roleToString = map[PermissionRole]string{
	UserRole:    "UserRole",
	ManagerRole: "ManagerRole",
}

var roleToID = map[string]PermissionRole{
	"UserRole":    UserRole,
	"ManagerRole": ManagerRole,
}

// MarshalJSON marshals the enum as a quoted json string
func (s PermissionRole) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(roleToString[s])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON unmashals a quoted json string to the enum value
func (s *PermissionRole) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	// Note that if the string cannot be found then it will be set to the zero value, 'UserRole' in this case.
	*s = roleToID[j]
	return nil
}

type User struct {
	Username string
	password string
	Name     string
	Token    string `json:",omitempty"`
	Role     PermissionRole
}

type Customer struct {
	ID     string
	Points int
}

type Order struct {
	ID              string
	UserID          string
	CustomerID      string
	DeliveryAddress string
	OrderStatus     string
	Timestamp       time.Time
	Cart            map[string]*ProductOrder
	Total           float64
	Discount        float64
	DiscountReasons []string
}

type InventoryStock struct {
	Product    string
	Quantity   int
	LowWarning int
}
type Product struct {
	ID    string
	Name  string
	Price float64
}

type ProductOrder struct {
	ID       string
	Quantity int
}

type Discount interface {
	// Discount applies discount to cart and returns the value to discount and the reason, if any
	Discount(map[string]*ProductOrder, map[string]*Product) (float64, string)
}

type PercentDiscount struct {
	ProductID  string
	Percentage float64
}

func (d *PercentDiscount) Discount(cart map[string]*ProductOrder, products map[string]*Product) (float64, string) {
	order, ook := cart[d.ProductID]
	prod, pok := products[d.ProductID]
	if ook && pok {
		reason := fmt.Sprintf("%d x %.2f Off for %s", order.Quantity, d.Percentage, prod.Name)
		return float64(order.Quantity) * (prod.Price * (d.Percentage / 100.0)), reason
	}
	return 0, ""
}

type AnyXForY struct {
	ProductIDs []string
	X          int
	Y          int
}

func (d *AnyXForY) Discount(cart map[string]*ProductOrder, products map[string]*Product) (float64, string) {
	acc := 0
	for _, p := range cart {
		if StringSliceContains(d.ProductIDs, p.ID) {
			acc += p.Quantity
		}
	}
	// division gives the number of times the discount should apply because it's integer division
	num := acc / d.X
	// if num is 0 (< 1) it's because there are no enough matches to apply discount
	if num < 1 {
		return 0, ""
	}
	prod := GetCheapestOf(products, d.ProductIDs)
	diff := d.X - d.Y
	reason := fmt.Sprintf("%d x %d for %d", num, d.X, d.Y)
	return float64(num) * float64(diff) * prod.Price, reason
}

func GetCheapestOf(pMap map[string]*Product, ids []string) *Product {
	var prod *Product = nil
	price := math.MaxFloat64
	for _, id := range ids {
		if p, ok := pMap[id]; ok && p.Price < price {
			price = p.Price
			prod = p
		}
	}
	return prod
}
