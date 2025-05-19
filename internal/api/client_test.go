package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"reflect"
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

					url := "https://test.api.com/landscapes/landscape1/versions/version1/diagrams"
					if req.URL.String() != url {
						return nil, fmt.Errorf("unexpected URL: %s", req.URL.String())
					}

					if tt.wantErr {
						return NewMockResponse(tt.statusCode, tt.respBody), nil
					}
					return NewMockResponse(200, `{"id":"diag123"}`), nil
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

func TestObject_Equal(t *testing.T) {
	tests := []struct {
		name string
		a, b *Object
		want bool
	}{
		{
			name: "identical objects",
			a:    &Object{Name: "A", Desc: "desc", Type: "system", Props: map[string]interface{}{"foo": "bar"}},
			b:    &Object{Name: "A", Desc: "desc", Type: "system", Props: map[string]interface{}{"foo": "bar"}},
			want: true,
		},
		{
			name: "different name",
			a:    &Object{Name: "A", Desc: "desc", Type: "system", Props: map[string]interface{}{}},
			b:    &Object{Name: "B", Desc: "desc", Type: "system", Props: map[string]interface{}{}},
			want: false,
		},
		{
			name: "different desc",
			a:    &Object{Name: "A", Desc: "desc1", Type: "system", Props: map[string]interface{}{}},
			b:    &Object{Name: "A", Desc: "desc2", Type: "system", Props: map[string]interface{}{}},
			want: false,
		},
		{
			name: "different type",
			a:    &Object{Name: "A", Desc: "desc", Type: "system", Props: map[string]interface{}{}},
			b:    &Object{Name: "A", Desc: "desc", Type: "actor", Props: map[string]interface{}{}},
			want: false,
		},
		{
			name: "different props",
			a:    &Object{Name: "A", Desc: "desc", Type: "system", Props: map[string]interface{}{"foo": "bar"}},
			b:    &Object{Name: "A", Desc: "desc", Type: "system", Props: map[string]interface{}{"foo": "baz"}},
			want: false,
		},
		{
			name: "nil vs non-nil",
			a:    nil,
			b:    &Object{Name: "A"},
			want: false,
		},
		{
			name: "both nil",
			a:    nil,
			b:    nil,
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Equal(tt.b); got != tt.want {
				t.Errorf("Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func isInterfaceNil(i interface{}) bool {
	if i == nil {
		return true
	}
	v := reflect.ValueOf(i)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return v.IsNil()
	}
	return false
}

func TestObject_Diff(t *testing.T) {
	a := &Object{Name: "A", Desc: "desc", Type: "system", Props: map[string]interface{}{"foo": "bar", "x": 1}}
	b := &Object{Name: "A", Desc: "desc", Type: "system", Props: map[string]interface{}{"foo": "bar", "x": 1}}
	c := &Object{Name: "B", Desc: "desc2", Type: "actor", Props: map[string]interface{}{"foo": "baz", "y": 2}}
	d := &Object{Name: "A", Desc: "desc", Type: "system", Props: map[string]interface{}{}}

	t.Run("identical objects", func(t *testing.T) {
		diff := a.Diff(b)
		if len(diff) != 0 {
			t.Errorf("Expected no diff, got %v", diff)
		}
	})

	t.Run("different fields and props", func(t *testing.T) {
		diff := a.Diff(c)
		if len(diff) == 0 {
			t.Errorf("Expected diff, got none")
		}
		if diff["Name"][0] != "A" || diff["Name"][1] != "B" {
			t.Errorf("Name diff incorrect: %v", diff["Name"])
		}
		if diff["Desc"][0] != "desc" || diff["Desc"][1] != "desc2" {
			t.Errorf("Desc diff incorrect: %v", diff["Desc"])
		}
		if diff["Type"][0] != "system" || diff["Type"][1] != "actor" {
			t.Errorf("Type diff incorrect: %v", diff["Type"])
		}
		if diff["Props.foo"][0] != "bar" || diff["Props.foo"][1] != "baz" {
			t.Errorf("Props.foo diff incorrect: %v", diff["Props.foo"])
		}
		if diff["Props.x"][0] != 1 || diff["Props.x"][1] != nil {
			t.Errorf("Props.x diff incorrect: %v", diff["Props.x"])
		}
		if diff["Props.y"][0] != nil || diff["Props.y"][1] != 2 {
			t.Errorf("Props.y diff incorrect: %v", diff["Props.y"])
		}
	})

	t.Run("empty vs non-empty props", func(t *testing.T) {
		diff := a.Diff(d)
		if len(diff) == 0 {
			t.Errorf("Expected diff, got none")
		}
		if diff["Props.foo"][0] != "bar" || diff["Props.foo"][1] != nil {
			t.Errorf("Props.foo diff incorrect: %v", diff["Props.foo"])
		}
		if diff["Props.x"][0] != 1 || diff["Props.x"][1] != nil {
			t.Errorf("Props.x diff incorrect: %v", diff["Props.x"])
		}
	})

	t.Run("nil handling", func(t *testing.T) {
		var n *Object
		diff := n.Diff(a)
		obj, ok := diff["nil"][1].(*Object)
		if !isInterfaceNil(diff["nil"][0]) || !ok || obj != a {
			t.Errorf("nil diff incorrect: got [nil %v], want [nil %p] (ok=%v)", diff["nil"][0], a, ok)
		}
	})
}

func TestIcePanelClient_ValidateLandscapeVersion(t *testing.T) {
	ctx := context.Background()

	t.Run("both exist", func(t *testing.T) {
		calls := 0
		mockClient := &MockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				calls++
				return NewMockResponse(200, `{}`), nil
			},
		}
		client := &IcePanelClient{httpClient: mockClient, baseURL: "https://test.api.com"}
		err := client.ValidateLandscapeVersion(ctx, "land1", "ver1")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if calls != 2 {
			t.Errorf("expected 2 calls, got %d", calls)
		}
	})

	t.Run("landscape missing", func(t *testing.T) {
		mockClient := &MockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return NewMockResponse(404, `{}`), nil
			},
		}
		client := &IcePanelClient{httpClient: mockClient, baseURL: "https://test.api.com"}
		err := client.ValidateLandscapeVersion(ctx, "land404", "ver1")
		if err == nil || !strings.Contains(err.Error(), "landscape land404 not found") {
			t.Errorf("expected landscape not found error, got %v", err)
		}
	})

	t.Run("version missing", func(t *testing.T) {
		calls := 0
		mockClient := &MockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				calls++
				if calls == 1 {
					return NewMockResponse(200, `{}`), nil
				}
				return NewMockResponse(404, `{}`), nil
			},
		}
		client := &IcePanelClient{httpClient: mockClient, baseURL: "https://test.api.com"}
		err := client.ValidateLandscapeVersion(ctx, "land1", "ver404")
		if err == nil || !strings.Contains(err.Error(), "version ver404 not found") {
			t.Errorf("expected version not found error, got %v", err)
		}
	})

	t.Run("network error", func(t *testing.T) {
		mockClient := &MockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return nil, fmt.Errorf("network fail")
			},
		}
		client := &IcePanelClient{httpClient: mockClient, baseURL: "https://test.api.com"}
		err := client.ValidateLandscapeVersion(ctx, "land1", "ver1")
		if err == nil || !strings.Contains(err.Error(), "network fail") {
			t.Errorf("expected network error, got %v", err)
		}
	})
}
