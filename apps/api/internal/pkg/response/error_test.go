package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteError_BasicResponse(t *testing.T) {
	w := httptest.NewRecorder()
	WriteError(w, http.StatusBadRequest, "INVALID_INPUT", "Field 'name' is required")

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusBadRequest)
	}

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}

	var body APIError
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	if body.Code != "INVALID_INPUT" {
		t.Errorf("Error code = %q, want %q", body.Code, "INVALID_INPUT")
	}
	if body.Message != "Field 'name' is required" {
		t.Errorf("Error message = %q, want %q", body.Message, "Field 'name' is required")
	}
}

func TestWriteError_AllHTTPStatusCodes(t *testing.T) {
	cases := []struct {
		name   string
		status int
		code   string
		msg    string
	}{
		{"Unauthorized", http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated"},
		{"Forbidden", http.StatusForbidden, "CSRF_FAILED", "Invalid CSRF token"},
		{"NotFound", http.StatusNotFound, "NOT_FOUND", "Resource not found"},
		{"Internal", http.StatusInternalServerError, "INTERNAL_ERROR", "Server error"},
		{"TooManyRequests", http.StatusTooManyRequests, "RATE_LIMIT", "Slow down"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			WriteError(w, tc.status, tc.code, tc.msg)

			if w.Code != tc.status {
				t.Errorf("Status = %d, want %d", w.Code, tc.status)
			}

			var body APIError
			if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
				t.Fatalf("Failed to decode: %v", err)
			}

			if body.Code != tc.code {
				t.Errorf("Code = %q, want %q", body.Code, tc.code)
			}
			if body.Message != tc.msg {
				t.Errorf("Message = %q, want %q", body.Message, tc.msg)
			}
		})
	}
}

func TestWriteError_EmptyFields(t *testing.T) {
	w := httptest.NewRecorder()
	WriteError(w, http.StatusOK, "", "")

	var body APIError
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}

	if body.Code != "" {
		t.Errorf("Code = %q, want empty", body.Code)
	}
	if body.Message != "" {
		t.Errorf("Message = %q, want empty", body.Message)
	}
}

func TestWriteError_SpecialCharactersInMessage(t *testing.T) {
	w := httptest.NewRecorder()
	msg := `Failed to parse: "SELECT * FROM users WHERE name = 'O\'Brien'"`
	WriteError(w, http.StatusBadRequest, "PARSE_ERROR", msg)

	var body APIError
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}

	if body.Message != msg {
		t.Errorf("Message = %q, want %q", body.Message, msg)
	}
}

func TestWriteError_LargeMessage(t *testing.T) {
	w := httptest.NewRecorder()
	longMsg := ""
	for i := 0; i < 10000; i++ {
		longMsg += "x"
	}
	WriteError(w, http.StatusInternalServerError, "HUGE_ERROR", longMsg)

	var body APIError
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}

	if len(body.Message) != 10000 {
		t.Errorf("Message length = %d, want 10000", len(body.Message))
	}
}

func TestAPIError_JSONSerialization(t *testing.T) {
	e := APIError{Code: "TEST", Message: "test message"}
	data, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded map[string]string
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded["code"] != "TEST" {
		t.Errorf("code = %q, want TEST", decoded["code"])
	}
	if decoded["message"] != "test message" {
		t.Errorf("message = %q, want 'test message'", decoded["message"])
	}

	// Verify only expected fields exist
	if len(decoded) != 2 {
		t.Errorf("Expected 2 JSON fields, got %d", len(decoded))
	}
}
