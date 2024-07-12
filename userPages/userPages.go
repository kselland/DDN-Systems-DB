package userPages

import (
	"context"
	"ddn/ddn/appPaths"
	"ddn/ddn/db"
	"ddn/ddn/lib"
	"net/http"
)

func IndexPage(s *db.Session, w http.ResponseWriter, r *http.Request) error {
	users, err := db.GetUsers()
	if err != nil {
		return err
	}

	return indexTemplate(
		s,
		*users,
	).Render(context.Background(), w)
}

func ViewPage(s *db.Session, w http.ResponseWriter, r *http.Request) error {
	return editTemplate().Render(context.Background(), w)
}

func NewPage(s *db.Session, w http.ResponseWriter, r *http.Request) error {
	if r.Method == "POST" {
		// TODO: Validate the fields
		r.ParseForm()
		email := r.PostForm.Get("email")
		password := r.PostForm.Get("password")
		roleString := r.PostForm.Get("role")
		name := r.PostForm.Get("name")

		passwordDigest, err := lib.GetDigest(password)
		if err != nil {
			return err
		}

		role := db.ParseUserRole(roleString)
		if role == nil {
			// TODO: Fail validation
		}

		err = db.InsertUser(db.User {
			Email: email,
			Password_digest: passwordDigest,
			Role: *role,
			Name: name,
		})
		if err != nil {
			return err
		}
		appPaths.Redirect(w, r, appPaths.UserListing.WithNoParams(), 303)
		return nil
	}

	roles, err := db.GetRoleOptions()
	if err != nil {
		return err
	}

	return newTemplate(s, NewTemplateDetails{name: "", email: "", emailTaken: false, roles: roles}).Render(context.Background(), w)
}
