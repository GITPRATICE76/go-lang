package main
 
import (

	"log"
	"time"

	"os"

	"time"
 
	"github.com/gin-contrib/cors"

	"github.com/gin-gonic/gin"

)
 
func main() {
 
	r := gin.Default()
 
	// ✅ CORS Configuration (Allow all origins for now)

	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
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

		auth.GET("/dashboard/summary", GetDashboardSummary)
		auth.GET("/employee/dashboard", GetEmployeeDashboardSummary)
		auth.GET("/holidays", GetHolidays)

	}

}

 