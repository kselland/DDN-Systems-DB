package db

import (
	"fmt"
	"strconv"
)

type Product struct {
	Id           int
	Name         string
	Product_Type ProductType
	Length       int
	Width        int
	Height       int
	Active       bool
	Price_Cents  int
	Color_Name   string
}

func (p *Product) ToFormProduct() FormProduct {
	return FormProduct{
		Name:         p.Name,
		Product_Type: string(p.Product_Type),
		Length:       strconv.Itoa(p.Length),
		Width:        strconv.Itoa(p.Width),
		Height:       strconv.Itoa(p.Height),
		Active:       p.Active,
		Price:        fmt.Sprintf("%d.%02d", p.Price_Cents/100, p.Price_Cents%100),
		Color_Name:   p.Color_Name,
		Id:           &p.Id,
	}
}

type DisplayableProduct struct {
	Id             int
	Name           string
	Product_Type   ProductType
	Length         int
	Width          int
	Height         int
	Active         bool
	Price_Cents    int
	Color_Name     string
	Color_Hex_Code string
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
	Color_Name   string
}

func GetProductOptions() ([]Option, error) {
	query, err := db.Query(`
		SELECT
			id AS value,
			name AS text
		FROM
			products
	`)
	if err != nil {
		return nil, err
	}
	return getTable[Option](query)
}

func GetProductById(id int) (*Product, error) {
	query, err := db.Query("SELECT * FROM products WHERE id = $1", id)
	if err != nil {
		return nil, err
	}

	product, err := getFirst[Product](query)
	if err != nil {
		return nil, err
	}

	return product, nil
}

func GetDisplayableProducts() (*[]DisplayableProduct, error) {
	query, err := db.Query(`
		SELECT
			p.id,
			p.name,
			p.product_type,
			p.length,
			p.width,
			p.height,
			p.active,
			p.price_cents,
			p.color_name,
			c.hex_code AS color_hex_code
		FROM
			products p
		LEFT JOIN
		    colors c
		ON
			p.color_name = c.name
	`)
	if err != nil {
		return nil, err
	}

	products, err := getTable[DisplayableProduct](query)

	return &products, err
}

// return &lib.RequestError{
// 	Message:    "Not Found",
// 	StatusCode: 404,
// }

func UpdateProduct(id int, p *Product) error {
	_, err := db.Exec(
		`
				UPDATE
					products
				SET 
					name           = $1,
					width          = $2, 
					length         = $3,
					height         = $4,
					active         = $5,
					product_type   = $6,
					color_name     = $7,
					price_cents    = $8
				WHERE
					id = $9
			`,
		p.Name,
		p.Width,
		p.Length,
		p.Height,
		p.Active,
		p.Product_Type,
		p.Color_Name,
		p.Price_Cents,
		id,
	)

	return err
}

func DeleteProduct(id int) error {
	_, err := db.Exec("DELETE FROM products WHERE id = $1", id)
	return err
}

func InsertProduct(p *Product) error {
	_, err := db.Exec(
		`
			INSERT INTO products (
				name,
				width, 
				length,
				height,
				product_type,
				active,
				price_cents,
				color_name
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`,
		p.Name,
		p.Width,
		p.Length,
		p.Height,
		p.Product_Type,
		p.Active,
		p.Price_Cents,
		p.Color_Name,
	)

	return err
}
