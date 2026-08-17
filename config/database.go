package config

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func DBconn() (*sql.DB, error) {
	err := godotenv.Load()

	if err != nil {
		log.Println("File .env tidak ditemukan")
	}

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	dsn := user + ":" + password +
		"@tcp(" + host + ":" + port + ")/" + dbname +
		"?parseTime=true"

	db, err := sql.Open("mysql", dsn)
	return db, err
}
