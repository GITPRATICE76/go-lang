package main

import (
	"database/sql"
	"fmt"

	"github.com/gin-gonic/gin"
)

func GetEmployeeDashboardSummary(c *gin.Context) {

	db, err := ConnectDB()
	if err != nil {
		c.JSON(500, gin.H{"message": "DB connection failed"})
		return
	}
	defer db.Close()

	// userID := c.GetInt("userID")
	userID := c.GetInt("user_id")

	if userID == 0 {
		c.JSON(401, gin.H{"message": "Invalid user"})
		return
	}

	var totalApproved, pending, rejected, casual, sick, onLeaveCount int

	// ✅ Approved
	err = db.QueryRow(`
		SELECT COUNT(*) 
		FROM leaves 
		WHERE user_id = @userID AND status = 'APPROVED'
	`, sql.Named("userID", userID)).Scan(&totalApproved)
	if err != nil {
		fmt.Println("DB ERROR:", err)
		c.JSON(500, gin.H{"message": err.Error()})
		return
	}

	// ✅ Pending
	err = db.QueryRow(`
		SELECT COUNT(*) 
		FROM leaves 
		WHERE user_id = @userID AND status = 'PENDING'
	`, sql.Named("userID", userID)).Scan(&pending)
	if err != nil {
		fmt.Println("DB ERROR:", err)
		c.JSON(500, gin.H{"message": err.Error()})
		return
	}

	// ✅ Rejected
	err = db.QueryRow(`
		SELECT COUNT(*) 
		FROM leaves 
		WHERE user_id = @userID AND status = 'REJECTED'
	`, sql.Named("userID", userID)).Scan(&rejected)
	if err != nil {
		fmt.Println("DB ERROR:", err)
		c.JSON(500, gin.H{"message": err.Error()})
		return
	}

	err = db.QueryRow(`
	SELECT ISNULL(SUM(DATEDIFF(DAY, from_date, to_date) + 1), 0)
	FROM leaves 
	WHERE user_id = @userID 
	AND status = 'APPROVED'
	AND leave_type = 'CASUAL'
`, sql.Named("userID", userID)).Scan(&casual)
	if err != nil {
		fmt.Println("DB ERROR:", err)
		c.JSON(500, gin.H{"message": err.Error()})
		return
	}

	err = db.QueryRow(`
	SELECT ISNULL(SUM(DATEDIFF(DAY, from_date, to_date) + 1), 0)
	FROM leaves 
	WHERE user_id = @userID 
	AND status = 'APPROVED'
	AND leave_type = 'SICK'
`, sql.Named("userID", userID)).Scan(&sick)
	if err != nil {
		fmt.Println("DB ERROR:", err)
		c.JSON(500, gin.H{"message": err.Error()})
		return
	}

	// ✅ Currently On Leave
	err = db.QueryRow(`
		SELECT COUNT(*) 
		FROM leaves
		WHERE user_id = @userID
		AND status = 'APPROVED'
		AND CAST(GETDATE() AS DATE) BETWEEN from_date AND to_date
	`, sql.Named("userID", userID)).Scan(&onLeaveCount)
	if err != nil {
		fmt.Println("DB ERROR:", err)
		c.JSON(500, gin.H{"message": err.Error()})
		return
	}

	currentlyOnLeave := onLeaveCount > 0

	// 🔥 NEW PART: Get Team Members On Leave

	var team string
	err = db.QueryRow(`
		SELECT team FROM users WHERE id = @userID
	`, sql.Named("userID", userID)).Scan(&team)

	if err != nil {
		fmt.Println("TEAM ERROR:", err)
		c.JSON(500, gin.H{"message": "Failed to get team"})
		return
	}

	rows, err := db.Query(`
		SELECT u.id, u.name
		FROM users u
		JOIN leaves l ON u.id = l.user_id
		WHERE u.team = @team
		AND u.id != @userID
		AND l.status = 'APPROVED'
		AND CAST(GETDATE() AS DATE) BETWEEN l.from_date AND l.to_date
	`,
		sql.Named("team", team),
		sql.Named("userID", userID),
	)

	if err != nil {
		fmt.Println("TEAM QUERY ERROR:", err)
		c.JSON(500, gin.H{"message": "Failed to fetch team members"})
		return
	}
	defer rows.Close()

	var teamMembers []gin.H

	for rows.Next() {
		var id int
		var name string
		rows.Scan(&id, &name)

		teamMembers = append(teamMembers, gin.H{
			"id":   id,
			"name": name,
		})
	}
	var leaveRemarks []gin.H

	rowsRemarks, err := db.Query(`
	SELECT TOP 10 id, status, remarks, leave_type, from_date, to_date
	FROM leaves
	WHERE user_id = @userID
	ORDER BY created_at DESC
`, sql.Named("userID", userID))

	if err != nil {
		fmt.Println("REMARKS ERROR:", err)
		c.JSON(500, gin.H{"message": "Failed to fetch remarks"})
		return
	}
	defer rowsRemarks.Close()

	for rowsRemarks.Next() {
		var id int
		var status, remarks, leaveType string
		var fromDate, toDate string

		rowsRemarks.Scan(&id, &status, &remarks, &leaveType, &fromDate, &toDate)

		leaveRemarks = append(leaveRemarks, gin.H{
			"id":         id,
			"status":     status,
			"remarks":    remarks,
			"leave_type": leaveType,
			"from_date":  fromDate,
			"to_date":    toDate,
		})
	}

	// ✅ Final Response (Added team data)
	c.JSON(200, gin.H{
		"total_leaves_taken":    totalApproved,
		"pending_requests":      pending,
		"rejected_requests":     rejected,
		"casual_leaves":         casual,
		"sick_leaves":           sick,
		"currently_on_leave":    currentlyOnLeave,
		"team_members_on_leave": teamMembers,
		"team_total_on_leave":   len(teamMembers),
		"leave_remarks":         leaveRemarks,
	})
}
