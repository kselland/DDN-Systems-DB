package appPaths

import (
	"ddn/ddn/db"
	"strings"

	"github.com/a-h/templ"
)

// TODO: should this system handle query params?

type AppPath string

const (
	Dashboard AppPath = "/app"

	ProductListing AppPath = "/app/products"
	ProductNew     AppPath = "/app/products/new"
	Product        AppPath = "/app/product/{id}"
	ProductDelete  AppPath = "/app/{id}/delete"

	StorageLocationListing AppPath = "/app/storage-locations"
	StorageLocationNew     AppPath = "/app/storage-locations/new"
	StorageLocation        AppPath = "/app/storage-location/{id}"
	StorageLocationDelete  AppPath = "/app/storage-location/{id}/delete"

	Inventory           AppPath = "/app/inventory"
	InventoryItemNew    AppPath = "/app/inventory/new"
	InventoryDeduct     AppPath = "/app/inventory/deduct"
	InventoryItem       AppPath = "/app/inventory-items/{id}"
	InventoryItemDelete AppPath = "/app/inventory-items/{id}/delete"

	UserListing AppPath = "/app/users"
	UserNew     AppPath = "/app/users/new"
	User        AppPath = "/app/user/{id}"
	UserDelete  AppPath = "/app/user/{id}/delete"

	Login  AppPath = "/login"
	Logout AppPath = "/logout"
)

func (p AppPath) WithParams(params map[string]string) templ.SafeURL {
	var builder strings.Builder
	var dynamic *strings.Builder

	for _, char := range string(p) {
		if char == '{' {
			var tmp strings.Builder
			dynamic = &tmp
		} else if char == '}' {
			builder.WriteString(params[dynamic.String()])
			dynamic = nil
		} else if dynamic != nil {
			dynamic.WriteRune(char)
		} else {
			builder.WriteRune(char)
		}
	}

	return templ.SafeURL(builder.String())
}

func (p AppPath) WithNoParams() templ.SafeURL {
	return p.WithParams(map[string]string{})
}

func (p *AppPath) Permissions() db.Permission {
	pathToPermsMap := map[AppPath]db.Permission{
		Dashboard: db.PermissionLoggedIn,

		ProductListing: db.PermissionLoggedIn,

		ProductNew: db.PermissionCreateProduct,

		Product:       db.PermissionEditProduct,
		ProductDelete: db.PermissionViewProducts,

		StorageLocationListing: db.PermissionViewStorageLocations,
		StorageLocationNew:     db.PermissionCreateStorageLocation,
		StorageLocation:        db.PermissionCreateStorageLocation,
		StorageLocationDelete:  db.PermissionCreateStorageLocation,

		Inventory:           db.PermissionViewInventory,
		InventoryItemNew:    db.PermissionCreateInventoryItem,
		InventoryItem:       db.PermissionCreateInventoryItem,
		InventoryItemDelete: db.PermissionCreateInventoryItem,

		UserNew: db.PermissionCreateUser,

		Login:  db.PermissionLoggedOut,
		Logout: db.PermissionLoggedIn,
	}

	return pathToPermsMap[*p]
}
