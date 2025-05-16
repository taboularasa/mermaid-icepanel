// icepanel_sync.go
// CLI tool: convert Mermaid C4 (System‑Context subset) to an IcePanel diagram and optionally wipe version.
// Build: `go build -o icepanel-sync main.go`
// Usage:
//
//	icepanel-sync -mmd proveout.mmd -landscape 123 -version 456 \
//	    -token $ICEPANEL_TOKEN -name "Proveout System Context" -wipe -v
package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	"mermaid-icepanel/internal/api"
	"mermaid-icepanel/internal/config"
	"mermaid-icepanel/internal/parser"
)

// ---------- main ----------

func run() error {
	// Define command line flags
	mmdFile := flag.String("mmd", "", "Path to Mermaid .mmd file")
	landscapeID := flag.String("landscape", "", "IcePanel landscape ID")
	versionID := flag.String("version", "", "IcePanel version ID")
	diagramName := flag.String("name", "Imported diagram", "Diagram name")
	token := flag.String("token", "", "API token (falls back to ICEPANEL_TOKEN env var)")
	wipe := flag.Bool("wipe", false, "Delete existing content before import")
	verbose := flag.Bool("v", false, "Verbose output")
	flag.Parse()

	// Check required fields
	if *mmdFile == "" || *landscapeID == "" || *versionID == "" {
		flag.Usage()
		return &requiredFieldError{msg: "Required fields: -mmd, -landscape, -version"}
	}

	// Configuration
	cfg := config.NewConfig()

	// Check token before setting up context
	if *token == "" && cfg.DefaultToken == "" {
		return &tokenError{
			msg: "API token is required. Provide it with -token flag " +
				"or set ICEPANEL_TOKEN environment variable",
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.RequestTimeout)
	defer cancel()

	httpClient := &api.DefaultHTTPClient{
		Client: http.DefaultClient,
	}
	icepanelClient := api.NewIcePanelClient(cfg, httpClient, *token)

	// Parse mermaid file
	diagram, err := parser.ParseMermaid(&parser.DefaultFileReader{}, *mmdFile)
	if err != nil {
		return err
	}

	// Set diagram name from command line
	diagram.Name = *diagramName

	// Wipe existing content if requested
	if *wipe {
		if *verbose {
			log.Printf("Wiping existing content in landscape %s, version %s", *landscapeID, *versionID)
		}
		if err := icepanelClient.WipeVersion(ctx, *landscapeID, *versionID); err != nil {
			return err
		}
	}

	// Upload diagram
	if *verbose {
		log.Printf("Uploading diagram '%s' with %d objects and %d connections",
			diagram.Name, len(diagram.Objects), len(diagram.Connections))
	}

	if err := icepanelClient.PostDiagram(ctx, *landscapeID, *versionID, diagram, *verbose); err != nil {
		return err
	}

	if *verbose {
		log.Println("Import completed successfully")
	}

	return nil
}

// Define custom errors.
type requiredFieldError struct {
	msg string
}

func (e *requiredFieldError) Error() string {
	return e.msg
}

type tokenError struct {
	msg string
}

func (e *tokenError) Error() string {
	return e.msg
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("ERROR: %v", err)
	}
}
