package main

import (
	"context"
	"ddn/ddn/inventoryItem"
	"ddn/ddn/lib"
	"ddn/ddn/product"
	"ddn/ddn/storageLocation"
	"embed"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type ErroringRoute func(w http.ResponseWriter, r *http.Request) error
type Route func(w http.ResponseWriter, r *http.Request)

func handleErrWith500(fn ErroringRoute) Route {
	return func(w http.ResponseWriter, r *http.Request) {
		err := fn(w, r)

		if err != nil {
			requestErr, ok := err.(*lib.RequestError);
			log.Println(requestErr, ok)
			if !ok {
				log.Println(err)
				requestErr = &lib.RequestError{
					Message: "An Error Occurred",
					StatusCode: 500,
				}
			}

			w.WriteHeader(requestErr.StatusCode)
			errPage(*requestErr).Render(context.Background(), w)
		}
	}
}

func fourOhFour(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	errPage(lib.RequestError{
		Message: "Page not found",
		StatusCode: 404,
	}).Render(context.Background(), w)
}

//go:embed static
var static embed.FS

func homePage(w http.ResponseWriter, r *http.Request) error {
	return homePageTemplate().Render(context.Background(), w)
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", handleErrWith500(homePage)).Methods("GET")

	r.HandleFunc("/products", handleErrWith500(product.IndexPage)).Methods("GET")
	r.HandleFunc("/products/new", handleErrWith500(product.NewPage)).Methods("GET", "POST")
	r.HandleFunc("/product/{id}", handleErrWith500(product.ViewPage)).Methods("GET", "POST")
	r.HandleFunc("/product/{id}/delete", handleErrWith500(product.DeletePage)).Methods("POST")

	r.HandleFunc("/storage-locations", handleErrWith500(storageLocation.IndexPage)).Methods("GET")
	r.HandleFunc("/storage-locations/new", handleErrWith500(storageLocation.NewPage)).Methods("GET", "POST")
	r.HandleFunc("/storage-location/{id}", handleErrWith500(storageLocation.ViewPage)).Methods("GET", "POST")
	r.HandleFunc("/storage-location/{id}/delete", handleErrWith500(storageLocation.DeletePage)).Methods("POST")

	r.HandleFunc("/inventory", handleErrWith500(inventoryItem.IndexPage)).Methods("GET")
	r.HandleFunc("/inventory/new", handleErrWith500(inventoryItem.NewPage)).Methods("GET", "POST")
	r.HandleFunc("/inventory-item/{id}", handleErrWith500(inventoryItem.ViewPage)).Methods("GET", "POST")
	r.HandleFunc("/inventory-item/{id}/delete", handleErrWith500(inventoryItem.DeletePage)).Methods("POST")

	r.PathPrefix("/").HandlerFunc(fourOhFour)

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
