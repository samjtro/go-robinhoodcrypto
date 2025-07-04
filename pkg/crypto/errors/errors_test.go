package errors

import (
	"encoding/json"
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name     string
		apiError APIError
		want     string
	}{
		{
			name: "validation error with single field",
			apiError: APIError{
				Type:       "validation_error",
				StatusCode: 400,
				Errors: []ErrorDetail{
					{Attr: "client_order_id", Detail: "Must be a valid UUID."},
				},
			},
			want: "API error (status 400): validation_error - client_order_id: Must be a valid UUID.",
		},
		{
			name: "validation error with multiple fields",
			apiError: APIError{
				Type:       "validation_error",
				StatusCode: 400,
				Errors: []ErrorDetail{
					{Attr: "symbol", Detail: "Invalid symbol"},
					{Attr: "quantity", Detail: "Must be positive"},
				},
			},
			want: "API error (status 400): validation_error - symbol: Invalid symbol; quantity: Must be positive",
		},
		{
			name: "client error without field",
			apiError: APIError{
				Type:       "client_error",
				StatusCode: 401,
				Errors: []ErrorDetail{
					{Attr: "", Detail: "Unauthorized"},
				},
			},
			want: "API error (status 401): client_error - Unauthorized",
		},
		{
			name: "server error without details",
			apiError: APIError{
				Type:       "server_error",
				StatusCode: 500,
				Errors:     []ErrorDetail{},
			},
			want: "API error (status 500): server_error",
		},
		{
			name: "non-field errors",
			apiError: APIError{
				Type:       "validation_error",
				StatusCode: 400,
				Errors: []ErrorDetail{
					{Attr: "non_field_errors", Detail: "Order would exceed position limit"},
				},
			},
			want: "API error (status 400): validation_error - non_field_errors: Order would exceed position limit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.apiError.Error()
			if got != tt.want {
				t.Errorf("APIError.Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseAPIError(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		statusCode int
		wantErr    bool
		wantType   string
		wantDetail string
	}{
		{
			name: "valid validation error",
			body: `{
				"type": "validation_error",
				"errors": [
					{
						"detail": "Must be a valid UUID.",
						"attr": "client_order_id"
					}
				]
			}`,
			statusCode: 400,
			wantErr:    false,
			wantType:   "validation_error",
			wantDetail: "Must be a valid UUID.",
		},
		{
			name: "rate limit error",
			body: `{
				"type": "client_error",
				"errors": [
					{
						"detail": "Too many requests",
						"attr": null
					}
				]
			}`,
			statusCode: 429,
			wantErr:    false,
			wantType:   "client_error",
			wantDetail: "Too many requests",
		},
		{
			name:       "invalid JSON",
			body:       `{invalid json`,
			statusCode: 400,
			wantErr:    true,
		},
		{
			name:       "non-JSON error response",
			body:       `Internal Server Error`,
			statusCode: 500,
			wantErr:    true,
		},
		{
			name: "empty errors array",
			body: `{
				"type": "server_error",
				"errors": []
			}`,
			statusCode: 503,
			wantErr:    false,
			wantType:   "server_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ParseAPIError([]byte(tt.body), tt.statusCode)
			
			if tt.wantErr {
				if err == nil {
					t.Error("ParseAPIError() expected error, got nil")
				}
				// For non-API errors, check if it's a regular error
				if _, ok := err.(*APIError); ok {
					t.Error("expected non-APIError for invalid JSON")
				}
				return
			}

			if err == nil {
				t.Fatal("ParseAPIError() returned nil error")
			}

			apiErr, ok := err.(*APIError)
			if !ok {
				t.Fatalf("ParseAPIError() returned %T, want *APIError", err)
			}

			if apiErr.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", apiErr.Type, tt.wantType)
			}

			if apiErr.StatusCode != tt.statusCode {
				t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, tt.statusCode)
			}

			if tt.wantDetail != "" && len(apiErr.Errors) > 0 {
				if apiErr.Errors[0].Detail != tt.wantDetail {
					t.Errorf("Detail = %q, want %q", apiErr.Errors[0].Detail, tt.wantDetail)
				}
			}
		})
	}
}

func TestAPIError_JSONMarshaling(t *testing.T) {
	apiErr := &APIError{
		Type:       "validation_error",
		StatusCode: 400,
		Errors: []ErrorDetail{
			{Attr: "symbol", Detail: "Invalid trading pair"},
			{Attr: "quantity", Detail: "Below minimum order size"},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(apiErr)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// Unmarshal back
	var decoded APIError
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	// Verify fields (StatusCode is not included in JSON)
	if decoded.Type != apiErr.Type {
		t.Errorf("Type = %q, want %q", decoded.Type, apiErr.Type)
	}

	if len(decoded.Errors) != len(apiErr.Errors) {
		t.Errorf("len(Errors) = %d, want %d", len(decoded.Errors), len(apiErr.Errors))
	}

	for i, detail := range decoded.Errors {
		if detail.Attr != apiErr.Errors[i].Attr {
			t.Errorf("Errors[%d].Attr = %q, want %q", i, detail.Attr, apiErr.Errors[i].Attr)
		}
		if detail.Detail != apiErr.Errors[i].Detail {
			t.Errorf("Errors[%d].Detail = %q, want %q", i, detail.Detail, apiErr.Errors[i].Detail)
		}
	}
}