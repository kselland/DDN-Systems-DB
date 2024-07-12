package db

import (
	"database/sql"
	_ "ddn/ddn/dotenv"
	"fmt"
	"log"
	"os"
	"reflect"
	"time"

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

var db *sql.DB

func getTable[T any](rows *sql.Rows) (out []T, err error) {
	var table []T = []T{}

	for rows.Next() {
		var data T
		s := reflect.ValueOf(&data).Elem()
		numCols := s.NumField()
		columns := make([]interface{}, numCols)

		for i := 0; i < numCols; i++ {
			field := s.Field(i)
			// Special handling for byte slices
			if field.Kind() == reflect.Slice && field.Type().Elem().Kind() == reflect.Uint8 {
				var byteData []byte
				columns[i] = &byteData
				field.Set(reflect.ValueOf(byteData))
			} else {
				columns[i] = field.Addr().Interface()
			}
		}

		if err := rows.Scan(columns...); err != nil {
			return nil, fmt.Errorf("case read error: %w", err)
		}

		for i := 0; i < numCols; i++ {
			field := s.Field(i)
			if field.Kind() == reflect.Slice && field.Type().Elem().Kind() == reflect.Uint8 {
				field.Set(reflect.ValueOf(*columns[i].(*[]byte)))
			}
		}

		table = append(table, data)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return table, nil
}

func getFirst[T any](rows *sql.Rows) (out *T, err error) {
	table, err := getTable[T](rows)
	if err != nil {
		return nil, err
	}

	if len(table) == 0 {
		return nil, nil
	}

	return &table[0], nil
}

func init() {
	db = getDb()

	// TODO: These settings are mostly arbitrary, find good values
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(10)
	db.SetConnMaxLifetime(time.Minute * 3)
}

type Option struct {
	Value string
	Text  string
}

