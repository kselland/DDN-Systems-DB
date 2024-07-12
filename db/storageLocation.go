package db

import (
	"net/http"
	"strconv"
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

func GetStorageLocationOptions() (*[]Option, error) {
	query, err := db.Query(`
		SELECT
			id AS value,
			bin AS text
		FROM
			storage_locations
	`)
	if err != nil {
		return nil, err
	}
	storageLocationOptions, err := getTable[Option](query)
	if err != nil {
		return nil, err
	}

	return &storageLocationOptions, nil
}

func GetStorageLocations() (*[]StorageLocation, error) {
	query, err := db.Query("SELECT * FROM storage_locations")
	if err != nil {
		return nil, err
	}

	storageLocations, err := getTable[StorageLocation](query)

	return &storageLocations, err
}

func StorageLocationToFormStorageLocation(s StorageLocation) FormStorageLocation {
	return FormStorageLocation{
		Bin:    s.Bin,
		Length: strconv.Itoa(s.Length),
		Width:  strconv.Itoa(s.Width),
		Height: strconv.Itoa(s.Height),
		Id:     &s.Id,
	}
}

func GetFormStorageLocationFromPost(r *http.Request, id *int) FormStorageLocation {
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

func ValidateFormStorageLocation(p FormStorageLocation, colors []Color) (StorageLocationValidation, *StorageLocation) {
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

func GetStorageLocationById(id int) (*StorageLocation, error) {
	storageLocationsQuery, err := db.Query("SELECT * FROM storage_locations WHERE id = $1", id)
	if err != nil {
		return nil, err
	}

	storageLocations, err := getTable[StorageLocation](storageLocationsQuery)
	if err != nil {
		return nil, err
	}

	if len(storageLocations) == 0 {
		return nil, nil
	}

	return &storageLocations[0], nil
}

func InsertStorageLocation(s *StorageLocation) error {
	_, err := db.Exec(
		`
			INSERT INTO storage_locations (
				bin,
				width, 
				length,
				height
			) VALUES ($1, $2, $3, $4)
		`,
		s.Bin,
		s.Width,
		s.Length,
		s.Height,
	)
	return err
}

func DeleteStorageLocation(id int) error {
	_, err := db.Exec("DELETE FROM storage_locations WHERE id = $1", id)
	return err
}

func UpdateStorageLocation(id int, s *StorageLocation) error {
	_, err := db.Exec(
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
		s.Bin,
		s.Width,
		s.Length,
		s.Height,
		id,
	)

	return err
}
