package config

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

func DBconn() (*sql.DB, error) {
	dbDriver := "mysql"
	dbUser := "root"
	dbPass := ""
	dbName := "datamahasiswa"

	db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@/"+dbName)
	return db, err
}
