package main

import (
	"database/sql"

	"github.com/gin-gonic/gin"
)

func GetHolidays(c *gin.Context) {

	db, err := ConnectDB()
	if err != nil {
		c.JSON(500, gin.H{"message": "DB connection failed"})
		return
	}
	defer db.Close()

	userID := c.GetInt("user_id")
	role := c.GetString("role")

	if userID == 0 {
		c.JSON(401, gin.H{"message": "Invalid user"})
		return
	}

	year := c.Query("year")
	month := c.Query("month")

	if year == "" || month == "" {
		c.JSON(400, gin.H{"message": "Year and month are required"})
		return
	}

	// =========================
	// 🔹 Get Holidays
	// =========================
	holidayRows, err := db.Query(`
		SELECT id, name, holiday_date
		FROM holidays
		WHERE YEAR(holiday_date) = @year
		AND MONTH(holiday_date) = @month
	`,
		sql.Named("year", year),
		sql.Named("month", month),
	)

	if err != nil {
		c.JSON(500, gin.H{"message": err.Error()})
		return
	}
	defer holidayRows.Close()

	var holidays []gin.H

	for holidayRows.Next() {
		var id int
		var name, date string
		holidayRows.Scan(&id, &name, &date)

		holidays = append(holidays, gin.H{
			"id":   id,
			"name": name,
			"date": date,
			"type": "HOLIDAY",
		})
	}

	// =========================
	// 🔹 Get Leaves Based On Role
	// =========================

	var leaveRows *sql.Rows

	if role == "MANAGER" {

		leaveRows, err = db.Query(`
			SELECT u.id, u.name, l.leave_type, l.from_date, l.to_date
			FROM leaves l
			JOIN users u ON l.user_id = u.id
			WHERE l.status = 'APPROVED'
			AND l.from_date <= EOMONTH(DATEFROMPARTS(@year,@month,1))
            AND l.to_date >= DATEFROMPARTS(@year,@month,1)
		`,
			sql.Named("year", year),
			sql.Named("month", month),
		)

	} else {

		// 👤 Employee → get team
		var team string
		err = db.QueryRow(`
			SELECT team FROM users WHERE id = @userID
		`,
			sql.Named("userID", userID),
		).Scan(&team)

		if err != nil {
			c.JSON(500, gin.H{"message": "Failed to get team"})
			return
		}

		leaveRows, err = db.Query(`
			SELECT u.id, u.name, l.leave_type, l.from_date, l.to_date
			FROM leaves l
			JOIN users u ON l.user_id = u.id
			WHERE u.team = @team
			AND l.status = 'APPROVED'
			AND YEAR(l.from_date) = @year
			AND MONTH(l.from_date) = @month
		`,
			sql.Named("team", team),
			sql.Named("year", year),
			sql.Named("month", month),
		)
	}

	if err != nil {
		c.JSON(500, gin.H{"message": err.Error()})
		return
	}
	defer leaveRows.Close()

	var leaves []gin.H

	for leaveRows.Next() {
		var id int
		var name, leaveType, fromDate, toDate string
		leaveRows.Scan(&id, &name, &leaveType, &fromDate, &toDate)

		leaves = append(leaves, gin.H{
			"user_id":    id,
			"name":       name,
			"leave_type": leaveType,
			"from_date":  fromDate,
			"to_date":    toDate,
		})
	}

	c.JSON(200, gin.H{
		"year":     year,
		"month":    month,
		"holidays": holidays,
		"leaves":   leaves,
	})
}
