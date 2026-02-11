package main

import (
	"database/sql"
	
	"time"

	"github.com/gin-gonic/gin"
)

func GetLeaveAnalytics(c *gin.Context) {

	startDate := c.Query("start")
	endDate := c.Query("end")

	db, err := ConnectDB()
	if err != nil {
		c.JSON(500, gin.H{"message": "DB connection failed"})
		return
	}
	defer db.Close()

	// ✅ Total active employees
	var totalResources int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM users 
		WHERE role = 'EMPLOYEE'
	`).Scan(&totalResources)

	if err != nil {
		c.JSON(500, gin.H{"message": "Failed to count employees"})
		return
	}

	start, _ := time.Parse("2006-01-02", startDate)
	end, _ := time.Parse("2006-01-02", endDate)

	var result []gin.H

	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {

		dateStr := d.Format("2006-01-02")

		// ✅ IMPORTANT FIX: Use BETWEEN
		rows, err := db.Query(`
			SELECT u.name
			FROM leaves l
			JOIN users u ON l.user_id = u.id
			WHERE l.status IN ('PENDING', 'APPROVED')
			AND @date BETWEEN l.from_date AND l.to_date
		`, sql.Named("date", dateStr))

		if err != nil {
			continue
		}

		var employees []string
		for rows.Next() {
			var name string
			rows.Scan(&name)
			employees = append(employees, name)
		}
		rows.Close()

		onLeave := len(employees)

		leavePercent := 0.0
		availablePercent := 0.0
		remainingAllowed := 8.0

		if totalResources > 0 {
			leavePercent = float64(onLeave) / float64(totalResources) * 100
			availablePercent = 100 - leavePercent
			remainingAllowed = 8 - leavePercent
		}

		result = append(result, gin.H{
			"date":                        dateStr,
			"total_resources":             totalResources,
			"on_leave":                    onLeave,
			"leave_percentage":            leavePercent,
			"available_percentage":        availablePercent,
			"remaining_allowed_percentage": remainingAllowed,
			"employees_on_leave":          employees,
		})
	}

	c.JSON(200, result)
}

