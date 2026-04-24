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

	UserID int    `json:"user_id"` // employee id

	IsManager bool `json:"is_manager"` // 🔥 NEW FLAG

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
 
	// 🔥 user filter logic

	userID := req.UserID

	if req.IsManager {

		userID = 0 // ignore filter → show all

	}
 
	query := `

	SELECT

		l.id,

		l.user_id,

		u.name AS employee_name,

		u.team,

		u.department,

		l.from_date,

		l.to_date,

		l.leave_type,

		l.reason,

		l.status,

		l.created_at,

		DATEDIFF(day, l.from_date, l.to_date) + 1 AS days

	FROM leaves l

	JOIN users u ON l.user_id = u.id

	WHERE

		l.status IN ('APPROVED', 'REJECTED')
 
		-- ✅ only past leaves

		AND l.to_date <= CAST(GETDATE() AS DATE)
 
		-- ✅ role-based filter

		AND (@p1 = 0 OR l.user_id = @p1)
 
		-- filters

		AND (@p2 = '' OR u.name LIKE '%' + @p2 + '%')

		AND (@p3 = '' OR l.from_date >= @p3)

		AND (@p4 = '' OR l.from_date <= @p4)

		AND (@p5 = '' OR @p5 = 'ALL' OR l.status = @p5)
 
	ORDER BY l.from_date DESC

	OFFSET @p6 ROWS FETCH NEXT @p7 ROWS ONLY

	`
 
	rows, err := db.Query(

		query,

		userID,     // p1

		req.Search, // p2

		req.Start,  // p3

		req.End,    // p4

		req.Status, // p5

		offset,     // p6

		req.Limit,  // p7

	)
 
	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return

	}

	defer rows.Close()
 
	var results []gin.H
 
	for rows.Next() {
 
		var (

			id, userId, days int

			name, team, dept string

			fromDate, toDate string

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

	JOIN users u ON l.user_id = u.id

	WHERE

		l.status IN ('APPROVED', 'REJECTED')

		AND l.to_date <= CAST(GETDATE() AS DATE)
 
		AND (@p1 = 0 OR l.user_id = @p1)
 
		AND (@p2 = '' OR u.name LIKE '%' + @p2 + '%')

		AND (@p3 = '' OR l.from_date >= @p3)

		AND (@p4 = '' OR l.from_date <= @p4)

		AND (@p5 = '' OR @p5 = 'ALL' OR l.status = @p5)

	`
 
	var total int
 
	err = db.QueryRow(

		countQuery,

		userID,

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
 