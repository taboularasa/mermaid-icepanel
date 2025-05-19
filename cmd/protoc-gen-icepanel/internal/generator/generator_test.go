// Package generator contains tests for the generator logic for IcePanel integration.
package generator

import (
	"reflect"
	"testing"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// TestProcessProtoFile tests the processProtoFile function with real protogen.File instances
func TestProcessProtoFile(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		pkg      string
		services []struct {
			name    string
			comment string
		}
		speculativePathPrefix string
		expected              []C4Object
	}{
		{
			name:     "basic file with one service",
			fileName: "example/service.proto",
			pkg:      "example.service",
			services: []struct {
				name    string
				comment string
			}{
				{
					name:    "UserService",
					comment: "// User management service",
				},
			},
			speculativePathPrefix: "",
			expected: []C4Object{
				{
					ID:            "boundary-example.service",
					Name:          "example.service",
					Description:   "Package: example.service",
					Type:          C4SystemBoundary,
					Package:       "example.service",
					IsSpeculative: false,
				},
				{
					ID:            "service-UserService",
					Name:          "UserService",
					Description:   "//// User management service\n",
					Type:          C4System,
					Package:       "example.service",
					IsSpeculative: false,
				},
			},
		},
		{
			name:     "file with multiple services",
			fileName: "example/multi.proto",
			pkg:      "example.multi",
			services: []struct {
				name    string
				comment string
			}{
				{
					name:    "UserService",
					comment: "// User management service",
				},
				{
					name:    "AuthService",
					comment: "// Authentication service",
				},
			},
			speculativePathPrefix: "",
			expected: []C4Object{
				{
					ID:            "boundary-example.multi",
					Name:          "example.multi",
					Description:   "Package: example.multi",
					Type:          C4SystemBoundary,
					Package:       "example.multi",
					IsSpeculative: false,
				},
				{
					ID:            "service-UserService",
					Name:          "UserService",
					Description:   "//// User management service\n",
					Type:          C4System,
					Package:       "example.multi",
					IsSpeculative: false,
				},
				{
					ID:            "service-AuthService",
					Name:          "AuthService",
					Description:   "//// Authentication service\n",
					Type:          C4System,
					Package:       "example.multi",
					IsSpeculative: false,
				},
			},
		},
		{
			name:     "speculative proto file",
			fileName: "tdd/protos/speculative_service.proto",
			pkg:      "tdd.protos.speculative",
			services: []struct {
				name    string
				comment string
			}{
				{
					name:    "SpeculativeService",
					comment: "// Speculative service for TDD",
				},
			},
			speculativePathPrefix: "tdd/protos/",
			expected: []C4Object{
				{
					ID:            "boundary-tdd.protos.speculative",
					Name:          "tdd.protos.speculative",
					Description:   "Package: tdd.protos.speculative",
					Type:          C4SystemBoundary,
					Package:       "tdd.protos.speculative",
					IsSpeculative: true,
				},
				{
					ID:            "service-SpeculativeService",
					Name:          "SpeculativeService",
					Description:   "//// Speculative service for TDD\n",
					Type:          C4System,
					Package:       "tdd.protos.speculative",
					IsSpeculative: true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a file with the descriptor
			file := &protogen.File{
				GoPackageName: protogen.GoPackageName(tt.pkg),
				Generate:      true,
			}

			// Create the file descriptor
			file.Desc = createFileDescriptor(tt.fileName, tt.pkg)

			// Create services for the file
			for _, svc := range tt.services {
				serviceDesc := createServiceDescriptor(file.Desc, svc.name)
				service := &protogen.Service{
					Comments: protogen.CommentSet{
						Leading: protogen.Comments(svc.comment),
					},
					Desc: serviceDesc,
				}
				file.Services = append(file.Services, service)
			}

			// Process the file
			result := processProtoFile(file, tt.speculativePathPrefix)

			// Check that the result matches expectations
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("processProtoFile result mismatch\nwant: %+v\ngot:  %+v", tt.expected, result)
			}
		})
	}
}

// Helper function to create a file descriptor for testing
func createFileDescriptor(path string, pkg string) protoreflect.FileDescriptor {
	return &testFileDescriptor{
		path: path,
		pkg:  protoreflect.FullName(pkg),
	}
}

// Helper function to create a service descriptor for testing
func createServiceDescriptor(file protoreflect.FileDescriptor, name string) protoreflect.ServiceDescriptor {
	return &testServiceDescriptor{
		file: file,
		name: protoreflect.Name(name),
	}
}

// Test implementation of FileDescriptor
type testFileDescriptor struct {
	protoreflect.FileDescriptor
	path string
	pkg  protoreflect.FullName
}

func (fd *testFileDescriptor) Path() string {
	return fd.path
}

func (fd *testFileDescriptor) Package() protoreflect.FullName {
	return fd.pkg
}

// Test implementation of ServiceDescriptor
type testServiceDescriptor struct {
	protoreflect.ServiceDescriptor
	file protoreflect.FileDescriptor
	name protoreflect.Name
}

func (sd *testServiceDescriptor) Name() protoreflect.Name {
	return sd.name
}

func (sd *testServiceDescriptor) Parent() protoreflect.Descriptor {
	return sd.file
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
