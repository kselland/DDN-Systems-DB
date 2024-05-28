package middleware

import (
	"ddn/ddn/lib"
	"ddn/ddn/session"
	"net/http"

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
		if err != nil {
			http.SetCookie(w, &http.Cookie{
				Name:  "GOTO_ROUTE",
				Value: r.URL.Path,
				Path:  "/",
				MaxAge: 2000,
			})
			http.Redirect(w, r, "/login", 303)
			return
		}
		session.CreateSession(r, w, s.User_Id)
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
