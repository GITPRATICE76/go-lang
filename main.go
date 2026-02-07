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
 

	r.Use(cors.New(cors.Config{

AllowOriginFunc: func(origin string) bool {
            return strings.HasPrefix(origin, "http://localhost")
        },
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},

		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},

		AllowCredentials: true,

		MaxAge:           12 * time.Hour,

	}))
 
	r.POST("/api/login", Login)

	r.POST("/api/createaccount", Register)

	r.GET("/api/me",Me)
	r.POST("/api/applyleave",ApplyLeave)
	r.GET("/api/leaves", GetLeaves)




 
	log.Println("Server running on port 8080")

	r.Run(":8080")

}

 