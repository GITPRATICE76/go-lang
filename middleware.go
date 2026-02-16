package main

import (
	
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {

        authHeader := c.GetHeader("Authorization")

        if authHeader == "" {
            c.JSON(401, gin.H{"message": "Authorization header missing"})
            c.Abort()
            return
        }

        tokenString := strings.TrimPrefix(authHeader, "Bearer ")

        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            return jwtSecret, nil
        })

        if err != nil || !token.Valid {
            c.JSON(401, gin.H{"message": "Invalid token"})
            c.Abort()
            return
        }

        // claims := token.Claims.(jwt.MapClaims)

        // c.Set("userID", claims["id"])
        claims := token.Claims.(jwt.MapClaims)

// JWT numbers are float64
idFloat, ok := claims["id"].(float64)
if !ok {
    c.JSON(401, gin.H{"message": "Invalid token id"})
    c.Abort()
    return
}

c.Set("userID", int(idFloat))
c.Set("role", claims["role"])

        // c.Set("role", claims["role"])

        c.Next()
    }
}

