// Package main provides the protoc-gen-icepanel tool, a protoc plugin for IcePanel object generation.
package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"mermaid-icepanel/cmd/protoc-gen-icepanel/internal/generator"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

// protoc-gen-icepanel is a protoc plugin for generating IcePanel objects from proto files.
//
// Parameters:
//   - landscape=<id>: The IcePanel landscape ID
//   - version=<id>: The IcePanel version ID
//   - wipe=true|false: Whether to wipe existing content before importing
//   - speculative_protos_path_prefix=DIR: Mark proto files under DIR as speculative (for TDD workflows)
//
// Usage:
//   protoc --icepanel_out=. \
//          --icepanel_opt=landscape=123,version=456,wipe=true \
//          example.proto
//
// This will:
//   1. Extract objects from the proto files
//   2. Generate a file named "icepanel_objects.json"
//   3. Include wipe=true in the output to indicate the version should be wiped before pushing

func printUsage() {
	_, err := fmt.Fprintf(os.Stdout, `protoc-gen-icepanel: IcePanel C4 object generator plugin for protoc

USAGE:
  protoc --icepanel_out=<options>:.

This plugin is intended to be run by protoc. It reads a CodeGeneratorRequest 
from stdin and writes a CodeGeneratorResponse to stdout.

Plugin options:
  landscape=<id>                       IcePanel landscape ID
  version=<id>                         IcePanel version ID
  wipe=true|false                      Whether to wipe existing content before importing
  speculative_protos_path_prefix=DIR   Mark proto files under DIR as speculative (for TDD workflows)

For more information, see the README or run with -h/--help.
`)
	if err != nil {
		log.Fatalf("Error writing usage information: %v", err)
	}
}

func main() {
	for _, arg := range os.Args[1:] {
		if arg == "-h" || arg == "--help" {
			printUsage()
			os.Exit(0)
		}
	}

	// Check if stdin is a terminal (not being run by protoc)
	fi, err := os.Stdin.Stat()
	if err == nil && (fi.Mode()&os.ModeCharDevice) != 0 {
		fmt.Fprintln(os.Stderr, "protoc-gen-icepanel: This plugin is intended to be run by protoc. Use -h for help.")
		os.Exit(2)
	}

	// Read request from stdin
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read input: %v\n", err)
		os.Exit(1)
	}

	// Parse request
	req := &pluginpb.CodeGeneratorRequest{}
	if err := proto.Unmarshal(data, req); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse input: %v\n", err)
		os.Exit(1)
	}

	// Generate code
	resp, err := generator.Generate(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to generate code: %v\n", err)
		os.Exit(1)
	}

	// Marshal response
	data, err = proto.Marshal(resp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal response: %v\n", err)
		os.Exit(1)
	}

	// Write response to stdout
	if _, err := os.Stdout.Write(data); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write response: %v\n", err)
		os.Exit(1)
	}
}
