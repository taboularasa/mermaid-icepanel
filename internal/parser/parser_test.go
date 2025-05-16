package parser

import (
	"io"
	"strings"
	"testing"
)

// MockFileReader implements FileReader for testing.
type MockFileReader struct {
	MockData string
	MockErr  error
}

// ReadFile implements FileReader interface for testing.
func (m *MockFileReader) ReadFile(_ string) (ReadCloser, error) {
	if m.MockErr != nil {
		return nil, m.MockErr
	}
	return &mockReadCloser{strings.NewReader(m.MockData)}, nil
}

// mockReadCloser implements ReadCloser for testing.
type mockReadCloser struct {
	io.Reader
}

func (m *mockReadCloser) Close() error {
	return nil
}

func TestParseMermaid(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantObjs int
		wantRels int
		wantErr  bool
	}{
		{
			name: "Simple diagram",
			content: `
Person(user, "User", "System user")
System(app, "Application", "Core system")
System_Ext(api, "External API", "Third-party service")
Rel(user, app, "Uses")
Rel(app, api, "Calls")
`,
			wantObjs: 3,
			wantRels: 2,
			wantErr:  false,
		},
		{
			name: "With boundary",
			content: `
System_Boundary(b1, "Boundary") {
  Person(user, "User", "System user")
  System(app, "Application", "Core system")
}
System_Ext(api, "External API", "Third-party service")
BiRel(app, api, "Syncs")
`,
			wantObjs: 4, // 3 elements + boundary
			wantRels: 2, // BiRel creates 2 relationships
			wantErr:  false,
		},
		{
			name:     "Empty content",
			content:  "",
			wantObjs: 0,
			wantRels: 0,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockReader := &MockFileReader{MockData: tt.content}
			got, err := ParseMermaid(mockReader, "dummy.mmd")

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMermaid() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			if len(got.Objects) != tt.wantObjs {
				t.Errorf("ParseMermaid() got %d objects, want %d", len(got.Objects), tt.wantObjs)
			}

			if len(got.Connections) != tt.wantRels {
				t.Errorf("ParseMermaid() got %d connections, want %d", len(got.Connections), tt.wantRels)
			}
		})
	}
}

func TestSlug(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Test", "test"},
		{"Test System", "test-system"},
		{"API Service", "api-service"},
		{"Multiple   Spaces", "multiple---spaces"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := slug(tt.input); got != tt.want {
				t.Errorf("slug() = %v, want %v", got, tt.want)
			}
		})
	}
}
