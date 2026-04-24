package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type LeaveHistoryRequest struct {
	Page   int    `json:"page"`
	Limit  int    `json:"limit"`
	Start  string `json:"start"`
	End    string `json:"end"`
	Search string `json:"search"`
	Status string `json:"status"`
}

func GetLeaveHistory(c *gin.Context) {

	var req LeaveHistoryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// defaults
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Limit == 0 {
		req.Limit = 10
	}

	offset := (req.Page - 1) * req.Limit

	db, err := ConnectDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB connection failed"})
		return
	}
	defer db.Close()

	query := `
	SELECT
		l.id,
		l.user_id,
		ISNULL(u.name, 'Unknown') AS employee_name,
		ISNULL(u.team, '-') AS team,
		ISNULL(u.department, '-') AS department,
		l.from_date,
		l.to_date,
		l.leave_type,
		l.reason,
		l.status,
		l.created_at,
		DATEDIFF(day, l.from_date, l.to_date) + 1 AS days
	FROM leaves l
	LEFT JOIN users u ON l.user_id = u.id
	WHERE
		-- ✅ Only past + today
		CAST(l.to_date AS DATE) <= CAST(GETDATE() AS DATE)

		-- ❌ Remove PENDING
		AND l.status IN ('APPROVED', 'REJECTED', 'PENDING')

		-- 🔍 Search
		AND (@p1 = '' OR u.name LIKE '%' + @p1 + '%')

		-- 📅 Correct date range
		AND (@p2 = '' OR l.to_date >= @p2)
		AND (@p3 = '' OR l.from_date <= @p3)

		-- 🎯 Optional status filter (still works for APPROVED/REJECTED)
		AND (@p4 = '' OR @p4 = 'ALL' OR l.status = @p4)

	ORDER BY l.from_date DESC
	OFFSET @p5 ROWS FETCH NEXT @p6 ROWS ONLY
	`

	rows, err := db.Query(
		query,
		req.Search,
		req.Start,
		req.End,
		req.Status,
		offset,
		req.Limit,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var results []gin.H

	for rows.Next() {

		var (
			id, userId, days                     int
			name, team, dept                     string
			fromDate, toDate                     string
			leaveType, reason, status, createdAt string
		)

		err := rows.Scan(
			&id,
			&userId,
			&name,
			&team,
			&dept,
			&fromDate,
			&toDate,
			&leaveType,
			&reason,
			&status,
			&createdAt,
			&days,
		)

		if err != nil {
			continue
		}

		results = append(results, gin.H{
			"id":            id,
			"user_id":       userId,
			"employee_name": name,
			"team":          team,
			"department":    dept,
			"from_date":     fromDate,
			"to_date":       toDate,
			"leave_type":    leaveType,
			"reason":        reason,
			"status":        status,
			"created_at":    createdAt,
			"days":          days,
		})
	}

	// 🔢 COUNT QUERY
	countQuery := `
	SELECT COUNT(*)
	FROM leaves l
	LEFT JOIN users u ON l.user_id = u.id
	WHERE
		CAST(l.to_date AS DATE) <= CAST(GETDATE() AS DATE)

		AND l.status IN ('APPROVED', 'REJECTED')

		AND (@p1 = '' OR u.name LIKE '%' + @p1 + '%')
		AND (@p2 = '' OR l.to_date >= @p2)
		AND (@p3 = '' OR l.from_date <= @p3)
		AND (@p4 = '' OR @p4 = 'ALL' OR l.status = @p4)
	`

	var total int

	err = db.QueryRow(
		countQuery,
		req.Search,
		req.Start,
		req.End,
		req.Status,
	).Scan(&total)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  results,
		"total": total,
		"page":  req.Page,
		"limit": req.Limit,
	})
}
