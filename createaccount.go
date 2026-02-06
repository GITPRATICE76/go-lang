package main
 
import (

	"net/http"
 "database/sql"

	"github.com/gin-gonic/gin"

)
 
type RegisterRequest struct {

	Name       string `json:"name"`

	Email      string `json:"email"`

	Password   string `json:"password"`

	Department string `json:"department"`

	Team       string `json:"team"`

}
 
func Register(c *gin.Context) {

	var req RegisterRequest
 
	// Read request body

	if err := c.ShouldBindJSON(&req); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{

			"message": "Invalid request body",

		})

		return

	}
 

	if !isValidDepartmentTeam(req.Department, req.Team) {

		c.JSON(http.StatusBadRequest, gin.H{

			"message": "Invalid department and team combination",

		})

		return

	}
 
	db, err := ConnectDB()

	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{

			"message": "Database connection failed",

		})

		return

	}

	defer db.Close()
 
	query := `

		INSERT INTO users (name, email, password, role, department, team)

		VALUES (@name, @email, @password, 'EMPLOYEE', @department, @team)

	`
 
	_, err = db.Exec(

		query,

		sql.Named("name", req.Name),

		sql.Named("email", req.Email),

		sql.Named("password", req.Password),

		sql.Named("department", req.Department),

		sql.Named("team", req.Team),

	)
 
	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{

			"message": "User already exists or DB error",

		})

		return

	}
 
	c.JSON(http.StatusOK, gin.H{

		"message": "User registered successfully",

	})

}
func isValidDepartmentTeam(dept, team string) bool {

	if dept == "QA" && team == "QA" {

		return true

	}
 
	if dept == "DEVELOPMENT" {

		validTeams := map[string]bool{

			"REACT":   true,

			"BACKEND": true,

			"WEB":     true,

			"DB":      true,

		}

		return validTeams[team]

	}
 
	return false

}

 

 