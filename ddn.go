package main

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
)

func get500(w http.ResponseWriter, r *http.Request) {
    // TODO: Make proper error page
    io.WriteString(w, "<h1>There was an error</h1>")
}

func getRoot(w http.ResponseWriter, r *http.Request) error {
    tmpl, err := template.ParseFS(templates, "templates/index.html")
    if err != nil {
        log.Printf("Failed to load template")
        return err
    }

    err = tmpl.Execute(w, 4)
    if err != nil {
        log.Printf("Failed to compile template")
        return err
    }

    return nil;
}

type ErroringRoute func(w http.ResponseWriter, r *http.Request) error
type Route func(w http.ResponseWriter, r *http.Request)
func handleErrWith500(fn ErroringRoute) Route {
    return func(w http.ResponseWriter, r *http.Request) {
        err := fn(w,r)

        if err != nil {
            get500(w,r)
        }
    }
}

//go:embed static
var static embed.FS

//go:embed templates
var templates embed.FS

func main() {
    fmt.Println("Listening on port 3000")

    http.HandleFunc("/", handleErrWith500(getRoot))

    http.Handle("/static/", http.FileServer(http.FS(static)))

    err := http.ListenAndServe(":3000", nil)

    if errors.Is(err, http.ErrServerClosed) {
        fmt.Println("server closed")
    } else if err != nil {
        fmt.Printf("error starting server: %s\n", err)
        os.Exit(1)
    }
}
