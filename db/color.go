package db

type Color struct {
	Hex_Code string
	Name     string
}

func GetColorsFromDb() ([]Color, error) {
	colorsQuery, colorsQueryErr := db.Query("SELECT hex_code, name FROM colors")
	if colorsQueryErr != nil {
		return nil, colorsQueryErr
	}

	return getTable[Color](colorsQuery)
}
