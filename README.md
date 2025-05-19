# Mermaid IcePanel

A toolkit for converting service definitions and architecture diagrams to IcePanel.

## Features

- Convert Mermaid C4 diagrams to IcePanel format
- Extract service definitions from Protocol Buffer files
- Support for Person, System, System_Ext, SystemDb, and System_Boundary elements
- Support for relationships (Rel and BiRel)
- Option to wipe existing content in an IcePanel version before importing
- Environment variable configuration
- Modular architecture with dependency injection for testability

## Project Structure

The project is organized into these packages:

```
mermaid-icepanel/
├── cmd/
│   └── protoc-gen-icepanel/  # Protocol Buffer plugin
│       ├── internal/         # Plugin internals
│       └── upload/           # Object uploader tool
├── internal/
│   ├── api/                  # IcePanel API client
│   ├── config/               # Configuration handling
│   └── parser/               # Mermaid diagram parser
├── .env.example              # Example environment variables
├── justfile                  # Task runner commands
├── main.go                   # CLI entry point for Mermaid tool
└── README.md                 # This file
```

## Installation

### Prerequisites

- Go 1.16+
- [Just](https://github.com/casey/just) command runner (optional but recommended)
- Protocol Buffer compiler (protoc) for the protoc-gen-icepanel plugin

### Building from source

```bash
# Clone the repository
git clone https://github.com/yourusername/mermaid-icepanel.git
cd mermaid-icepanel

# Build the Mermaid-to-IcePanel tool
just build

# Build the protoc-gen-icepanel plugin
just build-plugin

# Build the uploader tool
just build-uploader

# Optional: Install to $GOPATH/bin
just install
```

## Configuration

The application can be configured using environment variables or command-line flags.

### Environment Variables

Copy the example environment file and edit with your settings:

```bash
cp .env.example .env
```

Available environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `ICEPANEL_API_URL` | Base URL for IcePanel API | https://api.icepanel.io/v1 |
| `ICEPANEL_TIMEOUT_SECONDS` | Request timeout in seconds | 30 |
| `ICEPANEL_TOKEN` | Your IcePanel API token | - |

## Usage

### Mermaid-to-IcePanel Tool

#### Using Just

The easiest way to use the Mermaid tool is with the included Just commands:

```bash
# Display all available commands
just

# Import a Mermaid diagram to IcePanel
just import path/to/diagram.mmd landscape-id version-id "Diagram Name"

# Import and wipe existing content
just import path/to/diagram.mmd landscape-id version-id "Diagram Name" wipe

# Run with custom arguments
just sync -mmd path/to/diagram.mmd -landscape landscape-id -version version-id -name "My Diagram" -wipe -v
```

#### Direct CLI Usage

```bash
# Basic usage
./mermaid-icepanel -mmd path/to/diagram.mmd -landscape landscape-id -version version-id -name "Diagram Name"

# With all options
./mermaid-icepanel -mmd path/to/diagram.mmd -landscape landscape-id -version version-id -name "Diagram Name" -token your-token -wipe -v
```

#### Command Line Arguments

| Flag | Description | Required |
|------|-------------|----------|
| `-mmd` | Path to Mermaid .mmd file | Yes |
| `-landscape` | IcePanel landscape ID | Yes |
| `-version` | IcePanel version ID | Yes |
| `-name` | Diagram name | No (defaults to "Imported diagram") |
| `-token` | API token | No (falls back to ICEPANEL_TOKEN env variable) |
| `-wipe` | Delete existing content before import | No |
| `-v` | Verbose output | No |

### Proto-to-IcePanel Tool

The Proto-to-IcePanel tool consists of a protoc plugin and an uploader tool. It can extract service definitions from Proto files and upload them to IcePanel.

#### Using Just

```bash
# Generate IcePanel objects from Proto files
just generate-objects "path/to/*.proto" landscape-id version-id false

# Generate with wipe option (will wipe existing content before import)
just generate-objects "path/to/*.proto" landscape-id version-id true

# Upload generated objects to IcePanel
just upload-objects icepanel_objects.json verbose

# Generate and upload in one step (with wipe option)
just proto-to-icepanel "path/to/*.proto" landscape-id version-id true verbose
```

#### Direct Usage

```bash
# Generate IcePanel objects from Proto files
protoc --icepanel_out=. \
       --icepanel_opt=landscape=landscape-id,version=version-id,wipe=true \
       path/to/*.proto

# Upload generated objects to IcePanel
./uploader -file icepanel_objects.json -v
```

#### Command Line Arguments for Uploader

| Flag | Description | Required |
|------|-------------|----------|
| `-file` | Path to the generated objects file | No (defaults to "icepanel_objects.json") |
| `-token` | API token | No (falls back to ICEPANEL_TOKEN env variable) |
| `-landscape` | Override landscape ID from the file | No |
| `-version` | Override version ID from the file | No |
| `-dry-run` | Don't actually upload to IcePanel | No |
| `-v` | Verbose output | No |
| `-timeout` | Request timeout in seconds | No (defaults to 30) |

## Development

### Testing

The project uses Go's standard testing package with dependency injection to enable effective unit testing without external dependencies.

```bash
# Run all tests
just test

# Run tests with coverage summary
just coverage

# View coverage report in browser
just coverage-html

# Generate coverage report file
just coverage-report [output-file]
```

### Linting

The project uses [golangci-lint](https://golangci-lint.run/) with strict settings to enforce code quality and consistency.

```bash
# Run the linter
just lint

# Run the linter and automatically fix issues where possible
just lint-fix
```

The linting configuration in `.golangci.yml` enables over 30 linters with strict settings. The `just setup-lint` command will automatically install golangci-lint if it's not already available.

### Architecture

The application uses dependency injection to improve testability:

- **FileReader interface**: Abstracts file system operations
- **HTTPClient interface**: Abstracts HTTP requests
- **Config struct**: Centralizes configuration and environment variables

This design allows for easy mocking of external dependencies during testing.

## Mermaid C4 Syntax Support

The Mermaid tool supports the following Mermaid C4 syntax elements:

```
Person(id, "Label", "Optional Description")
System(id, "Label", "Optional Description")
System_Ext(id, "Label", "Optional Description")
SystemDb(id, "Label", "Optional Description")
System_Boundary(id, "Label") { ... }
Rel(from, to, "Label")
BiRel(from, to, "Label")
```

## Example

```mermaid
System_Boundary(b1, "Boundary") {
  Person(user, "User", "A user of the system")
  System(app, "Application", "Core application")
  SystemDb(db, "Database", "Stores user data")
}

System_Ext(api, "External API", "Third-party service")

Rel(user, app, "Uses")
Rel(app, db, "Reads/Writes")
BiRel(app, api, "Syncs data")
```

## License

[MIT License](LICENSE) 
