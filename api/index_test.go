package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler(t *testing.T) {
	tests := []struct {
		name           string
		queryParam     string
		expectedStatus int
	}{
		{
			name:           "Missing username",
			queryParam:     "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Valid username",
			queryParam:     "McCzarny",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/?username="+tt.queryParam, nil)
			w := httptest.NewRecorder()

			Handler(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
			if tt.expectedStatus == http.StatusOK {
				// Check if the content type is SVG
				contentType := resp.Header.Get("Content-Type")
				if contentType != "image/svg+xml" {
					t.Errorf("expected content type 'image/svg+xml', got '%s'", contentType)
				}
			}
			// Check if the response body is not empty
			if resp.StatusCode == http.StatusOK {
				body := w.Body.String()
				if body == "" {
					t.Error("expected non-empty response body")
				}
			}
		})
	}
}
