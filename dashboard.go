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

	// ✅ NEW FOR RO
	TeamTotalLeave int `json:"team_total_leave"`

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

	// 🔥 USER INFO
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

	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	var summary DashboardSummary

	// =====================================================
	// 🔴 1. Highest Leave Date
	// =====================================================
	query1 := `
	SELECT TOP 1
		d.date_value,
		COUNT(l.id)
	FROM (
		SELECT DATEADD(DAY, number, @today) as date_value
		FROM master..spt_values
		WHERE type='P'
		AND DATEADD(DAY, number, @today) <= @endDate
	) d
	LEFT JOIN leaves l
		ON l.status='APPROVED'
		AND d.date_value BETWEEN l.from_date AND l.to_date
	LEFT JOIN users u ON l.user_id = u.id
	WHERE (@team IS NULL OR u.team = @team)
	GROUP BY d.date_value
	ORDER BY COUNT(l.id) DESC, d.date_value ASC
	`

	var highestDate sql.NullTime
	var teamParam interface{}

	if role == "MANAGER" {
		teamParam = nil
	} else {
		teamParam = team
	}

	err = db.QueryRow(query1,
		sql.Named("today", today),
		sql.Named("endDate", endDate),
		sql.Named("team", teamParam),
	).Scan(&highestDate, &summary.HighestLeaveDate.Count)

	if err == nil && highestDate.Valid {
		summary.HighestLeaveDate.Date = highestDate.Time.Format("Jan 02")
	}

	// =====================================================
	// 🟣 2. Team Highest Leave (ONLY MANAGER)
	// =====================================================
	if role == "MANAGER" {
		query2 := `
		SELECT TOP 1
			u.team,
			SUM(DATEDIFF(DAY,
				CASE WHEN l.from_date < @startDate THEN @startDate ELSE l.from_date END,
				CASE WHEN l.to_date > @endDate THEN @endDate ELSE l.to_date END
			) + 1)
		FROM leaves l
		JOIN users u ON l.user_id = u.id
		WHERE l.status='APPROVED'
		AND l.from_date <= @endDate
		AND l.to_date >= @startDate
		GROUP BY u.team
		ORDER BY 2 DESC
		`

		db.QueryRow(query2,
			sql.Named("startDate", startDate),
			sql.Named("endDate", endDate),
		).Scan(&summary.TeamHighestLeave.Team,
			&summary.TeamHighestLeave.Count)
	}

	// =====================================================
	// 🟢 2B. Team Total Leave (ONLY RO)
	// =====================================================
	if role != "MANAGER" {
		err = db.QueryRow(`
			SELECT 
				ISNULL(SUM(DATEDIFF(DAY,
					CASE WHEN l.from_date < @startDate THEN @startDate ELSE l.from_date END,
					CASE WHEN l.to_date > @endDate THEN @endDate ELSE l.to_date END
				) + 1), 0)
			FROM leaves l
			JOIN users u ON l.user_id = u.id
			WHERE l.status='APPROVED'
			AND l.from_date <= @endDate
			AND l.to_date >= @startDate
			AND u.team = @team
		`,
			sql.Named("startDate", startDate),
			sql.Named("endDate", endDate),
			sql.Named("team", team),
		).Scan(&summary.TeamTotalLeave)
	}

	// =====================================================
	// 🟡 3. Peak Leave Week
	// =====================================================
	query3 := `
	SELECT TOP 1
		((DAY(d.date_value)-1)/7)+1,
		MIN(d.date_value),
		MAX(d.date_value),
		COUNT(l.id)
	FROM (
		SELECT DATEADD(DAY, number, @startDate) as date_value
		FROM master..spt_values
		WHERE type='P'
		AND DATEADD(DAY, number, @startDate) <= @endDate
	) d
	LEFT JOIN leaves l
		ON l.status='APPROVED'
		AND d.date_value BETWEEN l.from_date AND l.to_date
	LEFT JOIN users u ON l.user_id = u.id
	WHERE (@team IS NULL OR u.team = @team)
	GROUP BY ((DAY(d.date_value)-1)/7)+1
	ORDER BY COUNT(l.id) DESC
	`

	var ws, we sql.NullTime

	err = db.QueryRow(query3,
		sql.Named("startDate", startDate),
		sql.Named("endDate", endDate),
		sql.Named("team", teamParam),
	).Scan(&summary.PeakLeaveWeek.WeekNumber, &ws, &we, &summary.PeakLeaveWeek.Count)

	if err == nil {
		if ws.Valid {
			summary.PeakLeaveWeek.Start = ws.Time.Format("Jan 02")
		}
		if we.Valid {
			summary.PeakLeaveWeek.End = we.Time.Format("Jan 02")
		}
	}

	// =====================================================
	// 🔵 4. Top Leave Taker
	// =====================================================
	query4 := `
	SELECT TOP 1
		u.name,
		SUM(DATEDIFF(DAY,
			CASE WHEN l.from_date < @startDate THEN @startDate ELSE l.from_date END,
			CASE WHEN l.to_date > @endDate THEN @endDate ELSE l.to_date END
		) + 1)
	FROM leaves l
	JOIN users u ON l.user_id = u.id
	WHERE l.status='APPROVED'
	AND l.from_date <= @endDate
	AND l.to_date >= @startDate
	AND (@team IS NULL OR u.team = @team)
	GROUP BY u.name
	ORDER BY 2 DESC
	`

	db.QueryRow(query4,
		sql.Named("startDate", startDate),
		sql.Named("endDate", endDate),
		sql.Named("team", teamParam),
	).Scan(&summary.TopLeaveTaker.Name,
		&summary.TopLeaveTaker.Count)

	c.JSON(200, summary)
}
