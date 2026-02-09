package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type OrgUser struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	Role       string  `json:"role"`
	Department string  `json:"department"`
	Team       *string `json:"team"`
}

func GetOrgChart(c *gin.Context) {

	db, err := ConnectDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Database connection failed",
		})
		return
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT id, name, role, department, team
		FROM users
		ORDER BY role, department, team
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to fetch users",
		})
		return
	}
	defer rows.Close()

	users := []OrgUser{}

	for rows.Next() {
		var u OrgUser
		rows.Scan(
			&u.ID,
			&u.Name,
			&u.Role,
			&u.Department,
			&u.Team,
		)
		users = append(users, u)
	}

	c.JSON(http.StatusOK, users)
}
