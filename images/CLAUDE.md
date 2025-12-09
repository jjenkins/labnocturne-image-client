# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Lab Nocturne Images API is a developer-friendly image storage API built with Go and Fiber. The goal is to provide a "stupidly simple" image storage service that developers can start using in under 60 seconds - no dashboards, no configuration. The guiding principle is "curl-first": if it doesn't work beautifully in curl, it's not done.

**Key Product Goals:**
- Instant gratification: landing page to first upload in <60 seconds
- Test keys generated instantly without signup (`ln_test_*`)
- Production keys with email + payment (`ln_live_*`)
- Files stored in S3 with CloudFront CDN
- ULID-based file IDs with S3 partitioning

## Architecture

This codebase follows a **layered architecture** pattern with clear separation of concerns:

```
Handler → Service → Store → Database
```

### Layer Responsibilities

**Handler Layer** (`internal/handlers/`):
- HTTP request/response handling
- Request parameter extraction and validation
- Dependency instantiation (stores and services)
- HTTP status code management
- Returns JSON for this API-only service (no templates/HTMX)

**Service Layer** (`internal/service/`):
- Business logic implementation
- Data validation and transformation
- External API integration (S3, Stripe, Twilio if needed)
- Transaction orchestration
- Quota determination and usage calculations

**Store Layer** (`internal/store/`):
- Database access (CRUD operations)
- SQL query execution using parameterized queries
- Row scanning and mapping
- Use aggregation queries (SUM, COUNT) for efficient statistics calculation

**Model Layer** (`internal/model/`):
- Data structure definitions
- Business logic helpers (formatting, validation)
- Type conversions

### Entry Point Flow

```
main.go → cmd/root.go → cmd/serve.go (API server)
                      → cmd/worker.go (Cleanup jobs)
```

- `main.go`: Minimal entry point, calls `cmd.Execute()`
- `cmd/root.go`: Cobra root command definition
- `cmd/serve.go`: Server setup, DB connection, route registration, Fiber app initialization
- `cmd/worker.go`: Cleanup worker, runs once and exits (for cron scheduling)

### Dependency Injection Pattern

Use **constructor-based dependency injection**:

```go
// In handler
func HandleUpload(sqlDB *sql.DB) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // 1. Instantiate stores
        fileStore := store.NewFileStore(sqlDB)

        // 2. Instantiate services
        uploadService := service.NewUploadService(fileStore, s3Client)

        // 3. Call service layer
        result, err := uploadService.Upload(ctx, file)

        // 4. Return JSON response
        return c.JSON(result)
    }
}
```

## Key Technical Details

### ULID File Storage

Files use ULID (Universally Unique Lexicographically Sortable Identifier) instead of UUID:
- Time-ordered for better database index performance
- Compact: 26 characters vs 36 for UUID
- URL-safe, sortable

**S3 Partitioning:**
```
Pattern: {char1}/{char2}/{char3}/{ULID}.{ext}
Example: 0/1/a/01ARZ3NDEKTSV4RRFFQ69G5FAV.jpg
```

Use lowercase first 3 characters of ULID for partition path. This creates 36^3 = 46,656 partition buckets.

**External ID Format:**
- API returns: `img_{lowercase_ulid}`
- Database stores: uppercase ULID (canonical form)
- S3 key uses: uppercase ULID in filename

### Database Schema (PostgreSQL)

**Users table:**
- API keys with prefixes `ln_test_*` or `ln_live_*`
- Key type determines quotas and retention
- Optional email and Stripe customer ID

**Files table:**
- `id`: ULID (uppercase, canonical form)
- `external_id`: 'img_' + lowercase ULID (what API returns)
- `s3_key`: Full partition path
- `cdn_url`: CloudFront URL
- Soft delete with `deleted_at` timestamp

### API Design Philosophy

**Always return JSON, never HTML.** This is an API-only service.

**Error responses must be helpful:**
```json
{
  "error": {
    "message": "File size exceeds limit for test keys (10MB). Upgrade to increase limits.",
    "type": "file_too_large",
    "docs": "https://images.labnocturne.com/docs#limits",
    "code": "file_size_exceeded"
  }
}
```

## Development Commands

### Local Development

```bash
# Build the binary
go build -o bin/images .

# Run the server
./bin/images serve

# Run with custom port
./bin/images serve --port 3000
```

### Docker Development

```bash
# Start all services (app + PostgreSQL)
make dev
# or: docker-compose up --build

# View logs
make logs
# or: docker-compose logs -f app

# Restart services
make restart

# Rebuild and restart
make rebuild

# Stop services
make down

# Stop and remove volumes
make down-volumes

# Connect to PostgreSQL
make psql

# Clean up Docker resources
make clean
```

