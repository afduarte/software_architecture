package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

// General Util Funcs

func StringSliceContains(haystack []string, needle string) bool {
	for _, h := range haystack {
		if h == needle {
			return true
		}
	}
	return false
}

func ParseBearerToken(header string) string {
	split := strings.Split(header, "Bearer")
	if len(split) != 2 {
		return ""
	}
	return strings.TrimSpace(split[1])
}

func MustGetToken(c *gin.Context) string {
	token := ParseBearerToken(c.GetHeader("Authorization"))
	if token == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return ""
	}
	return token
}

// Middlewares

// HydrateUserMiddleware is a simple middleware that checks if a user is logged in and Hydrates
func HydrateUserMiddleware(s *Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := MustGetToken(c)
		user := FetchUser(s.config.authEndpoint, token)
		if user == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Set("user", user)
		// Continue down the chain to handler etc
		c.Next()
	}
}

// RequiresPermissionMiddleware is a simple middleware that checks if a user has permission to see the route. MUST COME AFTER HydrateUserMiddleware
func RequiresPermissionMiddleware(p PermissionRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := c.MustGet("user").(*User)
		if user.Role < p {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		// Continue down the chain to handler etc
		c.Next()
	}
}

// Funcs to communicate between microservices

func FetchUser(authEndpoint, token string) *User {
	formData := url.Values{"token": {token}}
	req, err := http.NewRequest("POST", authEndpoint+"/userinfo", strings.NewReader(formData.Encode()))
	if err != nil {
		return nil
	}
	req.Header.Add("Authorization", "Bearer "+token)
	response, err := http.DefaultClient.Do(req)
	if err != nil || response.StatusCode != http.StatusOK {
		return nil
	}
	var user User
	json.NewDecoder(response.Body).Decode(&user)
	return &user
}

func FetchInventory(inventoryEndpoint, token string) map[string]*InventoryStock {
	req, err := http.NewRequest("GET", inventoryEndpoint, nil)
	if err != nil {
		return nil
	}
	req.Header.Add("Authorization", "Bearer "+token)
	response, err := http.DefaultClient.Do(req)
	if err != nil || response.StatusCode != http.StatusOK {
		return nil
	}
	var stock map[string]*InventoryStock
	json.NewDecoder(response.Body).Decode(&stock)
	return stock
}

func FetchProductPrices(priceEndpoint, token string) map[string]*Product {
	req, err := http.NewRequest("GET", priceEndpoint, nil)
	if err != nil {
		return nil
	}
	req.Header.Add("Authorization", "Bearer "+token)
	response, err := http.DefaultClient.Do(req)
	if err != nil || response.StatusCode != http.StatusOK {
		return nil
	}
	var prices map[string]*Product
	json.NewDecoder(response.Body).Decode(&prices)
	return prices
}

func SendDecrementRequest(inventoryEndpoint, token string, decrements map[string]*ProductOrder) error {
	jsonDecrements, jsonErr := json.Marshal(decrements)
	if jsonErr != nil {
		return jsonErr
	}
	req, err := http.NewRequest("POST", inventoryEndpoint+"/decrement", bytes.NewBuffer(jsonDecrements))
	if err != nil {
		return errors.New("unable to send request to inventory server")
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(req)
	if err != nil || response.StatusCode != http.StatusOK {
		return errors.New("inventory server was unable to fulfil order")
	}
	return nil
}

func SendCalculateCartRequest(priceEndpoint, token string, cart map[string]*ProductOrder) (*CartValueResponse, error) {
	jsonDecrements, jsonErr := json.Marshal(cart)
	if jsonErr != nil {
		return nil, jsonErr
	}
	req, err := http.NewRequest("POST", priceEndpoint+"/calculate", bytes.NewBuffer(jsonDecrements))
	if err != nil {
		return nil, errors.New("unable to send request to price server")
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(req)
	if err != nil || response.StatusCode != http.StatusOK {
		return nil, errors.New("price server was unable to fulfil order")
	}
	var res CartValueResponse
	json.NewDecoder(response.Body).Decode(&res)
	return &res, nil
}

func SendUpdatePointsRequest(loyaltyEndpoint, token string, customerID string, cart map[string]*ProductOrder, usePoints int) (*UpdatePointsResponse, error) {
	jsonRequest, jsonErr := json.Marshal(UpdatePointsRequest{customerID, cart, usePoints})
	if jsonErr != nil {
		return nil, jsonErr
	}
	req, err := http.NewRequest("POST", loyaltyEndpoint+"/update-points", bytes.NewBuffer(jsonRequest))
	if err != nil {
		return nil, errors.New("unable to send request to loyalty server")
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(req)
	if err != nil || response.StatusCode != http.StatusOK {
		return nil, errors.New("loyalty server was unable to fulfil order")
	}
	var res UpdatePointsResponse
	json.NewDecoder(response.Body).Decode(&res)
	return &res, nil
}
