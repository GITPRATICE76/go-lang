package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type LeaveResponse struct {
	ID           int    `json:"id"`
	EmployeeName string `json:"employeeName"`
	LeaveType    string `json:"leaveType"`
	Status       string `json:"status"`
	Reason       string `json:"reason"`
	From         string `json:"from"`
	To           string `json:"to"`
	Days         int    `json:"days"`
}

func GetLeaves(c *gin.Context) {

	db, err := ConnectDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Database connection failed",
		})
		return
	}
	defer db.Close()

	query := `
	SELECT 
		l.id,
		u.name,
		l.leave_type,
		l.status,
		l.reason,
		l.from_date,
		l.to_date
	FROM leaves l
	JOIN users u ON l.user_id = u.id
	ORDER BY l.created_at DESC
`

	rows, err := db.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to fetch leaves",
		})
		return
	}
	defer rows.Close()

	var leaves []LeaveResponse

	for rows.Next() {
		var (
			leave LeaveResponse
			from  time.Time
			to    time.Time
		)

		err := rows.Scan(
			&leave.ID,
			&leave.EmployeeName,
			&leave.LeaveType,
			&leave.Status,
			&leave.Reason,
			&from,
			&to,
		)
		if err != nil {
			continue
		}

		leave.From = from.Format("Jan 02, 2006")
		leave.To = to.Format("Jan 02, 2006")
		leave.Days = int(to.Sub(from).Hours()/24) + 1

		leaves = append(leaves, leave)
	}

	c.JSON(http.StatusOK, leaves)
}
