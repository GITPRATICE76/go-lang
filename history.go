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
	UserID int    `json:"user_id"` // 🔥 important
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

	// 🔥 BASE QUERY
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
		CAST(l.to_date AS DATE) <= CAST(GETDATE() AS DATE)
	`

	args := []interface{}{}
	paramIndex := 1

	// 🔥 USER FILTER (KEY FIX)
	if req.UserID != 0 {
		query += " AND l.user_id = @p" + string(rune(paramIndex+48))
		args = append(args, req.UserID)
		paramIndex++
	}

	// 🔍 SEARCH
	if req.Search != "" {
		query += " AND u.name LIKE @p" + string(rune(paramIndex+48))
		args = append(args, "%"+req.Search+"%")
		paramIndex++
	}

	// 📅 DATE FILTER
	if req.Start != "" {
		query += " AND l.to_date >= @p" + string(rune(paramIndex+48))
		args = append(args, req.Start)
		paramIndex++
	}
	if req.End != "" {
		query += " AND l.from_date <= @p" + string(rune(paramIndex+48))
		args = append(args, req.End)
		paramIndex++
	}

	// 📊 STATUS FILTER
	if req.Status != "" && req.Status != "ALL" {
		query += " AND l.status = @p" + string(rune(paramIndex+48))
		args = append(args, req.Status)
		paramIndex++
	}

	// 🔥 PAGINATION
	query += `
	ORDER BY l.from_date DESC
	OFFSET @p` + string(rune(paramIndex+48)) + ` ROWS
	FETCH NEXT @p` + string(rune(paramIndex+49)) + ` ROWS ONLY
	`

	args = append(args, offset, req.Limit)

	rows, err := db.Query(query, args...)
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

	// 🔥 COUNT QUERY
	countQuery := `
	SELECT COUNT(*)
	FROM leaves l
	LEFT JOIN users u ON l.user_id = u.id
	WHERE
		CAST(l.to_date AS DATE) <= CAST(GETDATE() AS DATE)
	`

	countArgs := []interface{}{}
	paramIndex = 1

	if req.UserID != 0 {
		countQuery += " AND l.user_id = @p" + string(rune(paramIndex+48))
		countArgs = append(countArgs, req.UserID)
		paramIndex++
	}

	if req.Search != "" {
		countQuery += " AND u.name LIKE @p" + string(rune(paramIndex+48))
		countArgs = append(countArgs, "%"+req.Search+"%")
		paramIndex++
	}

	if req.Start != "" {
		countQuery += " AND l.to_date >= @p" + string(rune(paramIndex+48))
		countArgs = append(countArgs, req.Start)
		paramIndex++
	}
	if req.End != "" {
		countQuery += " AND l.from_date <= @p" + string(rune(paramIndex+48))
		countArgs = append(countArgs, req.End)
		paramIndex++
	}

	if req.Status != "" && req.Status != "ALL" {
		countQuery += " AND l.status = @p" + string(rune(paramIndex+48))
		countArgs = append(countArgs, req.Status)
		paramIndex++
	}

	var total int
	err = db.QueryRow(countQuery, countArgs...).Scan(&total)
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
