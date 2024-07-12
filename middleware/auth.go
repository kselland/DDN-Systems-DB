package middleware

import (
	"ddn/ddn/appPaths"
	"ddn/ddn/lib"
	"ddn/ddn/session"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/sessions"
)

type User struct {
	email string
}

type AuthMiddleware struct {
	handler http.Handler
}

func (mw AuthMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if lib.IsAuthenticatedPath(r.URL.Path) {
		s, err := session.AuthenticateSession(r)
		if err != nil || s == nil {
			fmt.Println(err)
			if strings.HasPrefix(r.URL.Path, "/app") {
				http.SetCookie(w, &http.Cookie{
					Name:   "GOTO_ROUTE",
					Value:  r.URL.Path,
					Path:   "/",
					MaxAge: 2000,
				})
			}
			appPaths.Redirect(w, r, appPaths.Login.WithNoParams(), 303)
			return
		}
		session.CreateSession(r, w, s.User.Id)
	}

	mw.handler.ServeHTTP(w, r)
}

func getUser(session *sessions.Session) *User {
	switch user := session.Values["user"].(type) {
	case User:
		return &user
	}

	return nil
}

func NewAuthMiddleware(handler http.Handler) AuthMiddleware {
	return AuthMiddleware{
		handler: handler,
	}
}
