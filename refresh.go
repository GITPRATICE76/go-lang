package main
 
import (

	"database/sql"

	"net/http"

	"strconv"
 
	"github.com/gin-gonic/gin"

)
 
func Me(c *gin.Context) {


	userIdStr := c.Query("user_id")

	if userIdStr == "" {

		c.JSON(http.StatusBadRequest, gin.H{

			"message": "user_id is required",

		})

		return

	}
 
	userId, err := strconv.Atoi(userIdStr)

	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{

			"message": "Invalid user_id",

		})

		return

	}
 
	// 2️⃣ Connect DB

	db, err := ConnectDB()

	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{

			"message": "Database connection failed",

		})

		return

	}

	defer db.Close()
 
	// 3️⃣ Query user

	query := `

		SELECT id, name, email, role, department, team

		FROM users

		WHERE id = @id

	`
 
	row := db.QueryRow(query, sql.Named("id", userId))
 
	var (

		id         int

		name       string

		email      string

		role       string

		department string

		team       *string

	)
 
	err = row.Scan(&id, &name, &email, &role, &department, &team)

	if err != nil {

		c.JSON(http.StatusNotFound, gin.H{

			"message": "User not found",

		})

		return

	}

	c.JSON(http.StatusOK, gin.H{

		"id":         id,

		"name":       name,

		"email":      email,

		"role":       role,

		"department": department,

		"team":       team,

	})

}

 