package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type OrgUser struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Role        string  `json:"role"`
	Department  string  `json:"department"`
	Team        *string `json:"team"`
	ReportingTo *int    `json:"reporting_to"` // Added this field
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

	// Query now selects reporting_to and orders by it to keep
	// subordinates closer to their managers in the list
	rows, err := db.Query(`
        SELECT id, name, role, department, team, reporting_to
        FROM users
        ORDER BY reporting_to ASC, role DESC
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
		err := rows.Scan(
			&u.ID,
			&u.Name,
			&u.Role,
			&u.Department,
			&u.Team,
			&u.ReportingTo, // Scan the new field
		)
		if err != nil {
			continue
		}
		users = append(users, u)
	}

	c.JSON(http.StatusOK, users)
}
