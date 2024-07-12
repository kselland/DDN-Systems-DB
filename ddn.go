package main

import (
	"bufio"
	"context"
	"ddn/ddn/appPaths"
	"ddn/ddn/auth"
	"ddn/ddn/components"
	"ddn/ddn/db"
	_ "ddn/ddn/dotenv"
	"ddn/ddn/inventoryItem"
	"ddn/ddn/lib"
	"ddn/ddn/middleware"
	"ddn/ddn/product"
	"ddn/ddn/session"
	"ddn/ddn/storageLocation"
	"ddn/ddn/userPages"
	"embed"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type ErroringRoute func(w http.ResponseWriter, r *http.Request) error
type ErroringRouteWithSession func(s *db.Session, w http.ResponseWriter, r *http.Request) error
type Route func(w http.ResponseWriter, r *http.Request)

func handleErr(err error, w http.ResponseWriter) {
	requestErr, ok := err.(*lib.RequestError)
	if !ok {
		requestErr = &lib.RequestError{
			Message:    "An Error Occurred",
			StatusCode: 500,
		}
	}

	fmt.Println(err)

	w.WriteHeader(requestErr.StatusCode)
	components.ErrPage(*requestErr).Render(context.Background(), w)
}

func handleErrAndSession(
	permissions db.Permission,
	fn ErroringRouteWithSession,
) Route {
	return func(w http.ResponseWriter, r *http.Request) {
		s, err := session.AuthenticateSession(r)
		if err != nil {
			handleErr(err, w)
			return
		}

		var u *db.User
		if s != nil {
			u = &s.User
		}

		if !db.UserHasPermission(u, permissions) {
			err := &lib.RequestError{Message: "You don't have permission to perform this action", StatusCode: 401}
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
		Message:    "Page not found",
		StatusCode: 404,
	}).Render(context.Background(), w)
}

func indexRedirect(w http.ResponseWriter, r *http.Request) {
	appPaths.Redirect(w, r, appPaths.Dashboard.WithNoParams(), 308)
}

//go:embed static
var static embed.FS

func homePage(s *db.Session, w http.ResponseWriter, r *http.Request) error {
	return homePageTemplate(s).Render(context.Background(), w)
}

func registerRoute(r *mux.Router, appPath appPaths.AppPath, methods []string, handler ErroringRouteWithSession) {
	r.HandleFunc(string(appPath), handleErrAndSession(appPath.Permissions(), handler))
}

func startServer() {
	portStr := os.Getenv("PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		panic("Invalid PORT variable")
	}

	r := mux.NewRouter()

	registerRoute(r, appPaths.Dashboard, []string{"GET"}, homePage)

	registerRoute(r, appPaths.ProductListing, []string{"GET"}, product.IndexPage)
	registerRoute(r, appPaths.ProductNew, []string{"GET", "POST"}, product.NewPage)
	registerRoute(r, appPaths.Product, []string{"GET", "POST"}, product.ViewPage)
	registerRoute(r, appPaths.ProductDelete, []string{"POST"}, product.DeletePage)

	registerRoute(r, appPaths.StorageLocationListing, []string{"GET"}, storageLocation.IndexPage)
	registerRoute(r, appPaths.StorageLocationNew, []string{"GET", "POST"}, storageLocation.NewPage)
	registerRoute(r, appPaths.StorageLocation, []string{"GET", "POST"}, storageLocation.ViewPage)
	registerRoute(r, appPaths.StorageLocationDelete, []string{"POST"}, storageLocation.DeletePage)

	registerRoute(r, appPaths.Inventory, []string{"GET"}, inventoryItem.IndexPage)
	registerRoute(r, appPaths.InventoryItemNew, []string{"GET", "POST"}, inventoryItem.NewPage)
	registerRoute(r, appPaths.InventoryItem, []string{"GET", "POST"}, inventoryItem.ViewPage)
	registerRoute(r, appPaths.InventoryItemDelete, []string{"POST"}, inventoryItem.DeletePage)
	registerRoute(r, appPaths.InventoryDeduct, []string{"GET", "POST"}, inventoryItem.DeductPage)

	registerRoute(r, appPaths.UserListing, []string{"GET"}, userPages.IndexPage)
	registerRoute(r, appPaths.User, []string{"GET", "POST"}, userPages.ViewPage)
	registerRoute(r, appPaths.UserNew, []string{"GET", "POST"}, userPages.NewPage)

	registerRoute(r, appPaths.Login, []string{"GET", "POST"}, auth.LoginPage)
	registerRoute(r, appPaths.Logout, []string{"POST"}, auth.LogoutPage)

	r.HandleFunc("/", indexRedirect).Methods("GET")
	r.PathPrefix("/").HandlerFunc(fourOhFour)

	http.Handle("/static/", http.FileServer(http.FS(static)))
	http.Handle("/", middleware.NewAuthMiddleware(middleware.NewCSRFMiddleware(r)))

	fmt.Printf("Listening on port %d\n", port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Println("server closed")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}

func initScript() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Name: ")
	name, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal("Failed to read name")
	}
	name = strings.TrimSpace(name)

	fmt.Print("Email: ")
	email, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal("Failed to read email")
	}
	email = strings.TrimSpace(email)

	fmt.Print("Password: ")
	password, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal("Failed to read password")
	}
	password = strings.TrimSpace(password)

	passwordDigest, err := lib.GetDigest(password)
	if err != nil {
		log.Fatal("Failed to get password digest")
	}

	err = db.InsertUser(db.User{
		Name:            name,
		Email:           email,
		Password_digest: passwordDigest,
		Role:            db.UserRoleSuperAdmin,
	})
	if err != nil {
		fmt.Println(err)
		log.Fatal("Failed to create user")
	}

}

func main() {
	args := os.Args[1:]
	if len(args) > 0 && args[0] == "init" {
		initScript()
	} else {
		startServer()
	}
}
