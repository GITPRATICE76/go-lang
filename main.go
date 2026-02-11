package main

import (
	"log"
	"time"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {

	r := gin.Default()

	// ✅ CORS
	r.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			return strings.HasPrefix(origin, "http://localhost")
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 🔓 Public Routes
	r.POST("/api/login", Login)
	r.POST("/api/createaccount", Register)

	// 🔐 Protected Routes (JWT required)
	auth := r.Group("/api")
	auth.Use(AuthMiddleware())
	{
		auth.GET("/me", Me)
		auth.POST("/applyleave", ApplyLeave)
		auth.GET("/leaves", GetLeaves)
		auth.POST("/leave/action", LeaveAction)
		auth.GET("/org-chart", GetOrgChart)
		auth.GET("/leave/analytics", GetLeaveAnalytics)
		
	}

	log.Println("Server running on port 8080")
	r.Run(":8080")
}
