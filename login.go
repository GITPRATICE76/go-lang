package main

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("my_secret_key") // change this in production

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request"})
		return
	}

	db, err := ConnectDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "DB connection failed"})
		return
	}
	defer db.Close()

	query := `
		SELECT id, name, role, department, team, password
		FROM users
		WHERE email = @email
	`

	row := db.QueryRow(
		query,
		sql.Named("email", req.Email),
	)

	var (
		id         int
		name       string
		role       string
		department string
		team       *string
		dbPassword string
	)

	err = row.Scan(&id, &name, &role, &department, &team, &dbPassword)
	if err != nil || dbPassword != req.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid email or password"})
		return
	}

	// 🔥 Create JWT Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":         id,
		"name":       name,
		"role":       role,
		"department": department,
		"exp":        time.Now().Add(time.Hour * 24).Unix(), // expires in 24h
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Could not generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":      tokenString,
		"id":         id,
		"name":       name,
		"role":       role,
		"department": department,
		"team":       team,
	})
}
