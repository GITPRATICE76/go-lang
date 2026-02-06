package main
 
import (

	"net/http"
	"database/sql"
 
	"github.com/gin-gonic/gin"

)
 
type LoginRequest struct {

	Email    string `json:"email"`

	Password string `json:"password"`

}
 
func Login(c *gin.Context) {

	var req LoginRequest


	if err := c.ShouldBindJSON(&req); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{

			"message": "Invalid request",

		})

		return

	}
 
	db, err := ConnectDB()

	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{

			"message": "DB connection failed",

		})

		return

	}

	defer db.Close()
 
	query := `

		SELECT id, name, role, department, team

		FROM users

		WHERE email = @email AND password = @password

	`
 
	row := db.QueryRow(

		query,

		sql.Named("email", req.Email),

		sql.Named("password", req.Password),

	)
 
	var (

		id         int

		name       string

		role       string

		department string

		team       *string

	)
 
	err = row.Scan(&id, &name, &role, &department, &team)

	if err != nil {

		c.JSON(http.StatusUnauthorized, gin.H{

			"message": "Invalid email or password",

		})

		return

	}
 
	c.JSON(http.StatusOK, gin.H{

		"id":         id,

		"name":       name,

		"role":       role,

		"department": department,

		"team":       team,

	})

}

 