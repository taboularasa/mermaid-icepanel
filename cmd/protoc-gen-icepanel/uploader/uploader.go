// Package uploader provides functionality to upload generated IcePanel objects
// to the IcePanel API, with the option to wipe the version first.
package uploader

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"mermaid-icepanel/internal/api"
	"mermaid-icepanel/internal/config"
)

// ObjectsFile represents the structure of the generated icepanel_objects.json file.
type ObjectsFile struct {
	Config struct {
		LandscapeID string `json:"landscapeId"`
		VersionID   string `json:"versionId"`
		Wipe        bool   `json:"wipe"`
	} `json:"config"`
	Objects []Object `json:"objects"`
}

// Object represents an IcePanel object in the objects file.
type Object struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Package     string `json:"package"`
}

// UploadOptions contains options for the Upload function.
type UploadOptions struct {
	FilePath       string
	Token          string
	Verbose        bool
	DryRun         bool
	ForceLandscape string
	ForceVersion   string
}

// Upload reads the generated objects file and uploads the objects to IcePanel.
func Upload(ctx context.Context, options UploadOptions) error {
	// Read and parse the objects file
	file, err := os.Open(options.FilePath)
	if err != nil {
		return fmt.Errorf("failed to open objects file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.Printf("Warning: error closing file: %v", closeErr)
		}
	}()

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read objects file: %w", err)
	}

	var objectsFile ObjectsFile
	if err := json.Unmarshal(data, &objectsFile); err != nil {
		return fmt.Errorf("failed to parse objects file: %w", err)
	}

	// Override landscape/version if provided
	if options.ForceLandscape != "" {
		objectsFile.Config.LandscapeID = options.ForceLandscape
	}
	if options.ForceVersion != "" {
		objectsFile.Config.VersionID = options.ForceVersion
	}

	// Validate configuration
	if objectsFile.Config.LandscapeID == "" || objectsFile.Config.VersionID == "" {
		return fmt.Errorf("missing landscape ID or version ID in configuration")
	}

	// Create IcePanel client
	cfg := config.NewConfig()

	// Use provided token or fall back to config
	token := options.Token
	if token == "" {
		token = cfg.DefaultToken
	}

	httpClient := &api.DefaultHTTPClient{
		Client: getHTTPClient(),
	}

	icepanelClient := api.NewIcePanelClient(cfg, httpClient, token)

	// Validate landscape and version
	if err := icepanelClient.ValidateLandscapeVersion(
		ctx, objectsFile.Config.LandscapeID, objectsFile.Config.VersionID); err != nil {
		return fmt.Errorf("failed to validate landscape/version: %w", err)
	}

	// Process wipe request if needed
	if err := handleWipeIfNeeded(ctx, icepanelClient, objectsFile.Config, options); err != nil {
		return err
	}

	// Upload each object
	if options.Verbose {
		log.Printf("Uploading %d objects to landscape %s, version %s",
			len(objectsFile.Objects), objectsFile.Config.LandscapeID, objectsFile.Config.VersionID)
	}

	for _, obj := range objectsFile.Objects {
		// Convert to IcePanel API object format
		icepanelObj := &api.Object{
			Handle: obj.ID,
			Name:   obj.Name,
			Desc:   obj.Description,
			Type:   obj.Type,
			Props: map[string]interface{}{
				"package": obj.Package,
			},
		}

		// Create the object
		if options.Verbose {
			log.Printf("Creating object: %s (%s)", obj.Name, obj.Type)
		}

		if !options.DryRun {
			if err := icepanelClient.CreateObject(ctx, objectsFile.Config.LandscapeID,
				objectsFile.Config.VersionID, icepanelObj, false); err != nil {
				return fmt.Errorf("failed to create object %s: %w", obj.ID, err)
			}
		}
	}

	return nil
}

// handleWipeIfNeeded performs a version wipe if requested.
func handleWipeIfNeeded(ctx context.Context, client *api.IcePanelClient,
	config struct {
		LandscapeID string `json:"landscapeId"`
		VersionID   string `json:"versionId"`
		Wipe        bool   `json:"wipe"`
	},
	options UploadOptions,
) error {
	if !config.Wipe {
		return nil
	}

	if options.Verbose {
		log.Printf("Wiping existing content in landscape %s, version %s",
			config.LandscapeID, config.VersionID)
	}

	if !options.DryRun {
		if err := client.WipeVersion(ctx, config.LandscapeID, config.VersionID); err != nil {
			return fmt.Errorf("failed to wipe version: %w", err)
		}
	} else {
		log.Printf("[DRY RUN] Would wipe landscape %s, version %s",
			config.LandscapeID, config.VersionID)
	}

	return nil
}

// getHTTPClient returns the HTTP client to use for API requests.
func getHTTPClient() *http.Client {
	return http.DefaultClient
}
