#!/usr/bin/env just --justfile
# Load environment variables from .env file
set dotenv-load

# List available commands
default:
    @just --list

# Build the binary
build:
    go build -o mermaid-icepanel main.go

# Run the CLI with arguments passed to it
run *ARGS:
    go run main.go {{ARGS}}

# Import a mermaid diagram to IcePanel
import MERMAID_FILE LANDSCAPE_ID VERSION_ID NAME="Imported Diagram" WIPE="":
    #!/usr/bin/env bash
    set -euo pipefail
    WIPE_FLAG=""
    if [ "{{WIPE}}" = "wipe" ]; then
        WIPE_FLAG="-wipe"
    fi
    go run main.go -mmd {{MERMAID_FILE}} -landscape {{LANDSCAPE_ID}} -version {{VERSION_ID}} -name "{{NAME}}" ${WIPE_FLAG}

# Run with full set of arguments for direct control
sync *ARGS:
    go run main.go {{ARGS}}

# Build and install the protoc-gen-icepanel plugin
build-plugin:
    go build -o $GOPATH/bin/protoc-gen-icepanel ./cmd/protoc-gen-icepanel

# Build the uploader tool
build-uploader:
    go build -o uploader ./cmd/protoc-gen-icepanel/upload

# Generate IcePanel objects from proto files
generate-objects PROTO_FILES LANDSCAPE_ID VERSION_ID WIPE="false":
    protoc --icepanel_out=. \
           --icepanel_opt=landscape={{LANDSCAPE_ID}},version={{VERSION_ID}},wipe={{WIPE}} \
           {{PROTO_FILES}}

# Upload generated objects to IcePanel
upload-objects FILE="icepanel_objects.json" VERBOSE="":
    #!/usr/bin/env bash
    set -euo pipefail
    VERBOSE_FLAG=""
    if [ "{{VERBOSE}}" = "verbose" ]; then
        VERBOSE_FLAG="-v"
    fi
    ./uploader -file {{FILE}} ${VERBOSE_FLAG}

# Generate and upload in one step
proto-to-icepanel PROTO_FILES LANDSCAPE_ID VERSION_ID WIPE="false" VERBOSE="":
    #!/usr/bin/env bash
    set -euo pipefail
    just build-plugin
    just build-uploader
    just generate-objects {{PROTO_FILES}} {{LANDSCAPE_ID}} {{VERSION_ID}} {{WIPE}}
    
    VERBOSE_FLAG=""
    if [ "{{VERBOSE}}" = "verbose" ]; then
        VERBOSE_FLAG="verbose"
    fi
    just upload-objects icepanel_objects.json ${VERBOSE_FLAG}

# Run tests
test:
    go test -v ./...

# Run test coverage
coverage:
    go test -coverprofile=coverage.out ./...
    go tool cover -func=coverage.out

# Show test coverage in browser
coverage-html:
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out

# Output test coverage to file
coverage-report OUTPUT="coverage.html":
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o {{OUTPUT}}

# Install golangci-lint if not available
setup-lint:
    #!/usr/bin/env bash
    if ! command -v golangci-lint &> /dev/null; then
        echo "Installing golangci-lint..."
        # Install latest version (v2.1.6)
        curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.1.6
    else
        INSTALLED_VERSION=$(golangci-lint --version | awk '{print $4}')
        if [ "$INSTALLED_VERSION" != "2.1.6" ]; then
            echo "Updating golangci-lint to v2.1.6..."
            curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.1.6
        else
            echo "golangci-lint v2.1.6 is already installed"
        fi
    fi

# Run linter with strict settings
lint: setup-lint
    gofumpt -w .
    golangci-lint run ./...

# Show and fix linting issues where possible
lint-fix: setup-lint
    golangci-lint run --fix ./...

# Clean build artifacts
clean:
    rm -f mermaid-icepanel
    rm -f coverage.out coverage.html
    rm -f icepanel_objects.json
    rm -f uploader

# Install binary to $GOPATH/bin
install:
    go install 
