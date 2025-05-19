package generator

import (
	"reflect"
	"testing"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Mock proto file descriptor for testing
type mockProtoFile struct {
	protogen.File
	path        string
	pkg         protoreflect.FullName
	svcs        []*mockService
	isGenerated bool
}

type mockService struct {
	name    protoreflect.Name
	comment string
}

func (m *mockProtoFile) Desc() protogen.FileDescriptor {
	return &mockFileDescriptor{
		path: m.path,
		pkg:  m.pkg,
	}
}

func (m *mockProtoFile) Services() []*protogen.Service {
	services := make([]*protogen.Service, len(m.svcs))
	for i, svc := range m.svcs {
		services[i] = &protogen.Service{
			Comments: protogen.CommentSet{
				Leading: protogen.Comments(svc.comment),
			},
			Desc: &mockServiceDescriptor{
				name: svc.name,
			},
		}
	}
	return services
}

func (m *mockProtoFile) Generate() bool {
	return m.isGenerated
}

type mockFileDescriptor struct {
	protogen.FileDescriptor
	path string
	pkg  protoreflect.FullName
}

func (m *mockFileDescriptor) Path() string {
	return m.path
}

func (m *mockFileDescriptor) Package() protoreflect.FullName {
	return m.pkg
}

type mockServiceDescriptor struct {
	protogen.ServiceDescriptor
	name protoreflect.Name
}

func (m *mockServiceDescriptor) Name() protoreflect.Name {
	return m.name
}

func TestProcessProtoFile(t *testing.T) {
	tests := []struct {
		name     string
		file     *mockProtoFile
		expected []C4Object
	}{
		{
			name: "basic file with one service",
			file: &mockProtoFile{
				path:        "example/service.proto",
				pkg:         "example.service",
				isGenerated: true,
				svcs: []*mockService{
					{
						name:    "UserService",
						comment: "// User management service",
					},
				},
			},
			expected: []C4Object{
				{
					ID:          "boundary-example.service",
					Name:        "example.service",
					Description: "Package: example.service",
					Type:        C4SystemBoundary,
					Package:     "example.service",
				},
				{
					ID:          "service-UserService",
					Name:        "UserService",
					Description: "// User management service",
					Type:        C4System,
					Package:     "example.service",
				},
			},
		},
		{
			name: "file with multiple services",
			file: &mockProtoFile{
				path:        "example/multi.proto",
				pkg:         "example.multi",
				isGenerated: true,
				svcs: []*mockService{
					{
						name:    "UserService",
						comment: "// User management service",
					},
					{
						name:    "AuthService",
						comment: "// Authentication service",
					},
				},
			},
			expected: []C4Object{
				{
					ID:          "boundary-example.multi",
					Name:        "example.multi",
					Description: "Package: example.multi",
					Type:        C4SystemBoundary,
					Package:     "example.multi",
				},
				{
					ID:          "service-UserService",
					Name:        "UserService",
					Description: "// User management service",
					Type:        C4System,
					Package:     "example.multi",
				},
				{
					ID:          "service-AuthService",
					Name:        "AuthService",
					Description: "// Authentication service",
					Type:        C4System,
					Package:     "example.multi",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := processProtoFile(tt.file)
			if err != nil {
				t.Fatalf("processProtoFile failed: %v", err)
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("processProtoFile result mismatch\nwant: %+v\ngot:  %+v", tt.expected, result)
			}
		})
	}
}

func TestDetermineServiceType(t *testing.T) {
	tests := []struct {
		name     string
		service  string
		comment  string
		expected C4ObjectType
	}{
		{
			name:     "default internal system",
			service:  "UserService",
			comment:  "",
			expected: C4System,
		},
		{
			name:     "external system by name",
			service:  "ExternalUserService",
			comment:  "",
			expected: C4SystemExt,
		},
		{
			name:     "database system by name",
			service:  "UserDatabaseService",
			comment:  "",
			expected: C4SystemDb,
		},
		{
			name:     "external system by comment",
			service:  "UserService",
			comment:  "This service integrates with an external third-party system",
			expected: C4SystemExt,
		},
		{
			name:     "database system by comment",
			service:  "UserService",
			comment:  "This service manages the database storage for users",
			expected: C4SystemDb,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineServiceType(tt.service, tt.comment)
			if result != tt.expected {
				t.Errorf("determineServiceType(%s, %s) = %s, want %s",
					tt.service, tt.comment, result, tt.expected)
			}
		})
	}
}
