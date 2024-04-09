package main

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func getTable[T any](rows *sql.Rows) (out []T) {
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

func get500(w http.ResponseWriter, r *http.Request) {
	// TODO: Make proper error page
	io.WriteString(w, "<h1>There was an error</h1>")
}

func homePage(w http.ResponseWriter, r *http.Request) error {
	return renderPage(w, r, "templates/index.html", nil)
}

func renderPage(w http.ResponseWriter, r *http.Request, page string, data any) error {
	tmpl, err := template.ParseFS(templates, page, "templates/base.html")
	if err != nil {
		log.Printf("Failed to load template")
		return err
	}

	err = tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Printf("Failed to compile template:")
		log.Println(err)
		return err
	}

	return nil
}

type Product struct {
	Id            int
	Name          string
	Category      string
	Length_Inches int
	Width_Inches  int
	Height_Inches int
	Active        bool
	Price         int
	Color         string
}

func productsPage(w http.ResponseWriter, r *http.Request) error {
	query, err := db.Query("SELECT * FROM products")
	if err != nil {
		log.Printf("Failed to construct query")
		log.Println(err)
		return err
	}

	products := getTable[Product](query)

	err = renderPage(w, r, "templates/products.html", products)
	if err != nil {
		return err
	}

	return nil

}

func newProductPage(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "POST" {
		name := r.PostFormValue("name")
		width, widthErr := strconv.Atoi(r.PostFormValue("width_inches"))
		length, lengthErr := strconv.Atoi(r.PostFormValue("length_inches"))
		height, heightErr := strconv.Atoi(r.PostFormValue("height_inches"))

		if widthErr != nil || lengthErr != nil || heightErr != nil || name == "" {
			return renderPage(w, r, "templates/new-product.html", "You have errors in the details of your submission")
		}

		_, err := db.Exec(
			`
				INSERT INTO products (
					name,
					width_inches, 
					length_inches,
					height_inches,
					category,
					active,
					price,
					color
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			`,
			name,
			width,
			length,
			height,
			"cabinets",
			true,
			12_00,
			"blue",
		)
		if err != nil {
			log.Println(err)
			return renderPage(w, r, "templates/new-product.html", "Error inserting product into the database")
		}

		http.Redirect(w, r, "/products", http.StatusSeeOther)

		return nil
	}
	return renderPage(w, r, "templates/new-product.html", nil)
}

type ErroringRoute func(w http.ResponseWriter, r *http.Request) error
type Route func(w http.ResponseWriter, r *http.Request)

func handleErrWith500(fn ErroringRoute) Route {
	return func(w http.ResponseWriter, r *http.Request) {
		err := fn(w, r)

		if err != nil {
			get500(w, r)
		}
	}
}

//go:embed static
var static embed.FS

//go:embed templates
var templates embed.FS

// TODO: Implement db connection pooling
var db *sql.DB

func init() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Failed to read .env file")
	}

	db = get_db()
}

func get_db() *sql.DB {
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

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", handleErrWith500(homePage)).Methods("Get")
	r.HandleFunc("/products", handleErrWith500(productsPage)).Methods("GET")
	r.HandleFunc("/products/new", handleErrWith500(newProductPage)).Methods("GET", "POST")

	http.Handle("/static/", http.FileServer(http.FS(static)))
	http.Handle("/", r)

	fmt.Println("Listening on port 3000")
	err := http.ListenAndServe(":3000", nil)

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Println("server closed")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
