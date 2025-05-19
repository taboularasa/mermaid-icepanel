package main

import (
	"fmt"
	"io"
	"os"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"

	"mermaid-icepanel/cmd/protoc-gen-icepanel/internal/generator"
)

func printUsage() {
	fmt.Fprintf(os.Stdout, `protoc-gen-icepanel: IcePanel C4 object generator plugin for protoc

USAGE:
  protoc --icepanel_out=speculative_protos_path_prefix=DIR:.

This plugin is intended to be run by protoc. It reads a CodeGeneratorRequest from stdin and writes a CodeGeneratorResponse to stdout.

Plugin options:
  speculative_protos_path_prefix=DIR   Mark proto files under DIR as speculative (for TDD workflows)

For more information, see the README or run with -h/--help.
`)
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
