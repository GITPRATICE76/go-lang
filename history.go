package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetLeaveHistory(c *gin.Context) {

	start := c.Query("start")
	end := c.Query("end")
	month := c.Query("month")
	year := c.Query("year")

	query := `
	SELECT 
		l.id,
		u.name AS employee_name,
		u.team,
		u.department,
		l.from_date,
		l.to_date,
		l.leave_type,
		l.reason,
		l.status,
		l.created_at
	FROM leaves l
	JOIN users u ON l.user_id = u.id
	`

	var args []interface{}

	if start != "" && end != "" {

		query += " WHERE l.from_date BETWEEN @p1 AND @p2"
		args = append(args, start, end)

	} else if month != "" && year != "" {

		query += " WHERE MONTH(l.from_date) = @p1 AND YEAR(l.from_date) = @p2"
		args = append(args, month, year)

	} else if year != "" {

		query += " WHERE YEAR(l.from_date) = @p1"
		args = append(args, year)

	}

	query += " ORDER BY l.from_date DESC"

	// connect DB using your existing db.go
	db, err := ConnectDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database connection failed",
		})
		return
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	defer rows.Close()

	var results []gin.H

	for rows.Next() {

		var (
			id        int
			name      string
			team      string
			dept      string
			fromDate  string
			toDate    string
			leaveType string
			reason    string
			status    string
			createdAt string
		)

		err := rows.Scan(
			&id,
			&name,
			&team,
			&dept,
			&fromDate,
			&toDate,
			&leaveType,
			&reason,
			&status,
			&createdAt,
		)

		if err != nil {
			continue
		}

		results = append(results, gin.H{
			"id":            id,
			"employee_name": name,
			"team":          team,
			"department":    dept,
			"from_date":     fromDate,
			"to_date":       toDate,
			"leave_type":    leaveType,
			"reason":        reason,
			"status":        status,
			"created_at":    createdAt,
		})
	}

	c.JSON(http.StatusOK, results)
}
