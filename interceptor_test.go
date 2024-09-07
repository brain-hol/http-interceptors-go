package interceptor

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"testing"
)

type mockRoundTripper struct {
	Response *http.Response
	Err      error
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.Response, m.Err
}

func TestBaseURLInterceptor(t *testing.T) {
	mockResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString("OK")),
	}

	mockRT := &mockRoundTripper{
		Response: mockResp,
	}

	baseURL, err := url.Parse("http://base.example.com/am")
	if err != nil {
		t.Fatal(err)
	}
	interceptor := BaseURL(*baseURL)(mockRT)

	tests := []struct {
		originalURL string
		expectedURL string
	}{
		{"/oauth2", "http://base.example.com/am/oauth2"},
		{"/oauth2/json", "http://base.example.com/am/oauth2/json"},
		{"oauth2/json", "http://base.example.com/am/oauth2/json"},
		{"https://google.com", "https://google.com"},
		{"https://google.com?test=asfd#first", "https://google.com?test=asfd#first"},
		{"../openidm", "http://base.example.com/openidm"},
		{"../openidm/query", "http://base.example.com/openidm/query"},
		{"../../openidm", "http://base.example.com/openidm"},
		{"../../../../openidm", "http://base.example.com/openidm"},
		{"../am/../other/openidm", "http://base.example.com/other/openidm"},
		{"/../am/../other/openidm", "http://base.example.com/other/openidm"},
		{"/oauth2/json?param1=value1&param2=value2#fragment", "http://base.example.com/am/oauth2/json?param1=value1&param2=value2#fragment"},
	}

	for _, test := range tests {
		req, err := http.NewRequest("GET", test.originalURL, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		resp, err := interceptor.RoundTrip(req)
		if err != nil {
			t.Fatalf("Failed to perform request: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}

		if req.URL.String() != test.expectedURL {
			t.Errorf("Expected URL to be '%s', got '%s'", test.expectedURL, req.URL.String())
		}
	}
}

func TestHeaderInterceptor(t *testing.T) {
	mockResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString("OK")),
	}

	mockRT := &mockRoundTripper{
		Response: mockResp,
	}

	tests := []struct {
		headerKey   string
		headerValue string
	}{
		{"Authorization", "Bearer example-token"},
		{"Random", "random value"},
		{"", "random value"},
	}

	for _, test := range tests {
		interceptor := Header(test.headerKey, test.headerValue)(mockRT)

		req, err := http.NewRequest("GET", "http://example.com", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		resp, err := interceptor.RoundTrip(req)
		if err != nil {
			t.Fatalf("Failed to perform request: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}

		got := req.Header.Get(test.headerKey)
		if test.headerKey != "" {
			if got != test.headerValue {
				t.Errorf("Expected header to be '%s', got '%s'", test.headerValue, got)
			}
		} else {
			if got != "" {
				t.Errorf("Expected header to not be set if no key was provided: got '%s'", got)
			}
		}
	}
}
