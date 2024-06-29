package db

import (
	"context"
	"ddn/ddn/lib"
	"errors"
	"net/http"
	"strconv"
)

type InventoryItem struct {
	Id                  int
	Product_Id          int
	Quantity            int
	Batch_Number        int
	Storage_Location_Id int
}

func (item *InventoryItem) ToFormData() FormInventoryItem {
	return FormInventoryItem{
		Product_Id:          strconv.Itoa(item.Product_Id),
		Quantity:            strconv.Itoa(item.Quantity),
		Batch_Number:        strconv.Itoa(item.Batch_Number),
		Storage_Location_Id: strconv.Itoa(item.Storage_Location_Id),
		Id:                  &item.Id,
	}
}

type DisplayableInventoryItem struct {
	Id                   int
	Product_Id           int
	Product_Name         string
	Quantity             int
	Batch_Number         int
	Storage_Location_Bin string
	Storage_Location_Id  int
}

type FormInventoryItem struct {
	Id                  *int
	Product_Id          string
	Quantity            string
	Batch_Number        string
	Storage_Location_Id string
}

type DeductionDataItem struct {
	Id       int
	Quantity int
}

func GetInventoryItem(id int) (*InventoryItem, error) {
	query, err := db.Query("SELECT * FROM inventory_items WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	result, err := getTable[InventoryItem](query)
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, &lib.RequestError{
			Message:    "Not Found",
			StatusCode: 404,
		}
	}
	return &result[0], nil
}

type InventoryItemsFilter struct {
	Search              string
	StorageLocationsStr string
	StorageLocations    []string
	MinQuantity         int
	BatchNumbersStr     string
	BatchNumbers        []string
	ProductIdsStr       string
	ProductIds          []string
}
type CountStruct struct {
	Count int
}

func CountFilteredInventoryItems(filter InventoryItemsFilter) (*int, error) {
	query, err := db.Query(`
		SELECT 
			COUNT(*) as count
		FROM
			inventory_items i
		LEFT JOIN
			storage_locations s
		ON
			i.storage_location_id = s.id
		LEFT JOIN
			products p
		ON
			i.product_id = p.id
		WHERE
			(
				i.id::text ILIKE '%' || $1 || '%'
				OR i.product_id::text ILIKE '%' || $1 || '%'
				OR p.name ILIKE '%' || $1 || '%'
				OR i.quantity::text ILIKE '%' || $1 || '%'
				OR i.batch_number::text ILIKE '%' || $1 || '%'
				OR s.bin ILIKE '%' || $1 || '%'
			)
			AND ($2 = '' OR i.storage_location_id = ANY (string_to_array($2, ',')::int[]))
			AND i.quantity >= $3
			AND ($4 = '' OR i.batch_number = ANY (string_to_array($4, ',')::int[]))
			AND ($5 = '' OR i.product_id = ANY(string_to_array($5, ',')::int[]))
		`,
		filter.Search,
		filter.StorageLocationsStr,
		filter.MinQuantity,
		filter.BatchNumbersStr,
		filter.ProductIdsStr,
	)
	if err != nil {
		return nil, err
	}
	countStruct, err := getFirst[CountStruct](query)
	if err != nil {
		return nil, err
	}

	return &countStruct.Count, nil
}

func GetFilteredDisplayableInventoryItems(filter InventoryItemsFilter, perPage int, page int) (*[]DisplayableInventoryItem, error) {
	query, err := db.Query(`
		SELECT 
			i.id,
			i.product_id,
			p.name product_name,
			i.quantity,
			i.batch_number,
			s.bin storage_location_bin,
			i.storage_location_id
		FROM
			inventory_items i
		LEFT JOIN
			storage_locations s
		ON
			i.storage_location_id = s.id
		LEFT JOIN
			products p
		ON
			i.product_id = p.id
		WHERE
			(
				i.id::text ILIKE '%' || $1 || '%'
				OR i.product_id::text ILIKE '%' || $1 || '%'
				OR p.name ILIKE '%' || $1 || '%'
				OR i.quantity::text ILIKE '%' || $1 || '%'
				OR i.batch_number::text ILIKE '%' || $1 || '%'
				OR s.bin ILIKE '%' || $1 || '%'
			)
			AND ($2 = '' OR i.storage_location_id = ANY (string_to_array($2, ',')::int[]))
			AND i.quantity >= $3
			AND ($4 = '' OR i.batch_number = ANY (string_to_array($4, ',')::int[]))
			AND ($5 = '' OR i.product_id = ANY(string_to_array($5, ',')::int[]))
		LIMIT
		    $6
		OFFSET
		    $7
		`,
		filter.Search,
		filter.StorageLocationsStr,
		filter.MinQuantity,
		filter.BatchNumbersStr,
		filter.ProductIdsStr,
		perPage,
		(page-1)*perPage,
	)
	if err != nil {
		return nil, err
	}
	inventoryItems, err := getTable[DisplayableInventoryItem](query)
	if err != nil {
		return nil, err
	}

	return &inventoryItems, nil
}

