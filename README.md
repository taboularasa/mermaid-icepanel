# Mermaid IcePanel

A CLI tool that converts Mermaid C4 (System-Context subset) diagrams to IcePanel diagrams.

## Features

- Convert Mermaid C4 diagrams to IcePanel format
- Support for Person, System, System_Ext, SystemDb, and System_Boundary elements
- Support for relationships (Rel and BiRel)
- Option to wipe existing content in an IcePanel version
- Environment variable configuration
- Modular architecture with dependency injection for testability

## Project Structure

The project is organized into these packages:

```
mermaid-icepanel/
├── internal/
│   ├── api/       # IcePanel API client
│   ├── config/    # Configuration handling
│   └── parser/    # Mermaid diagram parser
├── .env.example   # Example environment variables
├── justfile       # Task runner commands
├── main.go        # Entry point
└── README.md      # This file
```

## Installation

### Prerequisites

- Go 1.16+
- [Just](https://github.com/casey/just) command runner (optional but recommended)

### Building from source

```bash
# Clone the repository
git clone https://github.com/yourusername/mermaid-icepanel.git
cd mermaid-icepanel

# Build using Go directly
go build -o mermaid-icepanel main.go

# Or build using Just
just build

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

### Using Just

The easiest way to use the tool is with the included Just commands:

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

### Direct CLI Usage

```bash
# Basic usage
./mermaid-icepanel -mmd path/to/diagram.mmd -landscape landscape-id -version version-id -name "Diagram Name"

# With all options
./mermaid-icepanel -mmd path/to/diagram.mmd -landscape landscape-id -version version-id -name "Diagram Name" -token your-token -wipe -v
```

## Command Line Arguments

| Flag | Description | Required |
|------|-------------|----------|
| `-mmd` | Path to Mermaid .mmd file | Yes |
| `-landscape` | IcePanel landscape ID | Yes |
| `-version` | IcePanel version ID | Yes |
| `-name` | Diagram name | No (defaults to "Imported diagram") |
| `-token` | API token | No (falls back to ICEPANEL_TOKEN env variable) |
| `-wipe` | Delete existing content before import | No |
| `-v` | Verbose output | No |

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

The tool supports the following Mermaid C4 syntax elements:

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
