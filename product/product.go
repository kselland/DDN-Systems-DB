package product

import (
	"context"
	"ddn/ddn/appPaths"
	"ddn/ddn/db"
	"ddn/ddn/lib"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func getFormProductFromPost(r *http.Request, id *int) db.FormProduct {
	return db.FormProduct{
		Id:           id,
		Name:         r.PostFormValue("name"),
		Product_Type: r.PostFormValue("product_type"),
		Length:       r.PostFormValue("length"),
		Width:        r.PostFormValue("width"),
		Height:       r.PostFormValue("height"),
		Active:       r.PostForm.Has("active"),
		Price:        r.PostFormValue("price"),
		Color_Name:   r.PostFormValue("color_name"),
	}
}

type ProductValidation struct {
	Root         string
	Name         string
	Product_Type string
	Length       string
	Width        string
	Height       string
	Active       string
	Price        string
	Color_Name   string
}

func validateFormProduct(p db.FormProduct, colorProductTypes []db.ColorProductType) (ProductValidation, *db.Product) {
	valid := true
	validation := ProductValidation{}

	if p.Name == "" {
		validation.Name = "Name is required"
		valid = false
	}
	if p.Product_Type != "cabinet" && p.Product_Type != "accessory" {
		validation.Product_Type = "Product type must be \"cabinet\" or \"accessory\""
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
	if p.Price == "" {
		validation.Price = "Price is required"
		valid = false
	}
	if p.Color_Name == "" {
		validation.Color_Name = "Color is required"
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

	if validation.Product_Type == "" {
		productType := db.ProductType(p.Product_Type)

		colorIdValid := false
		for _, colorProductType := range colorProductTypes {
			if colorProductType.Color_Name == p.Color_Name && colorProductType.Product_Type == productType {
				colorIdValid = true
				break
			}
		}

		if !colorIdValid {
			validation.Color_Name = "Color with that id not found or not available for that product type"
			valid = false
		}
	}

	r := regexp.MustCompile(`^(\d+)(\.(\d*))?$`)
	match := r.FindStringSubmatch(p.Price)

	if len(match) == 0 {
		validation.Price = "Please enter a valid price"
		valid = false
	}

	dollars, err := strconv.Atoi(match[1])
	if err != nil {
		// This shouldn't be able to happen because the regex gaurantees it is a number
		panic("Couldn't convert dollars to int")
	}
	price := dollars * 100
	if len(match) == 3 && match[2] != "" {
		cents, err := strconv.Atoi(match[3])
		if err != nil {
			// This shouldn't be able to happen because the regex gaurantees it is a number
			panic("Couldn't convert cents to int")
		}
		price += cents
	}

	if !valid {
		if validation.Root == "" {
			validation.Root = "There are errors with your submission"
		}
		return validation, nil
	}

	return validation, &db.Product{
		Name:         p.Name,
		Product_Type: db.ProductType(p.Product_Type),
		Length:       length,
		Width:        width,
		Height:       height,
		Price_Cents:  price,
		Color_Name:   p.Color_Name,
	}
}

type EditableProductProps struct {
	FormProduct db.FormProduct
	Validation  ProductValidation
	Id          *int
}

func IndexPage(s *db.Session, w http.ResponseWriter, r *http.Request) error {
	products, err := db.GetDisplayableProducts()
	if err != nil {
		return err
	}

	return indexTemplate(
		s,
		*products,
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

	product, err := db.GetProductById(id)
	if err != nil {
		return err
	}

	if product == nil {
		return &lib.RequestError{
			Message:    "Not Found",
			StatusCode: 404,
		}
	}

	colors, colorsFetchingErr := db.GetColorsFromDb()
	if colorsFetchingErr != nil {
		return colorsFetchingErr
	}

	colorProductTypes, colorProductTypesFetchingErr := db.GetColorProductTypesFromDb()
	if colorProductTypesFetchingErr != nil {
		return colorProductTypesFetchingErr
	}

	if r.Method == "POST" {
		formProduct := getFormProductFromPost(r, &id)

		validation, product := validateFormProduct(formProduct, colorProductTypes)

		if product == nil {
			return viewTemplate(
				s,
				formProduct,
				validation,
				colors,
			).Render(context.Background(), w)
		}

		err = db.UpdateProduct(id, product)
		if err != nil {
			return viewTemplate(
				s,
				formProduct,
				ProductValidation{
					Root: "Failed to update product in DB",
				},
				colors,
			).Render(context.Background(), w)
		}

		appPaths.Redirect(w, r, appPaths.Product.WithParams(map[string]string{"id": fmt.Sprint(id)}), http.StatusSeeOther)
		return nil
	}

	return viewTemplate(
		s,
		product.ToFormProduct(),
		ProductValidation{},
		colors,
	).Render(context.Background(), w)
}

func DeletePage(s *db.Session, w http.ResponseWriter, r *http.Request) error {
	idString := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idString)
	if err != nil {
		return err
	}
	err = db.DeleteProduct(id)
	if err != nil {
		return err
	}

	appPaths.Redirect(w, r, appPaths.ProductListing.WithNoParams(), http.StatusSeeOther)
	return nil
}

func NewPage(s *db.Session, w http.ResponseWriter, r *http.Request) error {
	colors, err := db.GetColorsFromDb()
	if err != nil {
		return err
	}

	colorProductTypes, colorProductTypesFetchingErr := db.GetColorProductTypesFromDb()
	if colorProductTypesFetchingErr != nil {
		return colorProductTypesFetchingErr
	}

	if r.Method == "POST" {
		formProduct := getFormProductFromPost(r, nil)
		validation, product := validateFormProduct(formProduct, colorProductTypes)

		if product == nil {
			return newTemplate(
				s,
				formProduct,
				validation,
				colors,
			).Render(context.Background(), w)
		}

		err := db.InsertProduct(product)
		if err != nil {
			return newTemplate(
				s,
				formProduct,
				ProductValidation{Root: "Error saving product to database"},
				colors,
			).Render(context.Background(), w)
		}

		appPaths.Redirect(w, r, appPaths.ProductListing.WithNoParams(), http.StatusSeeOther)

		return nil
	}

	return newTemplate(
		s,
		db.FormProduct{},
		ProductValidation{},
		colors,
	).Render(context.Background(), w)
}
