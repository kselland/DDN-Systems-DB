package storageLocation

import (
	"context"
	"ddn/ddn/db"
	"ddn/ddn/lib"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type EditableStorageLocationProps struct {
	FormStorageLocation db.FormStorageLocation
	Validation          db.StorageLocationValidation
	Id                  *int
}

func IndexPage(s *db.Session, w http.ResponseWriter, r *http.Request) error {
	storageLocations, err := db.GetStorageLocations()
	if err != nil {
		return err
	}

	return indexTemplate(
		s,
		*storageLocations,
	).Render(context.Background(), w)
}

func ViewPage(s *db.Session, w http.ResponseWriter, r *http.Request) error {
	idString := mux.Vars(r)["id"]

	id, err := strconv.Atoi(idString)
	if err != nil {
		return &lib.RequestError{
			Message:    "Not Found",
			StatusCode: 404,
		}
	}


	storageLocation, err := db.GetStorageLocationById(id)
	if storageLocation == nil {
		return &lib.RequestError{
			Message:    "Not Found",
			StatusCode: 404,
		}
	}

	colors, colorsFetchingErr := db.GetColorsFromDb()
	if colorsFetchingErr != nil {
		return colorsFetchingErr
	}

	if r.Method == "POST" {
		formStorageLocation := db.GetFormStorageLocationFromPost(r, &id)

		validation, storageLocation := db.ValidateFormStorageLocation(formStorageLocation, colors)

		if storageLocation == nil {
			return viewTemplate(
				s,
				formStorageLocation,
				validation,
				colors,
			).Render(context.Background(), w)
		}

		err := db.UpdateStorageLocation(id, storageLocation)
		if err != nil {
			return viewTemplate(
				s,
				formStorageLocation,
				db.StorageLocationValidation{
					Root: "Failed to update storage location in DB",
				},
				colors,
			).Render(context.Background(), w)
		}

		http.Redirect(w, r, fmt.Sprintf("/storage-location/%d", id), http.StatusSeeOther)
		return nil
	}

	return viewTemplate(
		s,
		db.StorageLocationToFormStorageLocation(*storageLocation),
		db.StorageLocationValidation{},
		colors,
	).Render(context.Background(), w)
}

func DeletePage(s *db.Session, w http.ResponseWriter, r *http.Request) error {
	idString := mux.Vars(r)["id"]

	id, err := strconv.Atoi(idString)
	if err != nil {
		return &lib.RequestError{
			Message:    "Not Found",
			StatusCode: 404,
		}
	}

	storageLocation, err := db.GetStorageLocationById(id)

	if storageLocation == nil {
		return &lib.RequestError{
			Message:    "Not Found",
			StatusCode: 404,
		}
	}

	err = db.DeleteStorageLocation(id)
	if err != nil {
		return err
	}

	http.Redirect(w, r, "/storage-locations", http.StatusSeeOther)
	return nil
}

func NewPage(s *db.Session, w http.ResponseWriter, r *http.Request) error {
	colors, err := db.GetColorsFromDb()
	if err != nil {
		return err
	}

	if r.Method == "POST" {
		formStorageLocation := db.GetFormStorageLocationFromPost(r, nil)
		validation, storageLocation := db.ValidateFormStorageLocation(formStorageLocation, colors)

		if storageLocation == nil {
			return newTemplate(
				s,
				formStorageLocation,
				validation,
				colors,
			).Render(context.Background(), w)
		}

		err := db.InsertStorageLocation(storageLocation)
		if err != nil {
			return newTemplate(
				s,
				formStorageLocation,
				db.StorageLocationValidation{Root: "Error saving storage location to database"},
				colors,
			).Render(context.Background(), w)
		}

		http.Redirect(w, r, "/storage-locations", http.StatusSeeOther)

		return nil
	}

	return newTemplate(
		s,
		db.FormStorageLocation{},
		db.StorageLocationValidation{},
		colors,
	).Render(context.Background(), w)
}
