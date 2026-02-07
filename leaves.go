package main
 
import (

	"database/sql"

	"net/http"

	"time"
 
	"github.com/gin-gonic/gin"

)
 
type ApplyLeaveRequest struct {

	UserID    int    `json:"user_id"`

	LeaveType string `json:"leave_type"`

	FromDate  string `json:"from_date"` 

	ToDate    string `json:"to_date"`    

	Reason    string `json:"reason"`

}
 
func ApplyLeave(c *gin.Context) {

	var req ApplyLeaveRequest
 
	// 1️⃣ Bind JSON

	if err := c.ShouldBindJSON(&req); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{

			"message": "Invalid request payload",

		})

		return

	}
 
	// 2️⃣ Validate leave type

	validLeaveTypes := map[string]bool{

		"SICK":   true,

		"CASUAL": true,

		"EARNED": true,

	}
 
	if !validLeaveTypes[req.LeaveType] {

		c.JSON(http.StatusBadRequest, gin.H{

			"message": "Invalid leave type",

		})

		return

	}
 
	// 3️⃣ Validate dates

	fromDate, err1 := time.Parse("2006-01-02", req.FromDate)

	toDate, err2 := time.Parse("2006-01-02", req.ToDate)
 
	if err1 != nil || err2 != nil || toDate.Before(fromDate) {

		c.JSON(http.StatusBadRequest, gin.H{

			"message": "Invalid from_date or to_date",

		})

		return

	}
 
	// 4️⃣ Connect DB

	db, err := ConnectDB()

	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{

			"message": "Database connection failed",

		})

		return

	}

	defer db.Close()
 
	// 5️⃣ Check user role (only EMPLOYEE allowed)

	var role string

	err = db.QueryRow(

		"SELECT role FROM users WHERE id = @id",

		sql.Named("id", req.UserID),

	).Scan(&role)
 
	if err != nil || role != "EMPLOYEE" {

		c.JSON(http.StatusForbidden, gin.H{

			"message": "Only employees can apply for leave",

		})

		return

	}
 
	// 6️⃣ Insert leave

	query := `

		INSERT INTO leaves 

		(user_id, leave_type, from_date, to_date, reason, status)

		VALUES 

		(@user_id, @leave_type, @from_date, @to_date, @reason, 'PENDING')

	`
 
	_, err = db.Exec(

		query,

		sql.Named("user_id", req.UserID),

		sql.Named("leave_type", req.LeaveType),

		sql.Named("from_date", fromDate),

		sql.Named("to_date", toDate),

		sql.Named("reason", req.Reason),

	)
 
	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{

			"message": "Failed to apply leave",

		})

		return

	}
 
	// 7️⃣ Success response

	c.JSON(http.StatusOK, gin.H{

		"message": "Leave applied successfully",

	})

}

 