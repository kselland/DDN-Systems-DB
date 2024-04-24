package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"reflect"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func getDb() *sql.DB {
	db_host := os.Getenv("DB_HOST")
	db_port := os.Getenv("DB_PORT")
	db_database := os.Getenv("DB_DATABASE")
	db_username := os.Getenv("DB_USERNAME")
	db_password := os.Getenv("DB_PASSWORD")

	if db_host == "" || db_port == "" || db_database == "" || db_username == "" || db_password == "" {
		log.Fatal("Please ensure the DB_URL, DB_DATABASE, DB_USERNAME, and DB_PASSWORD environment variables are all defined")
	}

	connString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", db_host, db_port, db_username, db_password, db_database)

	DB, err := sql.Open("postgres", connString)
	if err != nil {
		log.Fatal("Failed to connect to DB. Please ensure your credentials are correct", err)
	}

	return DB
}

var Db *sql.DB

func GetTable[T any](rows *sql.Rows) (out []T) {
	var table []T
	for rows.Next() {
		var data T
		s := reflect.ValueOf(&data).Elem()
		numCols := s.NumField()
		columns := make([]interface{}, numCols)

		for i := 0; i < numCols; i++ {
			field := s.Field(i)
			columns[i] = field.Addr().Interface()
		}

		if err := rows.Scan(columns...); err != nil {
			fmt.Println("Case Read Error ", err)
		}

		table = append(table, data)
	}
	return table
}

func init() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Failed to read .env file")
	}

	Db = getDb()

	// TODO: These settings are mostly arbitrary, find good values
	Db.SetMaxIdleConns(10)
	Db.SetMaxOpenConns(10)
	Db.SetConnMaxLifetime(time.Minute * 3)
}
