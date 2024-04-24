package inventoryItem

import (
	"context"
	"ddn/ddn/db"
	"ddn/ddn/lib"
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

func IndexPage(w http.ResponseWriter, r *http.Request) error {
	query, err := db.Db.Query("SELECT * FROM inventory_items")
	if err != nil {
		return err
	}

	inventoryItems := db.GetTable[InventoryItem](query)

	return indexTemplate(
		inventoryItems,
	).Render(context.Background(), w)
}

func ViewPage(w http.ResponseWriter, r *http.Request) error {
	idString := mux.Vars(r)["id"]

	id, err := strconv.Atoi(idString)
	if err != nil {
		return &lib.RequestError{
			Message: "Not Found",
			StatusCode: 404,
		}
	}

	inventoryItemsQuery, inventoryItemsErr := db.Db.Query("SELECT * FROM inventory_items WHERE id = $1", id)
	if inventoryItemsErr != nil {
		return err
	}
	inventoryItems := db.GetTable[InventoryItem](inventoryItemsQuery)
	if len(inventoryItems) == 0 {
		return &lib.RequestError{
			Message: "Not Found",
			StatusCode: 404,
		}
	}
	inventoryItem := inventoryItems[0]

	if r.Method == "POST" {
		formInventoryItem := getFormDataFromPost(r, &id)

		validation, inventoryItem := formInventoryItem.validate()
		fmt.Println("formInventoryItem", formInventoryItem, validation, inventoryItem)

		if inventoryItem == nil {
			return viewTemplate(
				formInventoryItem,
				validation,
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
				formInventoryItem,
				InventoryItemValidation{
					Root: "Failed to update inventory  in DB",
				},
			).Render(context.Background(), w)
		}

		http.Redirect(w, r, fmt.Sprintf("/inventory/%d", id), http.StatusSeeOther)
		return nil
	}

	return viewTemplate(
		inventoryItem.toFormData(),
		InventoryItemValidation{},
	).Render(context.Background(), w)
}

func DeletePage(w http.ResponseWriter, r *http.Request) error {
	id := mux.Vars(r)["id"]

	query, err := db.Db.Query("SELECT * FROM inventory_items WHERE id = $1", id)
	if err != nil {
		return err
	}

	inventoryItems := db.GetTable[InventoryItem](query)

	if len(inventoryItems) == 0 {
		return &lib.RequestError{
			Message: "Not Found",
			StatusCode: 404,
		}
	}

	_, err2 := db.Db.Exec("DELETE FROM inventory_items WHERE id = $1", id)
	if err2 != nil {
		return err2
	}

	http.Redirect(w, r, "/inventory", http.StatusSeeOther)
	return nil
}

func NewPage(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "POST" {
		formInventoryItem := getFormDataFromPost(r, nil)
		validation, inventoryItem := formInventoryItem.validate()
		fmt.Println("formInventoryItem", formInventoryItem, validation, inventoryItem)

		if inventoryItem == nil {
			return newTemplate(
				formInventoryItem,
				validation,
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
				formInventoryItem,
				InventoryItemValidation{Root: "Error saving inventory to database"},
			).Render(context.Background(), w)
		}

		http.Redirect(w, r, "/inventory", http.StatusSeeOther)

		return nil
	}

	return newTemplate(
		FormInventoryItem{},
		InventoryItemValidation{},
	).Render(context.Background(), w)
}