func GetBatchNumberOptions() ([]Option, error) {
	query, err := db.Query(`SELECT DISTINCT batch_number value, batch_number text FROM inventory_items`)
	if err != nil {
		return nil, err
	}
	batchNumberOptions, err := getTable[Option](query)
	if err != nil {
		return nil, err
	}

	return batchNumberOptions, nil
}

func GetFormInventoryItemFromPost(r *http.Request, id *int) FormInventoryItem {
	return FormInventoryItem{
		Id:                  id,
		Product_Id:          r.PostFormValue("product_id"),
		Quantity:            r.PostFormValue("quantity"),
		Batch_Number:        r.PostFormValue("batch_number"),
		Storage_Location_Id: r.PostFormValue("storage_location_id"),
	}
}

type InventoryItemValidation struct {
	Root                string
	Product_Id          string
	Quantity            string
	Batch_Number        string
	Storage_Location_Id string
}

func (formData FormInventoryItem) Validate() (InventoryItemValidation, *InventoryItem) {
	valid := true
	validation := InventoryItemValidation{}

	var productId int
	var productIdErr error
	if formData.Product_Id == "" {
		validation.Product_Id = "Product id is required"
		valid = false
	} else if productId, productIdErr = strconv.Atoi(formData.Product_Id); productIdErr != nil {
		validation.Product_Id = "Product id must be an integer"
		valid = false
	} else {
		// TODO: Ensure the product exists in the db
	}

	var quantity int
	var quantityErr error
	if formData.Quantity == "" {
		validation.Quantity = "Quantity is required"
		valid = false
	} else if quantity, quantityErr = strconv.Atoi(formData.Quantity); quantityErr != nil {
		validation.Quantity = "Quantity must be an integer"
		valid = false
	}

	var batchNumber int
	var batchNumberErr error
	if formData.Batch_Number == "" {
		validation.Batch_Number = "Batch number is required"
		valid = false
	} else if batchNumber, batchNumberErr = strconv.Atoi(formData.Batch_Number); batchNumberErr != nil {
		validation.Batch_Number = "Batch number must be an integer"
		valid = false
	}

	var storageLocationId int
	var storageLocationIdErr error
	if formData.Storage_Location_Id == "" {
		validation.Storage_Location_Id = "Storage location is required"
		valid = false
	} else if storageLocationId, storageLocationIdErr = strconv.Atoi(formData.Storage_Location_Id); storageLocationIdErr != nil {
		validation.Storage_Location_Id = "Storage location id must be an integer"
		valid = false
	} else {
		// TODO: Ensure the storage location exists in the db
	}

	if !valid {
		if validation.Root == "" {
			validation.Root = "There are errors with your submission"
		}
		return validation, nil
	}

	return validation, &InventoryItem{
		Product_Id:          productId,
		Quantity:            quantity,
		Batch_Number:        batchNumber,
		Storage_Location_Id: storageLocationId,
	}
}

func DeleteInventoryItem(id int) error {
	_, err := db.Exec("DELETE FROM inventory_items WHERE id = $1", id)
	return err
}

func InsertInventoryItem(i InventoryItem) error {
	_, err := db.Exec(`
		INSERT INTO inventory_items (
			storage_location_id,
			product_id,
			quantity,
			batch_number
		) VALUES ($1, $2, $3, $4)
		`,
		i.Storage_Location_Id,
		i.Product_Id,
		i.Quantity,
		i.Batch_Number,
	)

	return err
}

func UpdateInventoryItem(id int, i InventoryItem) error {
	_, err := db.Exec(
		`
				UPDATE
					inventory_items
				SET 
                    storage_location_id = $1,
                    product_id          = $2,
                    batch_number        = $3,
                    quantity            = $4
				WHERE
					id = $5
			`,
		i.Storage_Location_Id,
		i.Product_Id,
		i.Batch_Number,
		i.Quantity,
		id,
	)

	return err
}

func GetInventoryItems() (*[]InventoryItem, error) {
	query, err := db.Query(`
		SELECT
			id,
			product_id,
			quantity,
			batch_number,
			storage_location_id
		FROM
			inventory_items
	`)
	if err != nil {
		return nil, err
	}

	res, err := getTable[InventoryItem](query)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func DeductInventoryItems(data []DeductionDataItem) error {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, deductionItem := range data {
		query, err := tx.Query(`
			SELECT
				quantity
			FROM
				inventory_items
			WHERE
				id = $1
		`, deductionItem.Id)
		if err != nil {
			return err
		}

		res, err := getFirst[struct{ Quantity int }](query)
		if err != nil {
			return err
		}
		if res == nil {
			return errors.New("Specified id isn't present")
		}
		if deductionItem.Quantity > res.Quantity {
			return errors.New("There was't enough inventory to process the request")
		}

		if deductionItem.Quantity == res.Quantity {
			_, err = tx.Exec("DELETE FROM inventory_items WHERE id = $1", deductionItem.Id)
			if err != nil {
				return err
			}
		} else {
			_, err = tx.Exec(`
				UPDATE
					inventory_items
				SET
					quantity = $1
			`, res.Quantity-deductionItem.Quantity)
			if err != nil {
				return err
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
