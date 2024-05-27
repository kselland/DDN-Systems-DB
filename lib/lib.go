package lib

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/scrypt"
)

var encryptionSalt = []byte(os.Getenv("ENCRYPTION_SALT"))

type RequestError struct {
	Message    string
	StatusCode int
}

func (e *RequestError) Error() string {
	return fmt.Sprintf("Error %d: %s", e.StatusCode, e.Message)
}

func GetDigest(password string) ([]byte, error) {
	// TODO: Make this more secure potentially
	return scrypt.Key([]byte(password), encryptionSalt, 32768, 8, 1, 32)
}

func GenerateToken() (string, error) {
    length := 12

    b := make([]byte, length)
    if _, err := rand.Read(b); err != nil {
        return "", err
    }
    return hex.EncodeToString(b), nil
}

func IsAuthenticatedPath(path string) bool {
    return strings.HasPrefix(path, "/app") || strings.HasPrefix(path, "app")
}
