package main

import (

	"database/sql"

	"fmt"
 
	_ "github.com/denisenkom/go-mssqldb"
)
func ConnectDB() (*sql.DB, error) {

	server := "192.168.4.92"

	port := 8181

	user := "DevTeam"

	password := "5m62Ra8A8Q0dZrs"

	database := "SQL_Practice_DB"
 
	connString := fmt.Sprintf(

		"server=%s;user id=%s;password=%s;port=%d;database=%s",

		server, user, password, port, database,

	)
 
	db, err := sql.Open("sqlserver", connString)

	if err != nil {

		return nil, err

	}
 
	return db, nil

}

//  this file is responsible for connecting to database