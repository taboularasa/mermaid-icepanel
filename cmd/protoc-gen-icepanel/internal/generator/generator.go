// Package generator provides logic for extracting C4 model objects from proto files for IcePanel integration.
package generator

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

// C4ObjectType defines the type of C4 model object.
type C4ObjectType string

const (
	C4System         C4ObjectType = "System"          // Internal system.
	C4SystemExt      C4ObjectType = "System_Ext"      // External system.
	C4SystemDb       C4ObjectType = "SystemDb"        // Database system.
	C4SystemBoundary C4ObjectType = "System_Boundary" // Package/namespace boundary.
)

// C4Object represents an object in the C4 model.
type C4Object struct {
	ID            string       // Unique identifier.
	Name          string       // Display name.
	Description   string       // Description/documentation.
	Type          C4ObjectType // Object type.
	Technology    string       // Technology stack (if applicable).
	Package       string       // Package/namespace.
	IsSpeculative bool         // True if derived from tdd/protos/.
}

// Generate processes the CodeGeneratorRequest and returns a CodeGeneratorResponse.
func Generate(req *pluginpb.CodeGeneratorRequest) (*pluginpb.CodeGeneratorResponse, error) {
	plugin, err := protogen.Options{}.New(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create protogen plugin: %w", err)
	}

	// Parse parameters.
	speculativePathPrefix := ""
	params := plugin.Request.GetParameter()
	if params != "" {
		parts := strings.Split(params, ",")
		for _, part := range parts {
			kv := strings.SplitN(part, "=", 2)
			if len(kv) == 2 && kv[0] == "speculative_protos_path_prefix" {
				speculativePathPrefix = kv[1]
				break
			}
		}
	}

	plugin.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)

	resp := &pluginpb.CodeGeneratorResponse{
		SupportedFeatures: &plugin.SupportedFeatures,
	}

	// Track all extracted objects.
	objects := make([]C4Object, 0)

	// Process each proto file.
	for _, file := range plugin.Files {
		if !file.Generate {
			continue
		}

		// Extract objects from proto file.
		fileObjects := processProtoFile(file, speculativePathPrefix)
		objects = append(objects, fileObjects...)
	}

	// Generate output file with IcePanel API calls.
	if len(objects) > 0 {
		content := generateIcePanelOutput(objects)
		resp.File = append(resp.File, &pluginpb.CodeGeneratorResponse_File{
			Name:    stringPtr("icepanel_objects.json"),
			Content: stringPtr(content),
		})
	}

	return resp, nil
}

// processProtoFile extracts C4 objects from a proto file.
func processProtoFile(file *protogen.File, speculativePathPrefix string) []C4Object {
	objects := make([]C4Object, 0)

	isSpeculative := speculativePathPrefix != "" && strings.HasPrefix(file.Desc.Path(), speculativePathPrefix)

	// Process package as system boundary.
	packageName := string(file.Desc.Package())
	if packageName != "" {
		boundary := C4Object{
			ID:            "boundary-" + packageName,
			Name:          packageName,
			Description:   "Package: " + packageName,
			Type:          C4SystemBoundary,
			Package:       packageName,
			IsSpeculative: isSpeculative,
		}
		objects = append(objects, boundary)
	}

	// Process services.
	for _, service := range file.Services {
		// Extract service metadata.
		serviceName := string(service.Desc.Name())
		serviceComment := service.Comments.Leading.String()

		// Determine service type based on naming convention and comments.
		objectType := determineServiceType(serviceName, serviceComment)

		// Create service object.
		serviceObj := C4Object{
			ID:            "service-" + serviceName,
			Name:          serviceName,
			Description:   serviceComment,
			Type:          objectType,
			Package:       packageName,
			IsSpeculative: isSpeculative,
		}
		objects = append(objects, serviceObj)
	}

	return objects
}

// determineServiceType categorizes services based on naming conventions and comments.
func determineServiceType(serviceName string, comment string) C4ObjectType {
	return ClassifyService(serviceName, comment)
}

// generateIcePanelOutput formats objects for IcePanel import.
func generateIcePanelOutput(objects []C4Object) string {
	// This is a placeholder for actual output formatting.
	// In a real implementation, we would create JSON that can be
	// consumed by the IcePanel API.

	// For now, just create a JSON array of objects.
	output := "[\n"
	for i, obj := range objects {
		output += "  {\n"
		output += fmt.Sprintf("    \"id\": \"%s\",\n", obj.ID)
		output += fmt.Sprintf("    \"name\": \"%s\",\n", obj.Name)
		output += fmt.Sprintf("    \"description\": \"%s\",\n", obj.Description)
		output += fmt.Sprintf("    \"type\": \"%s\",\n", obj.Type)
		output += fmt.Sprintf("    \"package\": \"%s\",\n", obj.Package)
		output += fmt.Sprintf("    \"isSpeculative\": %t\n", obj.IsSpeculative)
		output += "  }"
		if i < len(objects)-1 {
			output += ",\n"
		} else {
			output += "\n"
		}
	}
	output += "]\n"

	return output
}

// stringPtr creates a pointer to a string.
func stringPtr(s string) *string {
	return &s
}
