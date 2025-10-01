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