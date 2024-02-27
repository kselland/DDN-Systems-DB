package main

import (
    "fmt"
    "errors"
    "io"
    "net/http"
    "os"
    "html/template"
)

func getRoot(w http.ResponseWriter, r *http.Request) {
    io.WriteString(w, "This is my website")
}

func main() {
    fmt.Println("Hello world")

    http.HandleFunc("/", getRoot)

    err := http.ListenAndServe(":3000", nil)

    if errors.Is(err, http.ErrServerClosed) {
        fmt.Println("server closed")
    } else if err != nil {
        fmt.Printf("error starting server: %s\n", err)
        os.Exit(1)
    }
}
