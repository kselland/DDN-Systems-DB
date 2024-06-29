package db

type ProductType string

const (
	Cabinet   ProductType = "cabinet"
	Accessory ProductType = "accessory"
)

type ColorProductType struct {
	Id           int
	Product_Type ProductType
	Color_Name   string
}

func GetColorProductTypesFromDb() ([]ColorProductType, error) {
	colorsQuery, colorsQueryErr := db.Query("SELECT * FROM color_product_types")
	if colorsQueryErr != nil {
		return nil, colorsQueryErr
	}

	return getTable[ColorProductType](colorsQuery)
}
