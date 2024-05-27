package auth

import (
	"bytes"
	"context"
	"ddn/ddn/db"
	"ddn/ddn/lib"
	"ddn/ddn/session"
	"encoding/hex"
	"log"
	"net/http"
)

func LoginPage(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "POST" {
		email := r.FormValue("email")
		password := r.FormValue("password")
		passwordDigest, err := lib.GetDigest(password)
		if err != nil {
			return err
		}

		query, err := db.Db.Query(`
            SELECT
                id,
                password_digest
            FROM
                users
            WHERE
                email = $1
        `, email)
		if err != nil {
			return err
		}
		res, err := db.GetFirst[struct {
			Id              int
			Password_Digest []byte
		}](query)
		if err != nil {
			return err
		}
		if res == nil || !bytes.Equal(passwordDigest, res.Password_Digest) {
			return loginPageTemplate(LoginPageTemplateDetails{failed: true, email: email}).Render(context.Background(), w)
		}

		err = session.CreateSession(w, res.Id)
		if err != nil {
			return err
		}

		log.Println("Made it here")

		gotoRouteCookie, err := r.Cookie("GOTO_ROUTE")

		var gotoRoute string
		if err != nil {
			gotoRoute = "/app"
		} else {
			gotoRoute = gotoRouteCookie.Value
		}

		http.Redirect(w, r, gotoRoute, 303)
		return nil
	}

	return loginPageTemplate(LoginPageTemplateDetails{email: "", failed: false}).Render(context.Background(), w)
}

func LogoutPage(w http.ResponseWriter, r *http.Request) error {
	session.EndSession(w)
	http.Redirect(w, r, "/login", 303)
	return nil
}

func SignupPage(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "POST" {
		email := r.FormValue("email")
		password := r.FormValue("password")
		passwordDigest, err := lib.GetDigest(password)
		if err != nil {
			return err
		}
		log.Println(password, passwordDigest)

		_, err = db.Db.Exec(`
            INSERT INTO
                users
            (
                email,
                password_digest
            ) VALUES (
                $1,
                decode($2, 'hex')
            )
            
        `, email, hex.EncodeToString(passwordDigest))
		if err != nil {
			return err
		}
		http.Redirect(w, r, "/login", 303)
		return nil
	}
	return loginPageTemplate(LoginPageTemplateDetails{email: "", failed: false}).Render(context.Background(), w)
}
