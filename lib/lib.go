package lib

import "fmt"

type RequestError struct {
    Message string
    StatusCode int
}

func (e *RequestError) Error() string {
    return fmt.Sprintf("Error %d: %s", e.StatusCode, e.Message)
}
