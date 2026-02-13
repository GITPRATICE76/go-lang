package main

import (
	"database/sql"
	"time"

	"github.com/gin-gonic/gin"
)

type DashboardSummary struct {
	HighestLeaveDate struct {
		Date  string `json:"date"`
		Count int    `json:"count"`
	} `json:"highest_leave_date"`

	TeamHighestLeave struct {
		Team  string `json:"team"`
		Count int    `json:"count"`
	} `json:"team_highest_leave"`

	PeakLeaveWeek struct {
		WeekNumber int    `json:"week_number"`
		Start      string `json:"start"`
		End        string `json:"end"`
		Count      int    `json:"count"`
	} `json:"peak_leave_week"`

	TopLeaveTaker struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	} `json:"top_leave_taker"`
}

func GetDashboardSummary(c *gin.Context) {

	db, err := ConnectDB()
	if err != nil {
		c.JSON(500, gin.H{"message": "DB connection failed"})
		return
	}
	defer db.Close()

	year := c.Query("year")
	month := c.Query("month")

	if year == "" || month == "" {
		c.JSON(400, gin.H{"message": "year and month required"})
		return
	}

	startDate := year + "-" + month + "-01"

	// safer way to get end date
	startTime, _ := time.Parse("2006-1-2", startDate)
	endTime := startTime.AddDate(0, 1, -1)

	prevStart := startTime.AddDate(0, -1, 0)
	prevEnd := prevStart.AddDate(0, 1, -1)

	var summary DashboardSummary

	// ================================================
	// 🔴 Highest Leave Date (Current Month)
	// ================================================
	query1 := `
	SELECT TOP 1
		CAST(from_date AS DATE) as leave_date,
		COUNT(*) as total
	FROM leaves
	WHERE status = 'APPROVED'
	AND from_date BETWEEN @p1 AND @p2
	GROUP BY CAST(from_date AS DATE)
	ORDER BY total DESC
	`

	var leaveDate sql.NullTime

	err = db.QueryRow(query1, startTime, endTime).
		Scan(&leaveDate, &summary.HighestLeaveDate.Count)

	if err == nil && leaveDate.Valid {
		summary.HighestLeaveDate.Date =
			leaveDate.Time.Format("Jan 02")
	}

	// ================================================
	// 🟣 Team With Highest Leave (Previous Month)
	// ================================================
	query2 := `
	SELECT TOP 1
		u.team,
		COUNT(*) as total
	FROM leaves l
	JOIN users u ON l.user_id = u.id
	WHERE l.status = 'APPROVED'
	AND l.from_date BETWEEN @p1 AND @p2
	GROUP BY u.team
	ORDER BY total DESC
	`

	db.QueryRow(query2, prevStart, prevEnd).
		Scan(&summary.TeamHighestLeave.Team,
			&summary.TeamHighestLeave.Count)

	// ================================================
	// 🟡 Peak Leave Week (Current Month)
	// ================================================
	query3 := `
	SELECT TOP 1
		DATEPART(WEEK, from_date) as week_number,
		MIN(from_date) as start_date,
		MAX(from_date) as end_date,
		COUNT(*) as total
	FROM leaves
	WHERE status = 'APPROVED'
	AND from_date BETWEEN @p1 AND @p2
	GROUP BY DATEPART(WEEK, from_date)
	ORDER BY total DESC
	`

	var weekStart, weekEnd sql.NullTime

	err = db.QueryRow(query3, startTime, endTime).
		Scan(&summary.PeakLeaveWeek.WeekNumber,
			&weekStart,
			&weekEnd,
			&summary.PeakLeaveWeek.Count)

	if err == nil {
		if weekStart.Valid {
			summary.PeakLeaveWeek.Start =
				weekStart.Time.Format("Jan 02")
		}
		if weekEnd.Valid {
			summary.PeakLeaveWeek.End =
				weekEnd.Time.Format("Jan 02")
		}
	}

	// ================================================
	// 🔵 Top Leave Taker (Previous Month)
	// ================================================
	query4 := `
	SELECT TOP 1
		u.name,
		COUNT(*) as total
	FROM leaves l
	JOIN users u ON l.user_id = u.id
	WHERE l.status = 'APPROVED'
	AND l.from_date BETWEEN @p1 AND @p2
	GROUP BY u.name
	ORDER BY total DESC
	`

	db.QueryRow(query4, prevStart, prevEnd).
		Scan(&summary.TopLeaveTaker.Name,
			&summary.TopLeaveTaker.Count)

	c.JSON(200, summary)
}
