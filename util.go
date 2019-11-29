package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

// HydrateUserMiddleware is a simple middleware that checks if a user is logged in and Hydrates
func HydrateUserMiddleware(s *Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := ParseBearerToken(c.GetHeader("Authorization"))
		println("hydrate middleware got token " + token)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		user := FetchUser(s.config.authEndpoint, token)
		println("hydrate middleware got user " + user.Name)
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
		println("permish middleware got user " + user.Name)
		if user.Role < p {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		// Continue down the chain to handler etc
		c.Next()
	}
}

func ParseBearerToken(header string) string {
	split := strings.Split(header, "Bearer")
	if len(split) != 2 {
		return ""
	}
	return strings.TrimSpace(split[1])
}

func FetchUser(authEndpoint, token string) *User {
	formData := url.Values{"token": {token}}

	req, err := http.NewRequest("POST", authEndpoint+"/user", strings.NewReader(formData.Encode()))
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
