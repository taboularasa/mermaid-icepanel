// Package main provides a CLI tool to upload the generated IcePanel objects
// to the IcePanel API, with support for wiping the version before uploading.
package main

import (
	"context"
	"flag"
	"log"
	"time"

	"mermaid-icepanel/cmd/protoc-gen-icepanel/uploader"
)

func main() {
	// Parse command-line arguments
	filePath := flag.String("file", "icepanel_objects.json", "Path to the generated objects file")
	token := flag.String("token", "", "IcePanel API token (falls back to ICEPANEL_TOKEN env var)")
	landscapeID := flag.String("landscape", "", "Override landscape ID from the file")
	versionID := flag.String("version", "", "Override version ID from the file")
	dryRun := flag.Bool("dry-run", false, "Dry run mode (don't actually upload)")
	verbose := flag.Bool("v", false, "Verbose output")
	timeout := flag.Int("timeout", 30, "Request timeout in seconds")
	flag.Parse()

	// Create upload options
	options := uploader.UploadOptions{
		FilePath:       *filePath,
		Token:          *token,
		Verbose:        *verbose,
		DryRun:         *dryRun,
		ForceLandscape: *landscapeID,
		ForceVersion:   *versionID,
	}

	// Set up context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*timeout)*time.Second)
	defer cancel()

	// Upload the objects
	if *verbose {
		log.Printf("Uploading objects from %s", *filePath)
	}

	err := uploader.Upload(ctx, options)
	// Always execute cancel, even when errors occur
	cancel()

	if err != nil {
		log.Fatalf("Error uploading objects: %v", err)
	}

	if *verbose {
		log.Printf("Upload completed successfully")
	}
}
