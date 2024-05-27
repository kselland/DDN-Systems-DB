package main

import (
	"context"
	"ddn/ddn/auth"
	"ddn/ddn/components"
	"ddn/ddn/inventoryItem"
	"ddn/ddn/lib"
	"ddn/ddn/middleware"
	"ddn/ddn/product"
	"ddn/ddn/session"
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
type ErroringRouteWithSession func(s *session.Session, w http.ResponseWriter, r *http.Request) error
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
			components.ErrPage(*requestErr).Render(context.Background(), w)
		}
	}
}

func handleErr(err error, w http.ResponseWriter) {
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
			components.ErrPage(*requestErr).Render(context.Background(), w)
}

func handleErrAndSession(fn ErroringRouteWithSession) Route {
	return func(w http.ResponseWriter, r *http.Request) {
		s, err := session.AuthenticateSession(r)
		if err != nil {
			handleErr(err, w)
			return
		}

		err = fn(s, w, r)

		if err != nil {
			handleErr(err, w)
			return
		}
	}
}

func fourOhFour(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	components.ErrPage(lib.RequestError{
		Message: "Page not found",
		StatusCode: 404,
	}).Render(context.Background(), w)
}

func indexRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/app", 308)
}

//go:embed static
var static embed.FS

func homePage(s *session.Session, w http.ResponseWriter, r *http.Request) error {
	return homePageTemplate(s).Render(context.Background(), w)
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/app", handleErrAndSession(homePage)).Methods("GET")

	r.HandleFunc("/app/products", handleErrAndSession(product.IndexPage)).Methods("GET")
	r.HandleFunc("/app/products/new", handleErrAndSession(product.NewPage)).Methods("GET", "POST")
	r.HandleFunc("/app/product/{id}", handleErrAndSession(product.ViewPage)).Methods("GET", "POST")
	r.HandleFunc("/app/product/{id}/delete", handleErrWith500(product.DeletePage)).Methods("POST")

	r.HandleFunc("/app/storage-locations", handleErrAndSession(storageLocation.IndexPage)).Methods("GET")
	r.HandleFunc("/app/storage-locations/new", handleErrAndSession(storageLocation.NewPage)).Methods("GET", "POST")
	r.HandleFunc("/app/storage-location/{id}", handleErrAndSession(storageLocation.ViewPage)).Methods("GET", "POST")
	r.HandleFunc("/app/storage-location/{id}/delete", handleErrWith500(storageLocation.DeletePage)).Methods("POST")

	r.HandleFunc("/app/inventory", handleErrAndSession(inventoryItem.IndexPage)).Methods("GET")
	r.HandleFunc("/app/inventory/new", handleErrAndSession(inventoryItem.NewPage)).Methods("GET", "POST")
	r.HandleFunc("/app/inventory-item/{id}", handleErrAndSession(inventoryItem.ViewPage)).Methods("GET", "POST")
	r.HandleFunc("/app/inventory-item/{id}/delete", handleErrWith500(inventoryItem.DeletePage)).Methods("POST")

	r.HandleFunc("/login", handleErrWith500(auth.LoginPage)).Methods("GET", "POST")
	r.HandleFunc("/signup", handleErrWith500(auth.SignupPage)).Methods("GET", "POST")
	r.HandleFunc("/logout", handleErrWith500(auth.LogoutPage)).Methods("POST")
	r.HandleFunc("/", indexRedirect).Methods("GET")

	r.PathPrefix("/").HandlerFunc(fourOhFour)

	http.Handle("/static/", http.FileServer(http.FS(static)))
	http.Handle("/", middleware.NewAuthMiddleware(middleware.NewCSRFMiddleware(r)))

	fmt.Println("Listening on port 3001")
	err := http.ListenAndServe(":3001", nil)

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Println("server closed")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
