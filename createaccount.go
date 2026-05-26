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
	Name        string `json:"name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	Department  string `json:"department"`
	Team        string `json:"team"`
	ReportingTo int    `json:"reporting_to"` // ✅ ADDED
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

	// Trim inputs
	req.Department = strings.TrimSpace(req.Department)
	req.Team = strings.TrimSpace(req.Team)

	// Connect DB
	db, err := ConnectDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Database connection failed",
		})
		return
	}
	defer db.Close()

	// Debug
	fmt.Println("Dept:", "["+req.Department+"]")
	fmt.Println("Team:", "["+req.Team+"]")
	fmt.Println("ReportingTo:", req.ReportingTo) // ✅ DEBUG

	// Validate department + team
	if !isValidDepartmentTeam(db, req.Department, req.Team) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid department and team combination",
		})
		return
	}

	// Insert user (UPDATED)
	query := `
		INSERT INTO users (name, email, password, role, department, team, reporting_to)
		VALUES (@name, @email, @password, 'EMPLOYEE', @department, @team, @reporting_to)
	`

	_, err = db.Exec(
		query,
		sql.Named("name", req.Name),
		sql.Named("email", req.Email),
		sql.Named("password", req.Password),
		sql.Named("department", req.Department),
		sql.Named("team", req.Team),
		sql.Named("reporting_to", req.ReportingTo), // ✅ ADDED
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

// 🔥 Dynamic validation
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

	fmt.Println("Match Count:", count)

	return count > 0
}
func GetReportingTo(c *gin.Context) {

	db, err := ConnectDB()
	if err != nil {
		c.JSON(500, gin.H{"message": "DB connection failed"})
		return
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT id, name, role 
		FROM users 
		WHERE role IN ('MANAGER', 'RO')
	`)
	if err != nil {
		c.JSON(500, gin.H{"message": err.Error()})
		return
	}
	defer rows.Close()

	var data []gin.H

	for rows.Next() {
		var id int
		var name, role string

		err := rows.Scan(&id, &name, &role)
		if err != nil {
			continue
		}

		data = append(data, gin.H{
			"value": id,
			"label": name + " (" + role + ")",
		})
	}

	c.JSON(200, data)
}
