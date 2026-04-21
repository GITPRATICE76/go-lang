package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Request structure
type RegisterRequest struct {
	Name       string `json:"name"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	Department string `json:"department"`
	Team       string `json:"team"`
}

// 🔥 Register API
func Register(c *gin.Context) {

	var req RegisterRequest

	// Bind JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request body",
		})
		return
	}

	// Trim inputs (avoid hidden spaces)
	req.Department = strings.TrimSpace(req.Department)
	req.Team = strings.TrimSpace(req.Team)

	// ✅ Connect DB FIRST
	db, err := ConnectDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Database connection failed",
		})
		return
	}
	defer db.Close()

	// 🔍 Debug (optional – remove later)
	fmt.Println("Dept:", "["+req.Department+"]")
	fmt.Println("Team:", "["+req.Team+"]")

	// ✅ Dynamic validation from DB
	if !isValidDepartmentTeam(db, req.Department, req.Team) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid department and team combination",
		})
		return
	}

	// ✅ Insert user
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
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User registered successfully",
	})
}

// 🔥 Dynamic validation (NO HARDCODING)
func isValidDepartmentTeam(db *sql.DB, dept, team string) bool {

	query := `
		SELECT COUNT(1)
		FROM tbl_UserCodeDetail
		WHERE 
			UPPER(LTRIM(RTRIM(MasterID))) = UPPER(LTRIM(RTRIM(@p1)))
		AND UPPER(LTRIM(RTRIM(SubCodeID))) = UPPER(LTRIM(RTRIM(@p2)))
	`

	var count int

	err := db.QueryRow(query, dept, team).Scan(&count)
	if err != nil {
		fmt.Println("Query error:", err)
		return false
	}

	fmt.Println("Match Count:", count) // debug

	return count > 0
} 