**Hot Reload:**
The Docker setup uses [Air](https://github.com/cosmtrek/air) for automatic hot reloading during development. When you modify Go source files, the application automatically rebuilds and restarts within the container. You don't need to manually restart services - just save your changes and wait a few seconds for the reload.

**Worker (Cleanup Jobs):**
```bash
# Run worker locally
./bin/images worker

# Dry run (shows what would be deleted without deleting)
./bin/images worker --dry-run

# Run worker in Docker
docker-compose --profile worker run --rm worker
```

The worker command performs cleanup of expired files:
- **Test file expiration**: Hard deletes test key files older than 7 days
- **Soft-deleted cleanup**: Permanently removes soft-deleted files older than 30 days
- Deletes from S3 first, then database (ensures consistency)
- Logs individual file operations
- Runs once and exits (designed for cron scheduling)

### Database

The application uses PostgreSQL. Connection via `DATABASE_URL` environment variable or individual components (`DATABASE_USER`, `DATABASE_PASSWORD`, etc.).

**Database Best Practices:**

1. **Use aggregation queries for statistics**: Instead of loading all records and calculating in Go, use SQL aggregation functions (SUM, COUNT, AVG) for efficient calculations.
   ```sql
   -- Good: Single aggregation query
   SELECT COALESCE(SUM(size_bytes), 0) as total_bytes, COUNT(*) as file_count
   FROM files WHERE user_id = $1 AND deleted_at IS NULL

   -- Bad: Loading all records
   SELECT size_bytes FROM files WHERE user_id = $1 AND deleted_at IS NULL
   ```

2. **Use COALESCE for NULL handling**: When aggregating, use `COALESCE(SUM(...), 0)` to return 0 instead of NULL when no rows match.

3. **Leverage existing indexes**: Before writing queries, check what indexes exist (e.g., `idx_files_user_id`) and structure queries to use them.

4. **Parameterized queries only**: Always use `$1, $2, ...` placeholders to prevent SQL injection. Never concatenate user input into SQL strings.

## Environment Configuration

Configuration managed via `.envrc` (using direnv):

- `PORT` - Server port (default: 8080)
- `DATABASE_URL` - Full PostgreSQL connection string
- Or individual components: `DATABASE_USER`, `DATABASE_PASSWORD`, `DATABASE_NAME`, `DATABASE_HOST`, `DATABASE_PORT`

## Project Status

**Currently Implemented:**
- Basic web server with Fiber
- Health check endpoint (`GET /health`)
- API info endpoint (`GET /`)
- Database connection with PostgreSQL 18+
- Docker setup with hot reload via Air
- API key generation (`GET /key`) - test keys
- File upload (`POST /upload`) - S3 integration with ULID partitioning
- File retrieval (`GET /i/:ulid.:ext`) - Presigned URLs with redirect
- File deletion (`DELETE /i/:id`) - Soft delete
- File listing (`GET /files`) - Pagination and sorting
- Usage statistics (`GET /stats`) - Real-time storage tracking and quotas
- Background cleanup worker (`./bin/images worker`) - Expires test files and permanently deletes soft-deleted files

**To Be Implemented (see `notes/prd-proof-of-concept.md`):**
- Live API key generation with email/payment
- Rate limiting per API key (infrastructure in place, tracking pending)
- Bandwidth tracking from CloudFront logs
- File type validation improvements (magic bytes)

## Important Architectural Notes

1. **No frontend/templates**: This is a JSON API only. Do not add templ, HTMX, or HTML rendering.

2. **Database connection**: The database connection is **required** for the server to start. The application will fail to start if `DATABASE_URL` is not set or the database is unavailable.

3. **Fiber framework**: Use `fiber.Handler` return signature for all handlers. Routes are registered in `cmd/serve.go`.

4. **Error handling**: Services return errors with context (`fmt.Errorf("failed to X: %w", err)`). Handlers convert these to appropriate HTTP status codes and JSON responses.

5. **Authentication pattern**: For endpoints requiring authentication, follow this pattern:
   ```go
   // 1. Extract and validate Authorization header
   authHeader := c.Get("Authorization")
   if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
       return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
           "error": fiber.Map{
               "message": "Invalid API key",
               "type":    "unauthorized",
               "code":    "invalid_api_key",
           },
       })
   }

   // 2. Extract API key
   apiKey := strings.TrimPrefix(authHeader, "Bearer ")

   // 3. Authenticate user
   userStore := store.NewUserStore(db)
   user, err := userStore.FindByAPIKey(c.Context(), apiKey)
   if err != nil {
       return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
           "error": fiber.Map{
               "message": "Invalid API key",
               "type":    "unauthorized",
               "code":    "invalid_api_key",
           },
       })
   }
   ```

6. **Error response format**: All API errors must follow this structure:
   ```json
   {
     "error": {
       "message": "Human-readable error message",
       "type": "error_category",
       "code": "machine_readable_error_code"
     }
   }
   ```

7. **Docker volume mount**: PostgreSQL 18+ requires volume at `/var/lib/postgresql` (not `/var/lib/postgresql/data`). The docker-compose.yml uses a named volume `pgdata`.

## Documentation Reference

- Full PRD: `notes/prd-proof-of-concept.md`
- Architecture patterns: `notes/architecture.md`
- README: Standard build/run instructions
