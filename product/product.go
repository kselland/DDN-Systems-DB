package product

import (
	"context"
	"ddn/ddn/color"
	"ddn/ddn/db"
	"ddn/ddn/lib"
	"ddn/ddn/product_type"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type Product struct {
	Id           int
	Name         string
	Product_Type product_type.ProductType
	Length       int
	Width        int
	Height       int
	Active       bool
	Price_Cents  int
	Color_Id     int
	External_Id  int
}

type DisplayableProduct struct {
	Id             int
	Name           string
	Product_Type   product_type.ProductType
	Length         int
	Width          int
	Height         int
	Active         bool
	Price_Cents    int
	Color_Hex_Code string
	External_Id    int
}

type FormProduct struct {
	Id           *int
	Name         string
	Product_Type string
	Length       string
	Width        string
	Height       string
	Active       bool
	Price        string
	Color_Id     string
	External_Id  string
}

func productToFormProduct(p Product) FormProduct {
	return FormProduct{
		Name:         p.Name,
		Product_Type: string(p.Product_Type),
		Length:       strconv.Itoa(p.Length),
		Width:        strconv.Itoa(p.Width),
		Height:       strconv.Itoa(p.Height),
		Active:       p.Active,
		Price:        fmt.Sprintf("%d.%02d", p.Price_Cents/100, p.Price_Cents%100),
		Color_Id:     strconv.Itoa(p.Color_Id),
		External_Id:  strconv.Itoa(p.External_Id),
		Id:           &p.Id,
	}
}

func getFormProductFromPost(r *http.Request, id *int) FormProduct {
	return FormProduct{
		Id:           id,
		Name:         r.PostFormValue("name"),
		Product_Type: r.PostFormValue("product_type"),
		Length:       r.PostFormValue("length"),
		Width:        r.PostFormValue("width"),
		Height:       r.PostFormValue("height"),
		Active:       r.PostForm.Has("active"),
		Price:        r.PostFormValue("price"),
		Color_Id:     r.PostFormValue("color_id"),
		External_Id:  r.PostFormValue("external_id"),
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
	Color_Id     string
	External_Id  string
}

func validateFormProduct(p FormProduct, colors []color.Color) (ProductValidation, *Product) {
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
	if p.Color_Id == "" {
		validation.Color_Id = "Color is required"
		valid = false
	}
	if p.External_Id == "" {
		validation.External_Id = "External id is required"
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

	colorId, colorIdErr := strconv.Atoi(p.Color_Id)
	if colorIdErr != nil {
		validation.Color_Id = "Color id must be an integer"
		valid = false
	} else if validation.Product_Type == "" {
		productType := product_type.ProductType(p.Product_Type)

		colorIdValid := false
		for _, color := range colors {
			if color.Id == colorId {
				if color.Product_Type == productType {
					colorIdValid = true
					break
				} else {
					break
				}
			}
		}

		if !colorIdValid {
			validation.Color_Id = "Color with that id not found or wrong product type"
			valid = false
		}
	}

	externalId, externalIdErr := strconv.Atoi(p.External_Id)
	if externalIdErr != nil {
		validation.External_Id = "External id must be an integer"
		valid = false
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

	return validation, &Product{
		Name:         p.Name,
		Product_Type: product_type.ProductType(p.Product_Type),
		Length:       length,
		Width:        width,
		Height:       height,
		Price_Cents:  price,
		Color_Id:     colorId,
		External_Id:  externalId,
	}
}

type EditableProductProps struct {
	FormProduct FormProduct
	Validation  ProductValidation
	Id          *int
}

func IndexPage(w http.ResponseWriter, r *http.Request) error {
	query, err := db.Db.Query(`
		SELECT
			p.id,
			p.name,
			p.product_type,
			p.length,
			p.width,
			p.height,
			p.active,
			p.price_cents,
			c.hex_code color_hex_code,
			p.external_id
		FROM
			products p
		LEFT JOIN
		    colors c
		ON
			p.color_id = c.id
	`)
	if err != nil {
		return err
	}

	products := db.GetTable[DisplayableProduct](query)

	return indexTemplate(
		products,
	).Render(context.Background(), w)
}

func ViewPage(w http.ResponseWriter, r *http.Request) error {
	idString := mux.Vars(r)["id"]

	id, err := strconv.Atoi(idString)
	if err != nil {
		return &lib.RequestError{
			Message:    "Not Found",
			StatusCode: 404,
		}
	}

	productsQuery, productsErr := db.Db.Query("SELECT * FROM products WHERE id = $1", id)
	if productsErr != nil {
		return err
	}
	products := db.GetTable[Product](productsQuery)
	if len(products) == 0 {
		return &lib.RequestError{
			Message:    "Not Found",
			StatusCode: 404,
		}
	}
	product := products[0]

	colors, colorsFetchingErr := color.GetColorsFromDb()
	if colorsFetchingErr != nil {
		return colorsFetchingErr
	}

	if r.Method == "POST" {
		formProduct := getFormProductFromPost(r, &id)

		validation, product := validateFormProduct(formProduct, colors)
		fmt.Println("formProduct", formProduct, validation, product)

		if product == nil {
			return viewTemplate(
				formProduct,
				validation,
				colors,
			).Render(context.Background(), w)
		}

		_, err := db.Db.Exec(
			`
				UPDATE
					products
				SET 
					name         = $1,
					width        = $2, 
					length       = $3,
					height       = $4,
					active       = $5,
					product_type = $6,
					color_id     = $7,
					external_id  = $8,
					price_cents  = $9
				WHERE
					id = $10
			`,
			product.Name,
			product.Width,
			product.Length,
			product.Height,
			product.Active,
			product.Product_Type,
			product.Color_Id,
			product.External_Id,
			product.Price_Cents,
			id,
		)
		if err != nil {
			return viewTemplate(
				formProduct,
				ProductValidation{
					Root: "Failed to update product in DB",
				},
				colors,
			).Render(context.Background(), w)
		}

		http.Redirect(w, r, fmt.Sprintf("/product/%d", id), http.StatusSeeOther)
		return nil
	}

	return viewTemplate(
		productToFormProduct(product),
		ProductValidation{},
		colors,
	).Render(context.Background(), w)
}

func DeletePage(w http.ResponseWriter, r *http.Request) error {
	id := mux.Vars(r)["id"]

	query, err := db.Db.Query("SELECT * FROM products WHERE id = $1", id)
	if err != nil {
		return err
	}

	products := db.GetTable[Product](query)

	if len(products) == 0 {
		return &lib.RequestError{
			Message:    "Not Found",
			StatusCode: 404,
		}
	}

	_, err2 := db.Db.Exec("DELETE FROM products WHERE id = $1", id)
	if err2 != nil {
		return err2
	}

	http.Redirect(w, r, "/products", http.StatusSeeOther)
	return nil
}

func NewPage(w http.ResponseWriter, r *http.Request) error {
	colors, err := color.GetColorsFromDb()
	if err != nil {
		return err
	}

	if r.Method == "POST" {
		formProduct := getFormProductFromPost(r, nil)
		validation, product := validateFormProduct(formProduct, colors)
		fmt.Println("formProduct", formProduct, validation, product)

		if product == nil {
			return newTemplate(
				formProduct,
				validation,
				colors,
			).Render(context.Background(), w)
		}

		_, err = db.Db.Exec(
			`
				INSERT INTO products (
					name,
					width, 
					length,
					height,
					product_type,
					active,
					price_cents,
					external_id,
					color_id
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			`,
			product.Name,
			product.Width,
			product.Length,
			product.Height,
			product.Product_Type,
			product.Active,
			product.Price_Cents,
			product.External_Id,
			product.Color_Id,
		)
		if err != nil {
			log.Println(err)
			return newTemplate(
				formProduct,
				ProductValidation{Root: "Error saving product to database"},
				colors,
			).Render(context.Background(), w)
		}

		http.Redirect(w, r, "/products", http.StatusSeeOther)

		return nil
	}

	return newTemplate(
		FormProduct{},
		ProductValidation{},
		colors,
	).Render(context.Background(), w)
}
