package main

import (
	"log"

	"os"

	"time"

	"github.com/gin-contrib/cors"

	"github.com/gin-gonic/gin"
)

func main() {

	r := gin.Default()

	// ✅ CORS Configuration (Allow all origins for now)

	r.Use(cors.New(cors.Config{

		AllowOrigins: []string{"*"}, // Allow all (we can restrict later)

		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},

		AllowHeaders: []string{"Origin", "Content-Type", "Authorization"},

		AllowCredentials: false,

		MaxAge: 12 * time.Hour,
	}))

	// 🔓 Public Routes

	r.POST("/api/login", Login)

	r.POST("/api/createaccount", Register)

	auth := r.Group("/api")

	auth.Use(AuthMiddleware())

	{
		// refreash api
		auth.GET("/me", Me)
		// apply leave api for employees omly
		auth.POST("/applyleave", ApplyLeave)
		// get all applied negative by employee
		auth.POST("/leaves", GetLeaves)
		// manager reply to leaves applied by the customer
		auth.POST("/leave/action", LeaveAction)
		// get organization chart
		auth.GET("/org-chart", GetOrgChart)
		// barchart,card
		auth.GET("/leave/analytics", GetLeaveAnalytics)
		// highest leave date,peak leave weak,top leave taker
		auth.GET("/dashboard/summary", GetDashboardSummary)

		auth.GET("/employee/dashboard", GetEmployeeDashboardSummary)
		// holiday applied by empolyee and government leaves
		auth.GET("/holidays", GetHolidays)
		// leave history
		auth.POST("/leave/history", GetLeaveHistory)

	}

	// ✅ IMPORTANT: Use Render PORT

	port := os.Getenv("PORT")

	if port == "" {

		port = "8085" // for local development

	}

	log.Println("Server running on port", port)

	if err := r.Run(":" + port); err != nil {

		log.Fatal("Failed to start server:", err)

	}

}
