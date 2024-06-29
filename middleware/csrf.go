package middleware

import (
	"context"
	"ddn/ddn/components"
	"ddn/ddn/db"
	"ddn/ddn/lib"
	"ddn/ddn/session"
	"net/http"

	"github.com/jaevor/go-nanoid"
)

type CSRFMiddleware struct {
	handler http.Handler
}

func (mw CSRFMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" && !validateCSRF(r) && lib.IsAuthenticatedPath(r.URL.Path) {
		w.WriteHeader(403)
		components.ErrPage(lib.RequestError{Message: "Invalid CSRF token", StatusCode: 403}).Render(context.Background(), w)
		return
	}
	mw.handler.ServeHTTP(w, r)
}

func NewCSRFMiddleware(handler http.Handler) CSRFMiddleware {
	return CSRFMiddleware{
		handler: handler,
	}
}

var genId func() string

func init() {
	var err error
	genId, err = nanoid.Standard(21)

	if err != nil {
		panic(err)
	}
}

func validateCSRF(r *http.Request) bool {
	session, err := session.AuthenticateSession(r)
	if err != nil || session == nil {
		return false
	}

	csrf_token := r.FormValue("csrf_token")

	return session.Csrf_Token == csrf_token
}

func generateCSRF(userId int) (string, error) {
	id := genId()
	err := db.InsertCsrfToken(db.CsrfToken {
		Token: id,
		User_Id: userId,
	})
	if err != nil {
		return "", err
	}

	return id, nil
}
