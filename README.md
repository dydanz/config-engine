# config-engine
Configuration management service built in Go that provides schema-based validation, versioning, and rollback support for configuration data.

## Prerequisites

### For Docker Deployment (Recommended)
- **Docker** ([Download Docker Desktop](https://www.docker.com/products/docker-desktop))
- **Docker Compose** (included with Docker Desktop)

### For Local Development
- **Go 1.21 or higher** ([Download](https://golang.org/dl/))
- **Make** (optional, for using Makefile commands)
- **Git** (for cloning the repository)

### Dependency Installation

The service uses the following external dependencies:
- `github.com/gin-gonic/gin` - HTTP routing
- `github.com/xeipuuv/gojsonschema` - JSON Schema validation

Dependencies are managed via Go modules and will be automatically downloaded.

## Installation

### Step 1: Navigate to Project Directory

```bash
cd /Users/dandi/Sandbox/config-engine
```

### Step 2: Install Dependencies

```bash
make deps
```

Or manually:

```bash
go mod download
go mod tidy
```

### Step 3: Build the Application

```bash
make build
```

This will create a binary at `bin/config-engine`.

## Quick Start

### Option 1: Using Docker Compose (Recommended for Easy Setup)

**Prerequisites**: Docker and Docker Compose installed

```bash
# Build and start the service
docker-compose up -d

# View logs
docker-compose logs -f

# Stop the service
docker-compose down
```

The service will be available at `http://localhost:8080`

### Option 2: Using Make

```bash
# Build and run the service
make run
```

### Option 3: Using Go Directly

```bash
# Run without building binary (development mode)
go run main.go

# Or build and run
go build -o bin/config-engine main.go
./bin/config-engine
```

### Option 4: Development Mode

```bash
make dev
```

The service will start on `http://localhost:8080` by default (non-Docker).

You can specify a custom port:

```bash
./bin/config-engine -port=9000
```

### Verify Installation

Test the health endpoint:

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{"status":"healthy"}
```

## Design Decisions

### 1. In-Memory Storage

**Decision**: Use in-memory storage with mutex-based concurrency control.

**Rationale**:
- Fast read/write operations
- Simple deployment (no external dependencies)
- Thread-safe with `sync.RWMutex`, to avoid racing conditions on read/update/delete events
- Suitable for the rapid prototyping/small-scale research project

**Trade-offs**:
- Data is not persistent, wiped out once the apps are restarted/shutdown
- Limited by available memory, can overflow'ed your pc memory

### 2. Immutable Version History

**Decision**: Store complete configuration data for each version.

**Rationale**:
- Simple yet straightforward rollback implementation
- Complete audit trail, we can record who/when/what when the version changed.
- No need to replay changes
- Fast version retrieval

**Trade-offs**:
- Higher memory usage for large configs

### 3. Layered Architecture

**Decision**: Follow common pattern, separate concerns into distinct layers (handlers → service → repository).

**Rationale**:
- Clear separation of main functions, ie similar to domain segregation.

### 4. JSON Schema Validation

**Decision**: Use `gojsonschema` library for validation.

**Rationale**:
- Industry standard (JSON Schema specification)
- Flexible and extensible
- Rich validation capabilities
- Adequate error message handling

### 5. Graceful Shutdown

**Decision**: Implement graceful shutdown with timeout.

**Rationale**:
- Production-ready behaviour ensures clean resource cleanup, completion of in-flight requests, and prevention of data corruption.

### 6. Thread Safety

**Decision**: Use `sync.RWMutex` for concurrent access control.

**Rationale**:
- Production-ready for multi-user environments
- Read-optimized (multiple concurrent reads)
- Write-safe (exclusive write locks)
- No race conditions


## Project Structure

```
config-engine/
├── api.yml                 # OpenAPI v3 specification
├── main.go                 # Application entry point
├── go.mod                  # Go module definition
├── Makefile                # Build and test automation
├── README.md               # This file
├── internal/               # Internal packages
│   ├── models/             # Domain models and DTOs
│   │   └── config.go
│   ├── repository/         # Data storage layer
│   │   ├── repository.go
│   │   └── repository_test.go
│   ├── service/            # Business logic layer
│   │   ├── service.go
│   │   └── service_test.go
│   ├── validation/         # Schema validation
│   │   ├── validator.go
│   │   └── validator_test.go
│   └── handlers/           # HTTP handlers
│       └── handlers.go
└── tests/                  # Integration tests
    └── integration_test.go
```

### Package Descriptions

- **`internal/models`**: Core domain entities, request/response structures, and custom errors
- **`internal/repository`**: Thread-safe in-memory storage with versioning support
- **`internal/service`**: Business logic, validation orchestration, and use case implementations
- **`internal/validation`**: JSON Schema validation with extensible schema registry
- **`internal/handlers`**: HTTP request/response handling, routing, and middleware
- **`tests`**: End-to-end integration tests
