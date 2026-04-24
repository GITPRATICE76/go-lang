package main

import (
	"database/sql"
	"math"
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

	// 🔥 Get user info
	// userID := c.GetInt("user_id")
	// role := c.GetString("role")

	// var team string

	// // 👉 If not manager → get team
	// if role != "MANAGER" {
	// 	err = db.QueryRow(`
	// 		SELECT team FROM users WHERE id = @userID
	// 	`, sql.Named("userID", userID)).Scan(&team)

	// 	if err != nil {
	// 		c.JSON(500, gin.H{"message": "Failed to get team"})
	// 		return
	// 	}
	// }

	// 🔥 Total resources (role based)
	var totalResources int

	// if role == "MANAGER" {
	err = db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&totalResources)
	// } else {
	// 	err = db.QueryRow(`
	// 		SELECT COUNT(*) FROM users WHERE team = @team
	// 	`, sql.Named("team", team)).Scan(&totalResources)
	// }

	if err != nil {
		c.JSON(500, gin.H{"message": "Failed to count employees"})
		return
	}

	start, _ := time.Parse("2006-01-02", startDate)
	end, _ := time.Parse("2006-01-02", endDate)

	var result []gin.H

	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {

		dateStr := d.Format("2006-01-02")

		var rows *sql.Rows

		// 🔥 Role-based query
		rows, err = db.Query(`
	SELECT u.name, l.status
	FROM leaves l
	JOIN users u ON l.user_id = u.id
	WHERE l.status IN ('APPROVED', 'PENDING')
	AND @date BETWEEN l.from_date AND l.to_date
`, sql.Named("date", dateStr))

		if err != nil {
			continue
		}

		// 🔥 NEW: separate lists
		var approvedEmployees []string
		var pendingEmployees []string

		approvedCount := 0
		pendingCount := 0

		for rows.Next() {
			var name, status string
			rows.Scan(&name, &status)

			switch status {
			case "APPROVED":
				approvedCount++
				approvedEmployees = append(approvedEmployees, name)

			case "PENDING":
				pendingCount++
				pendingEmployees = append(pendingEmployees, name)
			}
		}

		rows.Close()

		
		onLeave := approvedCount 

		leavePercent := 0.0
		availablePercent := 0.0
		remainingAllowed := 8.0

		if totalResources > 0 {
			leavePercent = math.Round((float64(onLeave) / float64(totalResources)) * 100)
			availablePercent = math.Round(100 - leavePercent)
			remainingAllowed = math.Round(8 - leavePercent)
		}

		result = append(result, gin.H{
			"date":                         dateStr,
			"total_resources":              totalResources,
			"on_leave":                     onLeave,
			"leave_percentage":             leavePercent,
			"available_percentage":         availablePercent,
			"remaining_allowed_percentage": remainingAllowed,

			// 🔥 lists
			"employees_on_leave": approvedEmployees,
			"employees_pending":  pendingEmployees,

			// 🔥 counts
			"approved": approvedCount,
			"pending":  pendingCount,
		})
	}

	c.JSON(200, result)
}
