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
 
    now := time.Now()
 
    // ✅ Current Month Range
    startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
    endDate := startDate.AddDate(0, 1, -1)
 
    var summary DashboardSummary
 
    // =====================================================
    // 🔴 1. Highest Leave Date (overlapping days counted)
    // =====================================================
    query1 := `
    SELECT TOP 1
        d.date_value,
        COUNT(l.id) as total
    FROM (
        SELECT DATEADD(DAY, number, @startDate) as date_value
        FROM master..spt_values
        WHERE type='P'
        AND DATEADD(DAY, number, @startDate) <= @endDate
    ) d
    LEFT JOIN leaves l
        ON l.status='APPROVED'
        AND d.date_value BETWEEN l.from_date AND l.to_date
    GROUP BY d.date_value
    ORDER BY COUNT(l.id) DESC, d.date_value ASC
    `
 
    var highestDate sql.NullTime
 
    err = db.QueryRow(query1,
        sql.Named("startDate", startDate),
        sql.Named("endDate", endDate)).
        Scan(&highestDate, &summary.HighestLeaveDate.Count)
 
    if err == nil && highestDate.Valid {
        summary.HighestLeaveDate.Date = highestDate.Time.Format("Jan 02")
    }
 
    // =====================================================
    // 🟣 2. Team With Highest Leave (total leave days)
    // =====================================================
    query2 := `
    SELECT TOP 1
        u.team,
        SUM(DATEDIFF(DAY,
            CASE WHEN l.from_date < @startDate THEN @startDate ELSE l.from_date END,
            CASE WHEN l.to_date > @endDate THEN @endDate ELSE l.to_date END
        ) + 1) as total_days
    FROM leaves l
    JOIN users u ON l.user_id = u.id
    WHERE l.status='APPROVED'
    AND l.from_date <= @endDate
    AND l.to_date >= @startDate
    GROUP BY u.team
    ORDER BY total_days DESC, u.team ASC
    `
 
    db.QueryRow(query2,
        sql.Named("startDate", startDate),
        sql.Named("endDate", endDate)).
        Scan(&summary.TeamHighestLeave.Team,
            &summary.TeamHighestLeave.Count)
 
    // =====================================================
    // 🟡 3. Peak Leave Week (Week 1–4 of Month)
    // =====================================================
    query3 := `
    SELECT TOP 1
        ((DAY(d.date_value)-1)/7)+1 as week_number,
        MIN(d.date_value),
        MAX(d.date_value),
        COUNT(l.id) as total
    FROM (
        SELECT DATEADD(DAY, number, @startDate) as date_value
        FROM master..spt_values
        WHERE type='P'
        AND DATEADD(DAY, number, @startDate) <= @endDate
    ) d
    LEFT JOIN leaves l
        ON l.status='APPROVED'
        AND d.date_value BETWEEN l.from_date AND l.to_date
    GROUP BY ((DAY(d.date_value)-1)/7)+1
    ORDER BY COUNT(l.id) DESC, MIN(d.date_value) ASC
    `
 
    var weekStart, weekEnd sql.NullTime
 
    err = db.QueryRow(query3,
        sql.Named("startDate", startDate),
        sql.Named("endDate", endDate)).
        Scan(&summary.PeakLeaveWeek.WeekNumber,
            &weekStart,
            &weekEnd,
            &summary.PeakLeaveWeek.Count)
 
    if err == nil {
        if weekStart.Valid {
            summary.PeakLeaveWeek.Start = weekStart.Time.Format("Jan 02")
        }
        if weekEnd.Valid {
            summary.PeakLeaveWeek.End = weekEnd.Time.Format("Jan 02")
        }
    }
 
    // =====================================================
    // 🔵 4. Top Leave Taker (actual leave days counted)
    // =====================================================
    query4 := `
    SELECT TOP 1
        u.name,
        SUM(DATEDIFF(DAY,
            CASE WHEN l.from_date < @startDate THEN @startDate ELSE l.from_date END,
            CASE WHEN l.to_date > @endDate THEN @endDate ELSE l.to_date END
        ) + 1) as total_days
    FROM leaves l
    JOIN users u ON l.user_id = u.id
    WHERE l.status='APPROVED'
    AND l.from_date <= @endDate
    AND l.to_date >= @startDate
    GROUP BY u.name
    ORDER BY total_days DESC, u.name ASC
    `
 
    db.QueryRow(query4,
        sql.Named("startDate", startDate),
        sql.Named("endDate", endDate)).
        Scan(&summary.TopLeaveTaker.Name,
            &summary.TopLeaveTaker.Count)
 
    c.JSON(200, summary)
}
 
 