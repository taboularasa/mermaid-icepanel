package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"mermaid-icepanel/internal/config"
)

// MockHTTPClient is a mock implementation of HTTPClient for testing.
type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

// NewMockResponse creates a mock HTTP response for testing.
func NewMockResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

func TestNewIcePanelClient(t *testing.T) {
	cfg := &config.Config{
		APIBaseURL:     "https://test.api.com",
		DefaultToken:   "default-token",
		RequestTimeout: 0,
	}

	tests := []struct {
		name      string
		token     string
		wantToken string
		wantURL   string
	}{
		{
			name:      "with provided token",
			token:     "provided-token",
			wantToken: "provided-token",
			wantURL:   "https://test.api.com",
		},
		{
			name:      "fallback to default token",
			token:     "",
			wantToken: "default-token",
			wantURL:   "https://test.api.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{}
			client := NewIcePanelClient(cfg, mockClient, tt.token)

			if client.Token != tt.wantToken {
				t.Errorf("NewIcePanelClient() token = %v, want %v", client.Token, tt.wantToken)
			}

			if client.baseURL != tt.wantURL {
				t.Errorf("NewIcePanelClient() baseURL = %v, want %v", client.baseURL, tt.wantURL)
			}
		})
	}
}

func TestIcePanelClient_PostDiagram(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		statusCode int
		respBody   string
		wantErr    bool
	}{
		{
			name:       "successful post",
			statusCode: 200,
			respBody:   `{"id":"123"}`,
			wantErr:    false,
		},
		{
			name:       "error response",
			statusCode: 400,
			respBody:   `{"error":"Bad request"}`,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					// Validate request
					if req.Method != http.MethodPost {
						return nil, fmt.Errorf("expected POST method, got %s", req.Method)
					}

					url := "https://api.test/landscapes/land123/versions/ver123/diagrams"
					if req.URL.String() != url {
						return nil, fmt.Errorf("unexpected URL: %s", req.URL.String())
					}

					return NewMockResponse(http.StatusCreated, `{"id":"diag123"}`), nil
				},
			}

			client := &IcePanelClient{
				httpClient: mockClient,
				Token:      "test-token",
				baseURL:    "https://test.api.com",
			}

			diagram := &Diagram{
				Name: "Test Diagram",
				Type: "app-diagram",
				Objects: []*Object{
					{Handle: "h001", Name: "Test Object", Type: "system"},
				},
			}

			err := client.PostDiagram(ctx, "landscape1", "version1", diagram, false)

			if (err != nil) != tt.wantErr {
				t.Errorf("PostDiagram() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIcePanelClient_WipeVersion(t *testing.T) {
	ctx := context.Background()

	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			switch req.Method {
			case http.MethodGet:
				// Return a valid response for listIDs calls
				return NewMockResponse(http.StatusOK, `{"data":[{"id":"item1"},{"id":"item2"}]}`), nil
			case http.MethodDelete:
				// Return success for delete calls
				return NewMockResponse(http.StatusNoContent, ""), nil
			default:
				return nil, fmt.Errorf("unexpected method: %s", req.Method)
			}
		},
	}

	client := &IcePanelClient{
		httpClient: mockClient,
		Token:      "test-token",
		baseURL:    "https://test.api.com",
	}

	err := client.WipeVersion(ctx, "landscape1", "version1")
	if err != nil {
		t.Errorf("WipeVersion() unexpected error = %v", err)
	}
}

func TestNewMockResponse(t *testing.T) {
	resp := NewMockResponse(http.StatusOK, "test body")
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read body: %v", err)
	}

	// Close the response body
	if err := resp.Body.Close(); err != nil {
		t.Fatalf("Failed to close response body: %v", err)
	}

	if string(bodyBytes) != "test body" {
		t.Errorf("Expected body 'test body', got '%s'", string(bodyBytes))
	}
}
