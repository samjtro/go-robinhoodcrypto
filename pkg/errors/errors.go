package errors

import (
	"encoding/json"
	"fmt"
)

type ErrorDetail struct {
	Attr   string `json:"attr"`
	Detail string `json:"detail"`
}

type APIError struct {
	Type       string        `json:"type"`
	Errors     []ErrorDetail `json:"errors"`
	StatusCode int           `json:"-"`
}

func (e *APIError) Error() string {
	if len(e.Errors) == 0 {
		return fmt.Sprintf("API error (status %d): %s", e.StatusCode, e.Type)
	}
	
	details := ""
	for i, err := range e.Errors {
		if i > 0 {
			details += "; "
		}
		if err.Attr != "" {
			details += fmt.Sprintf("%s: %s", err.Attr, err.Detail)
		} else {
			details += err.Detail
		}
	}
	
	return fmt.Sprintf("API error (status %d): %s - %s", e.StatusCode, e.Type, details)
}

func ParseAPIError(body []byte, statusCode int) error {
	var apiErr APIError
	if err := json.Unmarshal(body, &apiErr); err != nil {
		return fmt.Errorf("status %d: %s", statusCode, string(body))
	}
	apiErr.StatusCode = statusCode
	return &apiErr
}