package inventoryItem

import (
	"context"
	"ddn/ddn/components"
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

type InventoryItem struct {
	Id                  int
	Product_Id          int
	Quantity            int
	Batch_Number        int
	Storage_Location_Id int
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

func (item InventoryItem) toFormData() FormInventoryItem {
	return FormInventoryItem{
		Product_Id:          strconv.Itoa(item.Product_Id),
		Quantity:            strconv.Itoa(item.Quantity),
		Batch_Number:        strconv.Itoa(item.Batch_Number),
		Storage_Location_Id: strconv.Itoa(item.Storage_Location_Id),
		Id:                  &item.Id,
	}
}

func getFormDataFromPost(r *http.Request, id *int) FormInventoryItem {
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

func (formData FormInventoryItem) validate() (InventoryItemValidation, *InventoryItem) {
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

type EditableInventoryItemProps struct {
	FormInventoryItem FormInventoryItem
	Validation        InventoryItemValidation
	Id                *int
}

func getProductOptions() ([]components.Option, error) {
	query, err := db.Db.Query(`
		SELECT
			id value,
			name text
		FROM
			products
	`)
	if err != nil {
		return nil, err
	}
	return db.GetTable[components.Option](query)
}

func getStorageLocationOptions() ([]components.Option, error) {
	query, err := db.Db.Query(`
		SELECT
			id value,
			bin text
		FROM
			storage_locations
	`)
	if err != nil {
		return nil, err
	}
	return db.GetTable[components.Option](query)
}

func getInventoryItem(id int) (*InventoryItem, error) {
	query, err := db.Db.Query("SELECT * FROM inventory_items WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	result, err := db.GetTable[InventoryItem](query)
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

func getPagination(numPages int, perPage int, page int) []PaginationItem {
	var pagination []PaginationItem

	if page != 1 {
		pagination = append(pagination, PaginationItem{
			paginationType: Back,
			to:             page - 1,
		})
	}

	if page-2 > 1 {
		pagination = append(pagination, PaginationItem{
			paginationType: Ellipsis,
			to:             page - 3,
		})
	}

	for i := max(1, page-2); i <= min(page+2, numPages); i++ {
		pagination = append(pagination, PaginationItem{
			paginationType: Number,
			to:             i,
		})
	}

	if page+2 < numPages {
		pagination = append(pagination, PaginationItem{
			paginationType: Ellipsis,
			to:             page + 3,
		})
	}

	if page != numPages {
		pagination = append(pagination, PaginationItem{
			paginationType: Next,
			to:             page + 1,
		})
	}

	return pagination
}

type CountStruct struct {
	Count int
}

func IndexPage(s *session.Session, w http.ResponseWriter, r *http.Request) error {
	search := r.URL.Query().Get("search")

	page, pageErr := strconv.Atoi(r.URL.Query().Get("page"))
	if pageErr != nil {
		page = 1
	}
	perPage, perPageErr := strconv.Atoi(r.URL.Query().Get("perPage"))
	if perPageErr != nil {
		perPage = 10
	}

	query, err := db.Db.Query(`
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
			i.id::text ILIKE '%' || $1 || '%'
			OR i.product_id::text ILIKE '%' || $1 || '%'
			OR p.name ILIKE '%' || $1 || '%'
			OR i.quantity::text ILIKE '%' || $1 || '%'
			OR i.batch_number::text ILIKE '%' || $1 || '%'
			OR s.bin ILIKE '%' || $1 || '%'
		LIMIT
		    $2
		OFFSET
		    $3
	`, search, perPage, (page-1)*perPage)
	if err != nil {
		return err
	}
	inventoryItems, err := db.GetTable[DisplayableInventoryItem](query)
	if err != nil {
		return err
	}

	query, err = db.Db.Query(`
		SELECT COUNT(*) as count FROM inventory_items;
	`)
	if err != nil {
		return err
	}
	countStruct, err := db.GetFirst[CountStruct](query)
	if err != nil {
		return err
	}

	return indexTemplate(
		s,
		IndexTemplateProps{
			filteredItems: inventoryItems,
			search:        search,
			pagination:    getPagination(ceilDivide(countStruct.Count, perPage), perPage, page),
			perPage:       perPage,
			page:          page,
		},
	).Render(context.Background(), w)
}

func ceilDivide(a int, b int) int {
	result := a / b
	if a%b != 0 {
		result++
	}

	return result
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

	inventoryItem, err := getInventoryItem(id)
	if err != nil {
		return nil
	}

	productOptions, err := getProductOptions()
	if err != nil {
		return nil
	}

	storageLocationOptions, err := getStorageLocationOptions()
	if err != nil {
		return nil
	}

	if r.Method == "POST" {
		formInventoryItem := getFormDataFromPost(r, &id)

		validation, inventoryItem := formInventoryItem.validate()
		fmt.Println("formInventoryItem", formInventoryItem, validation, inventoryItem)

		if inventoryItem == nil {
			return viewTemplate(
				s,
				formInventoryItem,
				validation,
				productOptions,
				storageLocationOptions,
			).Render(context.Background(), w)
		}

		_, err := db.Db.Exec(
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
			inventoryItem.Storage_Location_Id,
			inventoryItem.Product_Id,
			inventoryItem.Batch_Number,
			inventoryItem.Quantity,
			id,
		)
		if err != nil {
			return viewTemplate(
				s,
				formInventoryItem,
				InventoryItemValidation{
					Root: "Failed to update inventory  in DB",
				},
				productOptions,
				storageLocationOptions,
			).Render(context.Background(), w)
		}

		http.Redirect(w, r, fmt.Sprintf("/app/inventory-item/%d", id), http.StatusSeeOther)
		return nil
	}

	return viewTemplate(
		s,
		inventoryItem.toFormData(),
		InventoryItemValidation{},
		productOptions,
		storageLocationOptions,
	).Render(context.Background(), w)
}

func DeletePage(w http.ResponseWriter, r *http.Request) error {
	id := mux.Vars(r)["id"]

	query, err := db.Db.Query("SELECT * FROM inventory_items WHERE id = $1", id)
	if err != nil {
		return err
	}

	inventoryItems, err := db.GetTable[InventoryItem](query)
	if err != nil {
		return err
	}

	if len(inventoryItems) == 0 {
		return &lib.RequestError{
			Message:    "Not Found",
			StatusCode: 404,
		}
	}

	_, err2 := db.Db.Exec("DELETE FROM inventory_items WHERE id = $1", id)
	if err2 != nil {
		return err2
	}

	http.Redirect(w, r, "/app/inventory", http.StatusSeeOther)
	return nil
}

func NewPage(s *session.Session, w http.ResponseWriter, r *http.Request) error {
	productOptions, err := getProductOptions()
	if err != nil {
		return nil
	}

	storageLocationOptions, err := getStorageLocationOptions()
	if err != nil {
		return nil
	}

	if r.Method == "POST" {
		formInventoryItem := getFormDataFromPost(r, nil)
		validation, inventoryItem := formInventoryItem.validate()
		fmt.Println("formInventoryItem", formInventoryItem, validation, inventoryItem)

		if inventoryItem == nil {
			return newTemplate(
				s,
				formInventoryItem,
				validation,
				productOptions,
				storageLocationOptions,
			).Render(context.Background(), w)
		}

		_, err := db.Db.Exec(
			`
				INSERT INTO inventory_items (
                    storage_location_id,
                    product_id,
                    quantity,
                    batch_number
				) VALUES ($1, $2, $3, $4)
			`,
			inventoryItem.Storage_Location_Id,
			inventoryItem.Product_Id,
			inventoryItem.Quantity,
			inventoryItem.Batch_Number,
		)
		if err != nil {
			log.Println(err)
			return newTemplate(
				s,
				formInventoryItem,
				InventoryItemValidation{Root: "Error saving inventory to database"},
				productOptions,
				storageLocationOptions,
			).Render(context.Background(), w)
		}

		http.Redirect(w, r, "/app/inventory", http.StatusSeeOther)

		return nil
	}

	return newTemplate(
		s,
		FormInventoryItem{},
		InventoryItemValidation{},
		productOptions,
		storageLocationOptions,
	).Render(context.Background(), w)
}
