package main

import (
	"database/sql"


	"github.com/gin-gonic/gin"
)

func GetUserCodeDetails(c *gin.Context) {

	masterId := c.Query("masterId")
 
	db, err := ConnectDB()

	if err != nil {

		c.JSON(500, gin.H{"message": "DB connection failed"})

		return

	}

	defer db.Close()
 
	rows, err := db.Query(`

		SELECT SubCodeID, Description

		FROM tbl_UserCodeDetail1

		WHERE MasterID = @masterId

		ORDER BY DisplayOrder

	`, sql.Named("masterId", masterId))
 
	if err != nil {

		c.JSON(500, gin.H{"message": "Query failed"})

		return

	}
 
	var result []map[string]interface{}
 
	for rows.Next() {

		var subCode, desc string
 
		rows.Scan(&subCode, &desc)
 
		result = append(result, map[string]interface{}{

			"code": subCode,

			"name": desc,

		})

	}
 
	c.JSON(200, result)

}
 