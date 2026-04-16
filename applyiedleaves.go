package main

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

type LeaveFilter struct {
	Employee string `json:"employee"`
	Search   string `json:"search"`
	Start    string `json:"start"`
	End      string `json:"end"`
	Status   string `json:"status"`
	Page     int    `json:"page"`
	Limit    int    `json:"limit"`
}

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
		c.JSON(500, gin.H{"message": "DB failed"})
		return
	}
	defer db.Close()

	// 🔥 GET USER INFO
	userID := c.GetInt("user_id")
	role := c.GetString("role")

	var team string

	if role != "MANAGER" {
		err = db.QueryRow(`
			SELECT team FROM users WHERE id = @userID
		`, sql.Named("userID", userID)).Scan(&team)

		if err != nil {
			c.JSON(500, gin.H{"message": "Failed to get team"})
			return
		}
	}

	var filter LeaveFilter

	if err := c.BindJSON(&filter); err != nil {
		c.JSON(400, gin.H{"message": "Invalid payload"})
		return
	}

	// ✅ Pagination defaults
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}

	offset := (filter.Page - 1) * filter.Limit

	// 🔥 BASE QUERY
	baseQuery := `
	FROM leaves l
	JOIN users u ON l.user_id = u.id
	WHERE 1=1
	AND l.to_date >= CAST(GETDATE() AS DATE)
	`

	args := []interface{}{}
	paramIndex := 1

	// 🔥 ROLE FILTER (IMPORTANT)
	if role != "MANAGER" {
		baseQuery += " AND u.team = @p" + fmt.Sprint(paramIndex)
		args = append(args, team)
		paramIndex++
	}

	// 🔍 Employee filter
	if filter.Employee != "" {
		baseQuery += " AND u.name LIKE @p" + fmt.Sprint(paramIndex)
		args = append(args, "%"+filter.Employee+"%")
		paramIndex++
	}

	// 🔍 General search
	if filter.Search != "" {
		baseQuery += " AND (l.reason LIKE @p" + fmt.Sprint(paramIndex) +
			" OR l.leave_type LIKE @p" + fmt.Sprint(paramIndex) + ")"
		args = append(args, "%"+filter.Search+"%")
		paramIndex++
	}

	// 📅 Date filter
	if filter.Start != "" && filter.End != "" {
		baseQuery += " AND l.from_date BETWEEN @p" + fmt.Sprint(paramIndex) +
			" AND @p" + fmt.Sprint(paramIndex+1)
		args = append(args, filter.Start, filter.End)
		paramIndex += 2
	}

	// 📊 Status filter
	if filter.Status != "" && filter.Status != "ALL" {
		baseQuery += " AND l.status = @p" + fmt.Sprint(paramIndex)
		args = append(args, filter.Status)
		paramIndex++
	}

	// 🔥 COUNT QUERY
	countQuery := "SELECT COUNT(*) " + baseQuery

	var total int
	err = db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		c.JSON(500, gin.H{"message": err.Error()})
		return
	}

	// 🔥 MAIN QUERY
	mainQuery := `
	SELECT 
		l.id,
		u.name,
		l.leave_type,
		l.status,
		l.reason,
		l.from_date,
		l.to_date
	` + baseQuery + `
	ORDER BY l.created_at DESC
	OFFSET @p` + fmt.Sprint(paramIndex) + ` ROWS
	FETCH NEXT @p` + fmt.Sprint(paramIndex+1) + ` ROWS ONLY
	`

	args = append(args, offset, filter.Limit)

	rows, err := db.Query(mainQuery, args...)
	if err != nil {
		c.JSON(500, gin.H{"message": err.Error()})
		return
	}
	defer rows.Close()

	var leaves []LeaveResponse

	for rows.Next() {
		var leave LeaveResponse
		var from, to time.Time

		rows.Scan(
			&leave.ID,
			&leave.EmployeeName,
			&leave.LeaveType,
			&leave.Status,
			&leave.Reason,
			&from,
			&to,
		)

		leave.From = from.Format("Jan 02, 2006")
		leave.To = to.Format("Jan 02, 2006")
		leave.Days = int(to.Sub(from).Hours()/24) + 1

		leaves = append(leaves, leave)
	}

	c.JSON(200, gin.H{
		"data":  leaves,
		"total": total,
		"page":  filter.Page,
		"limit": filter.Limit,
	})
}
