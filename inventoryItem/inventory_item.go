package inventoryItem

import (
	"context"
	"ddn/ddn/db"
	"ddn/ddn/lib"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type EditableInventoryItemProps struct {
	FormInventoryItem db.FormInventoryItem
	Validation        db.InventoryItemValidation
	Id                *int
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

	if page != numPages && numPages != 0 {
		pagination = append(pagination, PaginationItem{
			paginationType: Next,
			to:             page + 1,
		})
	}

	return pagination
}

func getFilter(r *http.Request) db.InventoryItemsFilter {
	filter := db.InventoryItemsFilter{}

	filter.Search = r.URL.Query().Get("search")

	// TODO: Theoretically if a bin had a , in it then this approach would break
	filter.StorageLocations = r.URL.Query()["storageLocations"]
	filter.StorageLocationsStr = strings.Join(filter.StorageLocations, ",")

	minQuantityStr := r.URL.Query().Get("minQuantity")
	minQuantity, err := strconv.Atoi(minQuantityStr)
	if err != nil {
		minQuantity = 0
	}
	filter.MinQuantity = minQuantity

	filter.BatchNumbers = r.URL.Query()["batchNumbers"]
	filter.BatchNumbersStr = strings.Join(filter.BatchNumbers, ",")

	filter.ProductIds = r.URL.Query()["productIds"]
	filter.ProductIdsStr = strings.Join(filter.ProductIds, ",")

	return filter
}

func IndexPage(s *db.Session, w http.ResponseWriter, r *http.Request) error {
	filter := getFilter(r)

	page, pageErr := strconv.Atoi(r.URL.Query().Get("page"))
	if pageErr != nil {
		page = 1
	}
	perPage, perPageErr := strconv.Atoi(r.URL.Query().Get("perPage"))
	if perPageErr != nil {
		perPage = 10
	}

	count, err := db.CountFilteredInventoryItems(filter)
	if err != nil {
		return err
	}

	numPages := ceilDivide(*count, perPage)
	if page > numPages {
		page = 1
	}

	inventoryItems, err := db.GetFilteredDisplayableInventoryItems(filter, perPage, page)
	if err != nil {
		return err
	}

	storageLocationOptions, err := db.GetStorageLocationOptions()
	if err != nil {
		return err
	}

	productIdOptions, err := db.GetProductOptions()
	if err != nil {
		return err
	}

	batchNumberOptions, err := db.GetBatchNumberOptions()
	if err != nil {
		return err
	}

	return indexTemplate(
		s,
		IndexTemplateProps{
			filter:                 filter,
			storageLocationOptions: *storageLocationOptions,
			productIdOptions:       productIdOptions,
			batchNumberOptions:     batchNumberOptions,
			filteredItems:          *inventoryItems,
			pagination:             getPagination(numPages, perPage, page),
			perPage:                perPage,
			page:                   page,
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

func ViewPage(s *db.Session, w http.ResponseWriter, r *http.Request) error {
	idString := mux.Vars(r)["id"]

	id, err := strconv.Atoi(idString)
	if err != nil {
		return &lib.RequestError{
			Message:    "Not Found",
			StatusCode: 404,
		}
	}

	inventoryItem, err := db.GetInventoryItem(id)
	if err != nil {
		return nil
	}

	productOptions, err := db.GetProductOptions()
	if err != nil {
		return nil
	}

	storageLocationOptions, err := db.GetStorageLocationOptions()
	if err != nil {
		return nil
	}

	if r.Method == "POST" {
		formInventoryItem := db.GetFormInventoryItemFromPost(r, &id)

		validation, inventoryItem := formInventoryItem.Validate()

		if inventoryItem == nil {
			return viewTemplate(
				s,
				formInventoryItem,
				validation,
				productOptions,
				*storageLocationOptions,
			).Render(context.Background(), w)
		}

		err := db.UpdateInventoryItem(id, *inventoryItem)
		if err != nil {
			return viewTemplate(
				s,
				formInventoryItem,
				db.InventoryItemValidation{
					Root: "Failed to update inventory  in DB",
				},
				productOptions,
				*storageLocationOptions,
			).Render(context.Background(), w)
		}

		http.Redirect(w, r, fmt.Sprintf("/app/inventory-item/%d", id), http.StatusSeeOther)
		return nil
	}

	return viewTemplate(
		s,
		inventoryItem.ToFormData(),
		db.InventoryItemValidation{},
		productOptions,
		*storageLocationOptions,
	).Render(context.Background(), w)
}

func DeletePage(s *db.Session, w http.ResponseWriter, r *http.Request) error {
	idStr := mux.Vars(r)["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return err
	}

	inventoryItem, err := db.GetInventoryItem(id)

	if inventoryItem == nil {
		return &lib.RequestError{
			Message:    "Not Found",
			StatusCode: 404,
		}
	}

	db.DeleteInventoryItem(id)

	http.Redirect(w, r, "/app/inventory", http.StatusSeeOther)
	return nil
}

func NewPage(s *db.Session, w http.ResponseWriter, r *http.Request) error {
	productOptions, err := db.GetProductOptions()
	if err != nil {
		return err
	}

	storageLocationOptions, err := db.GetStorageLocationOptions()
	if err != nil {
		return err
	}

	if r.Method == "POST" {
		formInventoryItem := db.GetFormInventoryItemFromPost(r, nil)
		validation, inventoryItem := formInventoryItem.Validate()

		if inventoryItem == nil {
			return newTemplate(
				s,
				formInventoryItem,
				validation,
				productOptions,
				*storageLocationOptions,
			).Render(context.Background(), w)
		}

		err := db.InsertInventoryItem(*inventoryItem)
		if err != nil {
			return newTemplate(
				s,
				formInventoryItem,
				db.InventoryItemValidation{Root: "Error saving inventory to database"},
				productOptions,
				*storageLocationOptions,
			).Render(context.Background(), w)
		}

		http.Redirect(w, r, "/app/inventory", http.StatusSeeOther)

		return nil
	}

	return newTemplate(
		s,
		db.FormInventoryItem{},
		db.InventoryItemValidation{},
		productOptions,
		*storageLocationOptions,
	).Render(context.Background(), w)
}

func DeductPage(s *db.Session, w http.ResponseWriter, r *http.Request) error {
	if r.Method == "POST" {
		r.ParseForm()
		jsonString := r.PostForm.Get("json_deductions")

		var data []db.DeductionDataItem
		err := json.Unmarshal([]byte(jsonString), &data)
		if err != nil || data == nil {
			return err
		}

		err = db.DeductInventoryItems(data)
		if err != nil {
			return err
		}

		http.Redirect(w, r, "/app", http.StatusSeeOther)
		return nil
	}

	productOptions, err := db.GetProductOptions()
	if err != nil {
		return err
	}

	storageLocationOptions, err := db.GetStorageLocationOptions()
	if err != nil {
		return err
	}

	inventoryItems, err := db.GetInventoryItems()
	if err != nil {
		return err
	}

	type JsonData struct {
		ProductOptions         []db.Option
		StorageLocationOptions []db.Option
		InventoryItems         []db.InventoryItem
	}

	jsonStr, err := json.Marshal(JsonData{
		ProductOptions:         productOptions,
		StorageLocationOptions: *storageLocationOptions,
		InventoryItems:         *inventoryItems,
	})

	return deductTemplate(s, DeductTemplateProps{
		productOptions: productOptions,
		jsonData:       string(jsonStr),
	}).Render(context.Background(), w)
}
