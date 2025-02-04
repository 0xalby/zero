package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/charmbracelet/log"
	"github.com/glebarez/go-sqlite"
	"github.com/go-playground/validator/v10"
)

var Validator = validator.New(validator.WithRequiredStructEnabled())

// Validates JSON request body payloads
func Validate(w http.ResponseWriter, r *http.Request, payload any) error {
	if r.Body == nil {
		Response(w, http.StatusBadRequest, "empty request body")
		return fmt.Errorf("empty request")
	}
	// Decodes the request body in the payload
	err := json.NewDecoder(r.Body).Decode(payload)
	if err != nil {
		Response(w, http.StatusBadRequest, "invalid request body")
		return err
	}
	// Checks for validation errors in the payload
	if err := Validator.Struct(payload); err != nil {
		if verrs := err.(validator.ValidationErrors); verrs != nil {
			Response(w, http.StatusBadRequest, "failed to validate request body")
			return err
		}
	}
	return nil
}

// Handles possible SQLite3 errors
func HandleSQLiteErrors(err error) error {
	sqliteErr, ok := err.(*sqlite.Error)
	if ok {
		switch sqliteErr.Code() {
		// UNIQUE constraint
		case 19:
			return fmt.Errorf("UNIQUE constraint failed %v", sqliteErr)
		// IO failed
		case 10:
			return fmt.Errorf("I/O database error %v", sqliteErr)
		// MISMATCH
		case 20:
			return fmt.Errorf("MISMATCH of data in a row %v", sqliteErr)
		// MISUSE
		case 21:
			return fmt.Errorf("MISUSE of sqlite %v", sqliteErr)
		// READONLY database
		case 8:
			return fmt.Errorf("unable to modify database in read only mode %v", sqliteErr)
		// INTERRUPTED operation
		case 9:
			return fmt.Errorf("interrupted operation %v", sqliteErr)
		// UNHANDLED error code
		default:
			return fmt.Errorf("sqlite3 error: %v", err)
		}
	}
	return err
}

// Sends a response
func Response(w http.ResponseWriter, status int, v any) error {
	// Adding the http status as an header
	w.WriteHeader(status)
	// Encoding the payload
	return json.NewEncoder(w).Encode(v)
}

// Sends a request
func Request(method string, headers map[string]string, endpoint string, payload any) (*http.Response, error) {
	// Marshaling the payload
	marshal, err := json.Marshal(payload)
	if err != nil {
		log.Error("error reading payload", "err", err)
		return nil, fmt.Errorf("error reading payload")
	}
	// Creates a new http request
	req, err := http.NewRequest(method, endpoint, bytes.NewBuffer(marshal))
	if err != nil {
		log.Error("error creating request", "err", err)
		return nil, fmt.Errorf("error creating request")
	}
	for header, i := range headers {
		req.Header.Set(header, i)
	}
	// Creates a new http client
	client := &http.Client{Timeout: time.Minute * 10}
	// Sending the request
	resp, err := client.Do(req)
	if err != nil {
		log.Error("error sending request", "err", err)
		return resp, err
	}
	return resp, nil
}
