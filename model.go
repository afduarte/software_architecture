package main

import (
	"bytes"
	"encoding/json"

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
	// AdminRole > UserRole
	ManagerRole
)

func (s PermissionRole) String() string {
	return roleToString[s]
}

var roleToString = map[PermissionRole]string{
	UserRole:    "UserRole",
	ManagerRole: "AdminRole",
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
	Token    string
	Role     PermissionRole
}

type Order struct {
	ID       string
	Products []Product
	Discount []Discount
}

type InventoryStock struct {
	Product    string
	Quantity   int
	LowWarning int
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
