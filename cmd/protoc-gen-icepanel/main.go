package main

import (
	"fmt"
	"io"
	"os"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"

	"mermaid-icepanel/cmd/protoc-gen-icepanel/internal/generator"
)

func main() {
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
