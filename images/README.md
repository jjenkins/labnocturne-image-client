# Lab Nocturne Images API

A simple, developer-friendly image storage API.

## Quick Start

### Build and Run

```bash
# Build the application
go build -o bin/images .

# Run the server
./bin/images serve

# Or specify a custom port
./bin/images serve --port 3000
```

The server will start on port 8080 by default.

### Environment Variables

Configuration is managed via `.envrc` (using direnv):

- `PORT` - Server port
- `DATABASE_USER` - PostgreSQL username
- `DATABASE_PASSWORD` - PostgreSQL password
- `DATABASE_NAME` - Database name
- `DATABASE_HOST` - Database host
- `DATABASE_PORT` - Database port

Or use `DATABASE_URL` directly:
- `DATABASE_URL` - Full PostgreSQL connection string

### Available Commands

**Server:**
```bash
./bin/images serve [--port 8080]
```

**Worker (Cleanup):**
```bash
# Run cleanup jobs (deletes expired files)
./bin/images worker

# Dry run (show what would be deleted without deleting)
./bin/images worker --dry-run
```

The worker command:
- Deletes test key files older than 7 days (hard delete)
- Permanently removes soft-deleted files older than 30 days
- Runs once and exits (suitable for cron scheduling)
- Logs individual file deletions

### Available Endpoints

- `GET /` - API information
- `GET /health` - Health check

### Run with Docker

```bash
# Start all services (app + database)
docker-compose up -d

# Run worker once (cleanup jobs)
docker-compose --profile worker run --rm worker

# View logs
docker-compose logs -f app

# Stop services
docker-compose down

# Stop and remove volumes
docker-compose down -v
```

**Note:** The worker service uses a Docker Compose profile and only runs when explicitly invoked. This allows you to run cleanup jobs manually or via cron without keeping a long-running worker container.

### Test the API

```bash
# Get API info
curl http://localhost:8080/

# Health check
curl http://localhost:8080/health
```

## Project Structure

```
.
├── cmd/                    # CLI commands
│   ├── root.go            # Root command
│   ├── serve.go           # Server command
│   └── worker.go          # Cleanup worker command
├── internal/              # Internal packages
│   ├── handlers/          # HTTP handlers
│   ├── service/           # Business logic layer
│   │   ├── cleanup.go     # File cleanup service
│   │   └── file.go        # File operations service
│   └── store/             # Database layer
│       ├── file.go        # File store methods
│       └── file_test_helpers.go  # Test helpers for aging files
├── notes/                 # Documentation
│   ├── architecture.md    # Architecture patterns
│   └── prd-proof-of-concept.md  # Product requirements
├── main.go               # Application entry point
└── go.mod                # Go module definition
```

## Deployment

### Scheduling Cleanup Jobs

The worker command is designed to be invoked by cron or a similar scheduler:

**Example cron entry (run daily at 2 AM):**
```bash
0 2 * * * cd /path/to/images && ./bin/images worker >> /var/log/images-cleanup.log 2>&1
```

**Render Cron Jobs:**
On Render, create a Cron Job service:
- Command: `./bin/images worker`
- Schedule: `0 2 * * *` (daily at 2 AM UTC)
- Environment: Same as your main web service

The worker will:
1. Connect to the database and S3
2. Delete expired test files (>7 days old)
3. Permanently remove soft-deleted files (>30 days old)
4. Log results and exit

## Next Steps

Refer to the PRD in `notes/prd-proof-of-concept.md` for the full feature roadmap.
