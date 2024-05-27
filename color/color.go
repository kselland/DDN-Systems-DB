package color

import "ddn/ddn/product_type"
import "ddn/ddn/db"

type Color struct {
    Id int
    Hex_Code string
    Name string
    Product_Type product_type.ProductType
}

func GetColorsFromDb() ([]Color, error) {
	colorsQuery, colorsQueryErr := db.Db.Query("SELECT * FROM colors")
	if colorsQueryErr != nil {
		return nil, colorsQueryErr
	}
	
    return db.GetTable[Color](colorsQuery)
}
