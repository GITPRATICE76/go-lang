package main

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type LeaveActionRequest struct {
	UserID  int    `json:"user_id"`   // MANAGER ID
	LeaveID int    `json:"leave_id"`  // leaves.id
	Action  string `json:"action"`    // APPROVED / REJECTED
	Remarks string `json:"remarks"`   // manager remarks
}

func LeaveAction(c *gin.Context) {

	var req LeaveActionRequest

	// 1️⃣ Bind request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request payload",
		})
		return
	}

	// 2️⃣ Validate action
	if req.Action != "APPROVED" && req.Action != "REJECTED" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Action must be APPROVED or REJECTED",
		})
		return
	}

	// 3️⃣ Validate remarks
	if strings.TrimSpace(req.Remarks) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Remarks are required",
		})
		return
	}

	// 4️⃣ DB connection
	db, err := ConnectDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Database connection failed",
		})
		return
	}
	defer db.Close()

	// 5️⃣ Check manager role
	var role string
	err = db.QueryRow(
		`SELECT role FROM users WHERE id = @id`,
		sql.Named("id", req.UserID),
	).Scan(&role)

	if err != nil || role != "MANAGER" {
		c.JSON(http.StatusForbidden, gin.H{
			"message": "Only managers can approve or reject leaves",
		})
		return
	}

	// 6️⃣ UPDATE leave (🔥 NO STATUS RESTRICTION)
	result, err := db.Exec(
		`
		UPDATE leaves
		SET
			status = @status,
			remarks = @remarks,
			approved_by = @approved_by
		WHERE id = @leave_id
		`,
		sql.Named("status", req.Action),
		sql.Named("remarks", req.Remarks),
		sql.Named("approved_by", req.UserID),
		sql.Named("leave_id", req.LeaveID),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to update leave",
		})
		return
	}

	// 7️⃣ Check affected rows
	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Leave not found",
		})
		return
	}

	// 8️⃣ Success
	c.JSON(http.StatusOK, gin.H{
		"message": "Leave " + req.Action + " successfully",
	})
}
