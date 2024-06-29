package auth

import (
	"bytes"
	"context"
	"ddn/ddn/db"
	"ddn/ddn/lib"
	"ddn/ddn/session"
	"net/http"
)

func LoginPage(s *db.Session, w http.ResponseWriter, r *http.Request) error {
	if r.Method == "POST" {
		email := r.FormValue("email")
		password := r.FormValue("password")
		passwordDigest, err := lib.GetDigest(password)
		if err != nil {
			return err
		}

		authData, err := db.GetAuthDataByUserId(email)
		if err != nil {
			return err
		}

		if authData == nil || !bytes.Equal(passwordDigest, authData.Password_Digest) {
			return loginPageTemplate(LoginPageTemplateDetails{failed: true, email: email}).Render(context.Background(), w)
		}

		err = session.CreateSession(r, w, authData.Id)
		if err != nil {
			return err
		}

		gotoRouteCookie, err := r.Cookie("GOTO_ROUTE")

		var gotoRoute string
		if err != nil || gotoRouteCookie.Value == "" {
			gotoRoute = "/app"
		} else {
			gotoRoute = gotoRouteCookie.Value
		}

		http.Redirect(w, r, gotoRoute, http.StatusSeeOther)
		return nil
	}

	return loginPageTemplate(LoginPageTemplateDetails{email: "", failed: false}).Render(context.Background(), w)
}

func LogoutPage(s *db.Session, w http.ResponseWriter, r *http.Request) error {
	session.EndSession(w)
	http.Redirect(w, r, "/login", 303)
	return nil
}

