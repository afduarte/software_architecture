package main

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	seshkey = "supersafekey"
)

type LoginResponse struct {
	User    *User
	Message string
}

var UserMap = map[string]*User{
	"antero":  &User{"antero", "supersafepassword", "Antero Duarte", "", ManagerRole},
	"alex": &User{"alex", "supersafepassword", "Alex Smith", "", UserRole},
}

var LoggedInUsers map[string]*User = make(map[string]*User)

func UserLogin(username, password string) (*User, string) {
	u, ok := UserMap[strings.Trim(username, " ")]
	if !ok {
		return nil, "user does not exist"
	}
	if u.password != strings.Trim(password, " ") {
		return nil, "wrong password"
	}
	token := uuid.Must(uuid.NewRandom())
	u.Token = token.String()
	return u, "user logged in"
}

func login(s *Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user string
		var password string

		if user = c.PostForm("user"); user == "" {
			c.JSON(http.StatusBadRequest, gin.H{"Message": "user field missing"})
			return
		}

		if password = c.PostForm("pass"); password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"Message": "pass field missing"})
			return
		}
		u, message := UserLogin(user, password)
		if u == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"Message": "bad credentials"})
			return
		}
		LoggedInUsers[u.Token] = u
		c.JSON(http.StatusOK, LoginResponse{u, message})
	}
}

func userinfo(s *Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := ParseBearerToken(c.GetHeader("Authorization"))
		if token == "" {
			c.JSON(http.StatusBadRequest, gin.H{"Message": "token missing"})
			return
		}
		user, ok := LoggedInUsers[token]
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"Message": "bad credentials"})
			return
		}
		c.JSON(http.StatusOK, user)
	}
}

func users(s *Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := ParseBearerToken(c.GetHeader("Authorization"))
		if token == "" {
			c.JSON(http.StatusBadRequest, gin.H{"Message": "token missing"})
			return
		}

		user, ok := LoggedInUsers[token]
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"Message": "bad credentials"})
			return
		}
		c.JSON(http.StatusOK, user)
	}
}

func AuthRoutes(s *Server) {
	s.router.POST("/login", login(s))
	s.router.GET("/info", userinfo(s))

	private := s.router.Group("/user")	
	private.Use(HydrateUserMiddleware(s))
	private.GET("/list", users(s))
}
