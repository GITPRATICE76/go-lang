package main
 
import (

	"log"

	"time"
 
	"github.com/gin-contrib/cors"

	"github.com/gin-gonic/gin"

)
 
func main() {

	r := gin.Default()
 
	// ✅ CORS configuration

	r.Use(cors.New(cors.Config{

		AllowOrigins:     []string{"http://localhost:5174", "http://localhost:3000"},

		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},

		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},

		AllowCredentials: true,

		MaxAge:           12 * time.Hour,

	}))
 
	r.POST("/api/login", Login)

	r.POST("/api/register", Register)


 
	log.Println("Server running on port 8080")

	r.Run(":8080")

}

 