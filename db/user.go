package db

import "encoding/hex"

type AuthData struct {
	Id              int
	Password_Digest []byte
}

func GetAuthDataByUserId(email string) (*AuthData, error) {
	query, err := db.Query(`
        SELECT
            id,
            password_digest
        FROM
            users
        WHERE
            email = $1
    `, email)
	if err != nil {
		return nil, err
	}
	data, err := getFirst[AuthData](query)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func GetUsers() (*[]User, error) {
	query, err := db.Query(`
		SELECT
			u.id,
            u.email,
            u.password_digest,
            u.role,
            u.name
		FROM
			users u
	`)
	if err != nil {
		return nil, err
	}

	users, err := getTable[User](query)

	if err != nil {
		return nil, err
	}

	return &users, nil
}

type User struct {
	Id              int
	Email           string
	Password_digest []byte
	Role            UserRole
	Name            string
}

type UserRole string

func ParseUserRole(potential string) *UserRole {
	if potential == "superadmin" {
		a := UserRoleSuperAdmin
		return &a
	}
	if potential == "admin" {
		a := UserRoleAdmin
		return &a
	}
	if potential == "manager" {
		a := UserRoleManager
		return &a
	}
	if potential == "superadmin" {
		a := UserRoleSuperAdmin
		return &a
	}
	if potential == "viewer" {
		a := UserRoleViewer
		return &a
	}
	if potential == "builder" {
		a := UserRoleBuilder
		return &a
	}
	if potential == "driver" {
		a := UserRoleDriver
		return &a
	}
	return nil
}

const (
	UserRoleSuperAdmin UserRole = "superadmin"
	UserRoleAdmin      UserRole = "admin"
	UserRoleManager    UserRole = "manager"
	UserRoleViewer     UserRole = "viewer"
	UserRoleBuilder    UserRole = "builder"
	UserRoleDriver     UserRole = "driver"
)

type Permission int

const (
	PermissionCreateUser Permission = iota

	PermissionViewProducts
	PermissionEditProduct
	PermissionDeleteProduct
	PermissionCreateProduct

	PermissionViewStorageLocations
	PermissionEditStorageLocation
	PermissionDeleteStorageLocation
	PermissionCreateStorageLocation

	PermissionViewInventory
	PermissionEditInventoryItem
	PermissionDeleteInventoryItem
	PermissionCreateInventoryItem

	PermissionViewUsers
	PermissionEditUsers

	PermissionLoggedIn
	PermissionLoggedOut
	PermissionNone
)

func UserHasPermission(user *User, permission Permission) bool {
	if permission == PermissionNone {
		return true
	}

	if permission == PermissionLoggedIn && user != nil {
		return true
	}

	if user == nil {
		if permission == PermissionLoggedOut {
			return true
		} else {
			return false
		}
	}

	role := user.Role
	if role == UserRoleAdmin || role == UserRoleSuperAdmin {
		return true
	}

	// TODO: flesh out this function more

	return false
}

func InsertUser(user User) error {
	_, err := db.Exec(
		`
            INSERT INTO
                users
            (
                email,
                password_digest,
                role,
                name
            ) VALUES (
                $1,
                decode($2, 'hex'),
                $3,
                $4
            )
        `,
		user.Email,
		hex.EncodeToString(user.Password_digest),
		user.Role,
		user.Name,
	)

	return err
}

func GetRoleOptions() ([]Option, error) {
	query, err := db.Query(`
		SELECT 
			enumlabel AS value, 
			enumlabel AS text
		FROM 
			pg_enum
		JOIN 
			pg_type ON pg_enum.enumtypid = pg_type.oid
		WHERE 
			pg_type.typname = 'user_role';
	`)
	roles, err := getTable[Option](query)

    return roles, err
}
