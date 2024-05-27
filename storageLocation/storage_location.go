package storageLocation

import (
	"context"
	"ddn/ddn/color"
	"ddn/ddn/db"
	"ddn/ddn/lib"
	"ddn/ddn/session"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type StorageLocation struct {
	Id     int
	Bin    string
	Length int
	Width  int
	Height int
}

type FormStorageLocation struct {
	Id     *int
	Bin    string
	Length string
	Width  string
	Height string
}

func storageLocationToFormStorageLocation(s StorageLocation) FormStorageLocation {
	return FormStorageLocation{
		Bin:    s.Bin,
		Length: strconv.Itoa(s.Length),
		Width:  strconv.Itoa(s.Width),
		Height: strconv.Itoa(s.Height),
		Id:     &s.Id,
	}
}

func getFormStorageLocationFromPost(r *http.Request, id *int) FormStorageLocation {
	return FormStorageLocation{
		Id:     id,
		Bin:    r.PostFormValue("bin"),
		Length: r.PostFormValue("length"),
		Width:  r.PostFormValue("width"),
		Height: r.PostFormValue("height"),
	}
}

type StorageLocationValidation struct {
	Root   string
	Bin    string
	Length string
	Width  string
	Height string
}

func validateFormStorageLocation(p FormStorageLocation, colors []color.Color) (StorageLocationValidation, *StorageLocation) {
	valid := true
	validation := StorageLocationValidation{}

	if p.Bin == "" {
		validation.Bin = "Bin is required"
		valid = false
	}
	if p.Length == "" {
		validation.Length = "Length is required"
		valid = false
	}
	if p.Width == "" {
		validation.Width = "Width is required"
		valid = false
	}
	if p.Height == "" {
		validation.Height = "Height is required"
		valid = false
	}

	width, widthErr := strconv.Atoi(p.Width)
	if widthErr != nil {
		validation.Width = "Width must be an integer"
		valid = false
	}

	length, lengthErr := strconv.Atoi(p.Length)
	if lengthErr != nil {
		validation.Length = "Length must be an integer"
		valid = false
	}

	height, heightErr := strconv.Atoi(p.Height)
	if heightErr != nil {
		validation.Height = "Height must be an integer"
		valid = false
	}

	if !valid {
		if validation.Root == "" {
			validation.Root = "There are errors with your submission"
		}
		return validation, nil
	}

	return validation, &StorageLocation{
		Bin:    p.Bin,
		Length: length,
		Width:  width,
		Height: height,
	}
}

type EditableStorageLocationProps struct {
	FormStorageLocation FormStorageLocation
	Validation          StorageLocationValidation
	Id                  *int
}

func IndexPage(s *session.Session, w http.ResponseWriter, r *http.Request) error {
	query, err := db.Db.Query("SELECT * FROM storage_locations")
	if err != nil {
		return err
	}

	storageLocations, err := db.GetTable[StorageLocation](query)
	if err != nil {
		return err
	}

	return indexTemplate(
		s,
		storageLocations,
	).Render(context.Background(), w)
}

func ViewPage(s *session.Session, w http.ResponseWriter, r *http.Request) error {
	idString := mux.Vars(r)["id"]

	id, err := strconv.Atoi(idString)
	if err != nil {
		return &lib.RequestError{
			Message:    "Not Found",
			StatusCode: 404,
		}
	}

	storageLocationsQuery, storageLocationsErr := db.Db.Query("SELECT * FROM storage_locations WHERE id = $1", id)
	if storageLocationsErr != nil {
		return err
	}
	storageLocations, err := db.GetTable[StorageLocation](storageLocationsQuery)
	if err != nil {
		return err
	}
	if len(storageLocations) == 0 {
		return &lib.RequestError{
			Message:    "Not Found",
			StatusCode: 404,
		}
	}
	storageLocation := storageLocations[0]

	colors, colorsFetchingErr := color.GetColorsFromDb()
	if colorsFetchingErr != nil {
		return colorsFetchingErr
	}

	if r.Method == "POST" {
		formStorageLocation := getFormStorageLocationFromPost(r, &id)

		validation, storageLocation := validateFormStorageLocation(formStorageLocation, colors)

		if storageLocation == nil {
			return viewTemplate(
				s,
				formStorageLocation,
				validation,
				colors,
			).Render(context.Background(), w)
		}

		_, err := db.Db.Exec(
			`
				UPDATE
					storage_locations
				SET 
					bin    = $1,
					width  = $2, 
					length = $3,
					height = $4
				WHERE
					id = $5
			`,
			storageLocation.Bin,
			storageLocation.Width,
			storageLocation.Length,
			storageLocation.Height,
			id,
		)
		if err != nil {
			log.Println(err)

			return viewTemplate(
				s,
				formStorageLocation,
				StorageLocationValidation{
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
		storageLocationToFormStorageLocation(storageLocation),
		StorageLocationValidation{},
		colors,
	).Render(context.Background(), w)
}

func DeletePage(w http.ResponseWriter, r *http.Request) error {
	id := mux.Vars(r)["id"]

	query, err := db.Db.Query("SELECT * FROM storage_locations WHERE id = $1", id)
	if err != nil {
		return err
	}

	storageLocations, err := db.GetTable[StorageLocation](query)
	if err != nil {
		return err
	}

	if len(storageLocations) == 0 {
		return &lib.RequestError{
			Message:    "Not Found",
			StatusCode: 404,
		}
	}

	_, err2 := db.Db.Exec("DELETE FROM storage_locations WHERE id = $1", id)
	if err2 != nil {
		return err2
	}

	http.Redirect(w, r, "/storage-locations", http.StatusSeeOther)
	return nil
}

func NewPage(s *session.Session, w http.ResponseWriter, r *http.Request) error {
	colors, err := color.GetColorsFromDb()
	if err != nil {
		return err
	}

	if r.Method == "POST" {
		formStorageLocation := getFormStorageLocationFromPost(r, nil)
		validation, storageLocation := validateFormStorageLocation(formStorageLocation, colors)

		if storageLocation == nil {
			return newTemplate(
				s,
				formStorageLocation,
				validation,
				colors,
			).Render(context.Background(), w)
		}

		_, err = db.Db.Exec(
			`
				INSERT INTO storage_locations (
					bin,
					width, 
					length,
					height
				) VALUES ($1, $2, $3, $4)
			`,
			storageLocation.Bin,
			storageLocation.Width,
			storageLocation.Length,
			storageLocation.Height,
		)
		if err != nil {
			log.Println(err)
			return newTemplate(
				s,
				formStorageLocation,
				StorageLocationValidation{Root: "Error saving storage location to database"},
				colors,
			).Render(context.Background(), w)
		}

		http.Redirect(w, r, "/storage-locations", http.StatusSeeOther)

		return nil
	}

	return newTemplate(
		s,
		FormStorageLocation{},
		StorageLocationValidation{},
		colors,
	).Render(context.Background(), w)
}
