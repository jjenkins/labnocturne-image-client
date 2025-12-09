# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Flex is a Go web application framework built on Fiber (web framework), Templ (type-safe templates), and HTMX (dynamic HTML). It follows a layered architecture with server-side rendering and minimal JavaScript.

**Key Technologies:**
- **Fiber v2**: Fast HTTP web framework
- **Templ**: Type-safe Go templating for HTML generation
- **HTMX**: Declarative AJAX/dynamic updates without complex JavaScript
- **Tailwind CSS**: Utility-first CSS framework
- **Cobra**: CLI framework for commands

## Development Commands

### Running the Application

```bash
# Build the application
go build -o flex

# Run the server (port 8080 by default)
./flex serve

# Run on custom port
./flex serve --port 3000
# or
PORT=3000 ./flex serve
```

### Template Development

```bash
# Generate Go code from .templ files (required after editing templates)
templ generate

# The templates live in internal/templates/
# After editing .templ files, always run templ generate before building
```

### Dependency Management

```bash
# Install/update dependencies
go mod tidy

# Add a new dependency
go get github.com/package/name
```

### Docker Development (if database needed)

```bash
# Start development environment with hot reload
make dev

# View logs
make logs

# Stop services
make down

# Connect to postgres console
make psql
```

## Architecture

This application follows a **strict layered architecture** as documented in `architecture.md`. Read that file for comprehensive details. Key points:

### Request Flow

```
HTTP Request → Handler → Service → Store → Database
                 ↓         ↓         ↓
              Template  Business  Data Access
               Logic     Logic      Layer
```

### Layer Responsibilities

**Handler Layer** (`internal/handlers/`):
- HTTP request/response handling
- Request validation and parameter extraction
- Dependency instantiation (stores, services)
- Template rendering via Templ
- **Pattern**: Handlers instantiate their dependencies at function scope

**Service Layer** (`service/` - if present):
- Business logic implementation
- Data validation and transformation
- External API integration
- Transaction orchestration

**Store Layer** (`internal/store/` - if present):
- Database CRUD operations
- SQL query execution
- Row scanning and mapping

**Model Layer** (`internal/model/` - if present):
- Data structure definitions
- Helper methods for formatting/validation
- Type conversions

### Critical Patterns

**Dependency Injection**:
```go
// Handlers instantiate their own dependencies
func HomeHandler() fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Instantiate stores/services here
        // Call business logic
        // Render template
    }
}
```

**Template Rendering**:
```go
// Standard pattern for all Templ templates
page := templates.Home()
handler := adaptor.HTTPHandler(templ.Handler(page))
return handler(c)
```

**Database Type Handling**:
- Some stores require `*sql.DB` (standard library)
- Others require custom DB wrapper
- Check store constructor signature and wrap if needed

## Frontend Architecture: HTMX + Templ

This is a **server-side rendered application**. The server returns HTML, not JSON.

### HTMX Principles

- HTMX attributes on HTML elements trigger AJAX requests
- Server returns HTML fragments, not JSON
- HTMX swaps fragments into the DOM
- Minimal JavaScript required

**Common HTMX attributes:**
```html
<button hx-post="/api/action" hx-target="#result" hx-swap="innerHTML">
    Click Me
</button>
```

### Template Structure

Templates are in `internal/templates/`:
- `layouts/base.templ`: Main page layout
- `home.templ`: Home page content
- Templates use `.templ` extension
- **Always run `templ generate` after editing templates**

### JavaScript Guidelines

**DO use JavaScript for:**
- Client-side UI interactions (dropdowns, modals)
- Form validation feedback
- Date/time pickers

**DON'T use JavaScript for:**
- Data fetching (use HTMX)
- Form submission (use HTMX)
- Page navigation (use HTMX)

## Project Structure

```
flex/
├── cmd/                    # CLI commands (Cobra)
│   ├── root.go            # Root command setup
│   └── serve.go           # Server startup and routing
├── internal/
│   ├── handlers/          # HTTP handlers
│   │   └── home.go
│   ├── templates/         # Templ templates
│   │   ├── layouts/
│   │   │   └── base.templ
│   │   └── home.templ
│   ├── store/             # Data access layer (if present)
│   └── model/             # Data models (if present)
├── main.go                # Application entry point
├── go.mod                 # Go dependencies
└── architecture.md        # Detailed architecture docs
```

## Adding a New Page

1. **Create template** in `internal/templates/`:
```go
// internal/templates/about.templ
package templates

import "github.com/jjenkins/labnocturne/flex/internal/templates/layouts"

templ About() {
    @layouts.Base("About") {
        <div class="max-w-4xl mx-auto">
            <h1 class="text-4xl font-bold">About</h1>
        </div>
    }
}
```

2. **Generate template code**:
```bash
templ generate
```

3. **Create handler** in `internal/handlers/`:
```go
// internal/handlers/about.go
package handlers

import (
    "github.com/a-h/templ"
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/adaptor"
    "github.com/jjenkins/labnocturne/flex/internal/templates"
)

func AboutHandler() fiber.Handler {
    return func(c *fiber.Ctx) error {
        page := templates.About()
        handler := adaptor.HTTPHandler(templ.Handler(page))
        return handler(c)
    }
}
```

4. **Register route** in `cmd/serve.go`:
```go
app.Get("/about", handlers.AboutHandler())
```

5. **Build and run**:
```bash
go build -o flex
./flex serve
```

## Important Files

- **`architecture.md`**: Comprehensive architecture documentation - read this for detailed patterns
- **`cmd/serve.go`**: Application setup, route registration, middleware configuration
- **`go.mod`**: Dependencies and module definition
- **`internal/templates/layouts/base.templ`**: Main layout template with sidebar, styling

## Common Tasks

### Adding a Route
Routes are registered in `cmd/serve.go` in the `serveCmd` Run function.

### Modifying Layout
Edit `internal/templates/layouts/base.templ`, then run `templ generate`.

### Changing Styles
The app uses Tailwind CSS via CDN. Utility classes are applied directly in templates.
Custom styles are in the `<style>` block of `base.templ`.

### Adding HTMX Interactivity
Add HTMX attributes to templates:
```html
<button hx-get="/data" hx-target="#result">Load</button>
<div id="result"></div>
```

Create handler that returns HTML fragment:
```go
func DataHandler() fiber.Handler {
    return func(c *fiber.Ctx) error {
        fragment := templates.DataFragment()
        return adaptor.HTTPHandler(templ.Handler(fragment))(c)
    }
}
```

## Module and Import Paths

- Module: `github.com/jjenkins/labnocturne/flex`
- All internal imports use this module path
- Example: `import "github.com/jjenkins/labnocturne/flex/internal/handlers"`

## Key Dependencies

From `go.mod`:
- `github.com/a-h/templ` - Template engine
- `github.com/gofiber/fiber/v2` - Web framework
- `github.com/spf13/cobra` - CLI framework

## Environment Variables

- `PORT`: Server port (default: 8080)
- `DATABASE_URL`: PostgreSQL connection string (if using database)

## Additional Resources

For detailed architectural patterns, dependency injection, HTMX usage, error handling, and complete examples, see `architecture.md`.
