# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Development Commands

```bash
# Run the API server (generates Swagger docs and starts server)
make run-api

# Build the CLI
go build -o jorge-cli main.go

# Build CLI from cmd/cli directory
cd cmd/cli && go build -o cli main.go

# Run tests
go test ./...
```

## Project Architecture

Jorge CLI is a dual-component Go project: a REST API server and a CLI client for SRE/infrastructure tasks.

### Two Entry Points

- **`cmd/api/main.go`** - REST API server using Gin framework with SQLite storage
- **`cmd/cli/main.go`** - CLI client using Cobra, communicates with the API

### API Server Structure (`internal/`)

Follows clean architecture with dependency injection:

```
Repository (SQLite) → UseCase (business logic) → Handler (HTTP) → Router (Gin)
```

- `internal/repository/` - Data access layer with `ResourceRepository` interface
- `internal/usecase/` - Individual use case structs (CreateResource, DeleteResource, etc.)
- `internal/handler/` - HTTP handlers, one file per operation
- `internal/router/` - Gin router initialization
- `internal/schemas/` - GORM models (Resource with Name, Dns fields)

### CLI Structure (`cmd/cli/`)

- `commands/` - Cobra command definitions (list, create, update, delete, debug)
- `internal/usecase/` - CLI-specific usecases (resource operations via HTTP, network diagnostics)
- `internal/config/` - Viper-based config loading from environment variables
- `internal/utils/` - HTTP client, logging, signal handling, display helpers
- `internal/models/` - CLI-specific data models

### Key Libraries

- **Cobra** for CLI commands
- **Gin** for HTTP API
- **GORM** with SQLite for persistence
- **Viper** for configuration
- **Zap** for structured logging
- **Lipgloss** for CLI output styling
- **Swaggo** for API documentation

### Configuration

CLI reads from environment variables:
- `API_BASE_URL` (default: `http://localhost:8080/api/v1`)
- `TIMEOUT` (default: 10 seconds)
- `VERSION` (default: `v1.0.0`)
