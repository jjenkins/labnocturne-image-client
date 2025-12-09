# Architecture Documentation

This document describes the architectural patterns and conventions used in this Go web application built with the Fiber framework.

## Table of Contents
- [Overview](#overview)
- [Application Entry Point](#application-entry-point)
- [Layered Architecture](#layered-architecture)
- [Frontend Architecture](#frontend-architecture)
- [Handler Layer](#handler-layer)
- [Service Layer](#service-layer)
- [Store Layer](#store-layer)
- [Model Layer](#model-layer)
- [Middleware Layer](#middleware-layer)
- [Dependency Injection](#dependency-injection)
- [Error Handling](#error-handling)
- [Database Patterns](#database-patterns)

## Overview

The application follows a **layered architecture** pattern with clear separation of concerns:

```
Handler → Service → Store → Database
   ↓         ↓        ↓
Template  Business  Data Access
 Logic     Logic     Layer
```

Each layer has specific responsibilities and communicates only with adjacent layers.

## Frontend Architecture

This application uses a **server-side rendered HTMX architecture** with minimal JavaScript.

### Technology Stack

- **Templ**: Type-safe Go templating engine for server-side HTML generation
- **HTMX**: Declarative HTML attributes for AJAX, WebSockets, and Server-Sent Events
- **Tailwind CSS**: Utility-first CSS framework for styling
- **Alpine.js** (minimal): Used sparingly for client-side interactions that don't warrant a server round-trip

### HTMX Patterns

**Core Principle**: The server returns HTML fragments, not JSON. HTMX swaps these fragments into the DOM.

**Common HTMX Attributes**:
```html
<!-- Make any element trigger AJAX requests -->
<button hx-post="/api/action" hx-target="#result">
    Click Me
</button>

<!-- Swap responses into specific elements -->
<div id="result" hx-get="/data" hx-trigger="load">
    Loading...
</div>

<!-- Common patterns -->
hx-get="/path"          <!-- GET request -->
hx-post="/path"         <!-- POST request -->
hx-delete="/path"       <!-- DELETE request -->
hx-target="#element"    <!-- Where to put response -->
hx-swap="innerHTML"     <!-- How to swap (innerHTML, outerHTML, etc.) -->
hx-trigger="click"      <!-- What triggers request (click, load, etc.) -->
```

**Handler Response Patterns for HTMX**:

**Full Page Response** (initial page load):
```go
func HandlePage(db *sql.DB) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Return full page with layout
        page := component.FullPageLayout(data)
        return adaptor.HTTPHandler(templ.Handler(page))(c)
    }
}
```

**Partial HTML Response** (HTMX requests):
```go
func HandlePartialUpdate(db *sql.DB) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Return only the fragment that changed
        fragment := component.TableRowFragment(data)
        return adaptor.HTTPHandler(templ.Handler(fragment))(c)
    }
}
```

**HTMX Redirects**:
```go
func HandleAction(db *sql.DB) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Process action...

        // Redirect for HTMX clients
        c.Set("HX-Redirect", "/destination")
        return c.Redirect("/destination")
    }
}
```

**HTMX Headers for Control**:
```go
// Trigger client-side event after response
c.Set("HX-Trigger", "dataUpdated")

// Redirect the entire page
c.Set("HX-Redirect", "/new-page")

// Refresh the page
c.Set("HX-Refresh", "true")

// Replace URL in browser (for back button support)
c.Set("HX-Push-Url", "/new-url")
```

### Server-Side Rendering Benefits

1. **Simplicity**: No complex frontend build process or state management
2. **Performance**: Minimal JavaScript payload, fast initial page load
3. **SEO**: All content is server-rendered and indexable
4. **Progressive Enhancement**: Works without JavaScript enabled
5. **Type Safety**: Templ provides compile-time template checking
6. **Reduced Complexity**: Single language (Go) for business logic and rendering

### JavaScript Usage Guidelines

**DO use JavaScript for**:
- Form validation feedback (alongside server validation)
- Client-side UI interactions (dropdowns, modals)
- Phone number formatting as user types
- Date/time pickers

**DON'T use JavaScript for**:
- Data fetching (use HTMX)
- Form submission (use HTMX)
- Page navigation (use HTMX)
- State management (server handles state)

### Template Component Patterns

**Layout Components**:
```go
// Full page layout
templ Layout(title string) {
    <!DOCTYPE html>
    <html>
        <head>
            <title>{title}</title>
            <script src="https://unpkg.com/htmx.org@1.9.10"></script>
            <link href="/static/output.css" rel="stylesheet">
        </head>
        <body>
            { children... }
        </body>
    </html>
}
```

**Reusable Components**:
```go
// Button component with HTMX
templ Button(text string, endpoint string, target string) {
    <button
        hx-post={endpoint}
        hx-target={target}
        class="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded">
        {text}
    </button>
}
```

**Form Components with HTMX**:
```go
// Form that submits via HTMX and swaps response
templ ResourceForm(resource *model.Resource) {
    <form hx-post="/resources" hx-target="#result" hx-swap="innerHTML">
        <input type="text" name="name" value={resource.Name} />
        <button type="submit">Save</button>
    </form>
    <div id="result"></div>
}
```

## Application Entry Point

**Entry Flow**: `main.go` → `cmd/root.go` → `cmd/serve.go`

### main.go
- Single responsibility: Call `cmd.Execute()`
- Minimal logic, delegates to Cobra CLI framework

### cmd/root.go
- Defines the root Cobra command
- Sets up CLI structure

### cmd/serve.go
- Main application setup and configuration
- Database connection initialization
- Route registration
- Middleware configuration
- Server startup

**Key Responsibilities**:
```go
// Database initialization
db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))

// Fiber app setup
app := fiber.New()
app.Use(logger.New())

// Route registration
app.Get("/resources", handler.HandleResourceList(db))
app.Post("/resources/:id/action", handler.HandleResourceAction(db))

// Middleware groups
admin := app.Group("/admin")
admin.Use(basicauth.New(basicauth.Config{...}))
```

## Layered Architecture

### Handler Layer

**Location**: `handler/`

**Responsibilities**:
- HTTP request/response handling
- Request parameter extraction and validation
- Dependency instantiation (stores and services)
- Template rendering
- HTTP status code management

**Pattern**:
```go
func HandleResourceAction(sqlDB *sql.DB) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // 1. Wrap database if needed
        dbWrapper := &db.DB{DB: sqlDB}

        // 2. Instantiate stores
        resourceStore := store.NewResourceStore(dbWrapper)

        // 3. Instantiate services
        resourceService := service.NewResourceServiceImpl(resourceStore)

        // 4. Extract request parameters
        id := c.Params("id")
        formValue := c.FormValue("field")

        // 5. Call service layer
        ctx := context.Background()
        result, err := resourceService.PerformOperation(ctx, id, formValue)
        if err != nil {
            return c.Status(fiber.StatusBadRequest).SendString(err.Error())
        }

        // 6. Render response (template or redirect)
        page := component.ResourcePage(result)
        handler := adaptor.HTTPHandler(templ.Handler(page))
        return handler(c)
    }
}
```

**Database Type Handling**:
- Some stores require `*sql.DB` (standard library)
- Others require `*db.DB` (custom wrapper)
- Always check store constructor signature
- Wrap when needed: `dbWrapper := &db.DB{DB: sqlDB}`

**Form Handling**:
- Extract values using `c.FormValue("field_name")`
- Validate required fields before processing
- Centralize complex parsing in helper functions
- Return descriptive error messages

**Template Rendering**:
```go
// Standard pattern for Templ templates
page := component.TemplateName(data)
handler := adaptor.HTTPHandler(templ.Handler(page))
return handler(c)

// For redirects
return c.Redirect("/path")
```

### Service Layer

**Location**: `service/`

**Responsibilities**:
- Business logic implementation
- Data validation and transformation
- Cross-cutting concerns (email, SMS, payments)
- External API integration (Stripe, Twilio, Resend)
- Transaction orchestration

**Pattern**:
```go
// Service interface (optional but recommended)
type ResourceService interface {
    CreateResource(ctx context.Context, resource *model.Resource) error
    GetResource(ctx context.Context, id string) (*model.Resource, error)
}

// Service implementation
type ResourceServiceImpl struct {
    resourceStore *store.ResourceStore
    // other dependencies
}

// Constructor with dependency injection
func NewResourceServiceImpl(resourceStore *store.ResourceStore) *ResourceServiceImpl {
    return &ResourceServiceImpl{
        resourceStore: resourceStore,
    }
}

// Business logic methods
func (s *ResourceServiceImpl) CreateResource(ctx context.Context, resource *model.Resource) error {
    // 1. Validate business rules
    if resource.Field < 0 {
        return fmt.Errorf("field cannot be negative")
    }

    // 2. Perform business logic
    // 3. Call store layer
    return s.resourceStore.Create(ctx, resource)
}
```

**External Integration Patterns**:

**Stripe Integration**:
```go
func (s *ServiceImpl) CreatePaymentLink(...) (string, error) {
    // Set API key at method start
    stripe.Key = os.Getenv("STRIPE_API_KEY")

    // Create dynamic prices (don't hardcode price IDs)
    priceParams := &stripe.PriceParams{
        Currency:   stripe.String("usd"),
        UnitAmount: stripe.Int64(amountInCents),
        ProductData: &stripe.PriceProductDataParams{
            Name: stripe.String("Product Name"),
        },
    }
    price, err := price.New(priceParams)

    // Create payment link
    params := &stripe.PaymentLinkParams{
        LineItems: []*stripe.PaymentLinkLineItemParams{{
            Price:    stripe.String(price.ID),
            Quantity: stripe.Int64(1),
        }},
        Metadata: map[string]string{
            "internal_id": resourceID,
            "type": "resource_type",
        },
    }
    paymentLink, err := paymentlink.New(params)
    return paymentLink.URL, nil
}
```

**SMS Integration**:
```go
func (s *ServiceImpl) sendSMS(to, message string) error {
    from := "+1234567890" // Your Twilio number

    // Check environment for dev mode
    environment := strings.ToLower(os.Getenv("ENVIRONMENT"))
    if environment == "development" {
        log.Printf("SMS [DEV] To: %s, Message: %s", to, message)
        return nil
    }

    // Production SMS via Twilio
    accountSid := os.Getenv("TWILIO_ACCOUNT_SID")
    authToken := os.Getenv("TWILIO_AUTH_TOKEN")

    if accountSid == "" || authToken == "" {
        return fmt.Errorf("Twilio credentials not configured")
    }

    twilio := gotwilio.NewTwilioClient(accountSid, authToken)
    _, _, err := twilio.SendSMS(from, to, message, "", "")
    return err
}
```

**Error Handling**:
- Return descriptive errors: `fmt.Errorf("failed to X: %w", err)`
- Validate inputs before database operations
- Handle edge cases explicitly
- Log important events and errors

### Store Layer

**Location**: `store/`

**Responsibilities**:
- Database access (CRUD operations)
- SQL query execution
- Row scanning and mapping
- Database transaction management (when needed)

**Pattern**:
```go
type ResourceStore struct {
    db *db.DB
}

func NewResourceStore(database *db.DB) *ResourceStore {
    return &ResourceStore{db: database}
}

// Create operation
func (s *ResourceStore) Create(ctx context.Context, resource *model.Resource) error {
    stmt := `
        INSERT INTO resources (field1, field2, field3)
        VALUES ($1, $2, $3)
        RETURNING id, created_at, updated_at
    `

    err := s.db.QueryRow(ctx, stmt,
        resource.Field1,
        resource.Field2,
        resource.Field3,
    ).Scan(
        &resource.ID,
        &resource.CreatedAt,
        &resource.UpdatedAt,
    )

    return err
}

// Read operation
func (s *ResourceStore) FindByID(ctx context.Context, id string) (*model.Resource, error) {
    stmt := `
        SELECT id, field1, field2, created_at, updated_at
        FROM resources
        WHERE id = $1
    `

    resource := &model.Resource{}
    err := s.db.QueryRow(ctx, stmt, id).Scan(
        &resource.ID,
        &resource.Field1,
        &resource.Field2,
        &resource.CreatedAt,
        &resource.UpdatedAt,
    )

    if err != nil {
        return nil, err
    }

    return resource, nil
}

// List operation
func (s *ResourceStore) FindAll(ctx context.Context) ([]*model.Resource, error) {
    stmt := `
        SELECT id, field1, field2, created_at, updated_at
        FROM resources
        ORDER BY created_at DESC
    `

    rows, err := s.db.Query(ctx, stmt)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var resources []*model.Resource
    for rows.Next() {
        resource := &model.Resource{}
        err := rows.Scan(
            &resource.ID,
            &resource.Field1,
            &resource.Field2,
            &resource.CreatedAt,
            &resource.UpdatedAt,
        )
        if err != nil {
            return nil, err
        }
        resources = append(resources, resource)
    }

    // Check for errors from iterating
    if err := rows.Err(); err != nil {
        return nil, err
    }

    return resources, nil
}

// Update operation
func (s *ResourceStore) Update(ctx context.Context, resource *model.Resource) error {
    stmt := `
        UPDATE resources
        SET field1 = $2, field2 = $3, updated_at = NOW()
        WHERE id = $1
    `

    result, err := s.db.Exec(ctx, stmt,
        resource.ID,
        resource.Field1,
        resource.Field2,
    )
    if err != nil {
        return err
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }

    if rowsAffected == 0 {
        return fmt.Errorf("no resource found with id %s", resource.ID)
    }

    return nil
}

// Delete operation
func (s *ResourceStore) Delete(ctx context.Context, id string) error {
    stmt := `DELETE FROM resources WHERE id = $1`

    result, err := s.db.Exec(ctx, stmt, id)
    if err != nil {
        return err
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }

    if rowsAffected == 0 {
        return fmt.Errorf("no resource found with id %s", id)
    }

    return nil
}
```

**Query Patterns**:
- Use parameterized queries ($1, $2, etc.) to prevent SQL injection
- Always check `rows.Err()` after iteration
- Close rows with `defer rows.Close()`
- Use `QueryRow` for single results, `Query` for multiple
- Use `Exec` for operations that don't return rows

### Model Layer

**Location**: `model/`

**Responsibilities**:
- Data structure definitions
- Business logic helpers (formatting, validation)
- Type conversions
- Constants and enums

**Pattern**:
```go
package model

import (
    "database/sql"
    "fmt"
    "time"

    "github.com/guregu/null/v5"
)

// Struct definition with tags
type Resource struct {
    ID          string         `json:"id" db:"id"`
    Field1      string         `json:"field1" db:"field1"`
    Field2      int            `json:"field2" db:"field2"`
    OptionalStr null.String    `json:"optional_str" db:"optional_str"` // nullable DB field
    OptionalInt *int           `json:"optional_int,omitempty"`         // optional primitive
    CreatedAt   time.Time      `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time      `json:"updated_at" db:"updated_at"`
}

// Constants for enum-like values
const (
    ResourceStatusActive   = "active"
    ResourceStatusInactive = "inactive"
)

// Boolean checkers
func (r *Resource) IsActive() bool {
    return r.Status == ResourceStatusActive
}

// Formatters (with timezone handling)
func (r *Resource) FormatCreatedAt() string {
    loc, err := time.LoadLocation("America/Los_Angeles")
    if err != nil {
        return r.CreatedAt.Format("Jan 2, 2006")
    }

    localTime := r.CreatedAt.In(loc)
    return localTime.Format("Jan 2, 2006 at 3:04 PM")
}

// Getters with safe defaults
func (r *Resource) GetOptionalStr() string {
    if r.OptionalStr.Valid {
        return r.OptionalStr.String
    }
    return "default value"
}

// Validators
func (r *Resource) Validate() error {
    if r.Field1 == "" {
        return fmt.Errorf("field1 is required")
    }
    if r.Field2 < 0 {
        return fmt.Errorf("field2 must be non-negative")
    }
    return nil
}
```

**Timezone Handling**:
- Store timestamps in UTC in database
- Convert to local timezone when formatting for display
- Use consistent timezone throughout application
- Handle timezone loading errors gracefully

**Nullable Types**:
- `null.String`, `null.Time` from `guregu/null/v5` for DB nullable fields
- `*string`, `*int`, `*time.Time` for optional primitives
- Plain types for required fields

### Middleware Layer

**Location**: `middleware/`

**Responsibilities**:
- Request/response interception
- Authentication and authorization
- Session validation
- Context enrichment

**Pattern**:
```go
func RequireAuth(sqlDB *sql.DB) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // 1. Extract authentication token
        sessionCookie := c.Cookies("session")

        // 2. Validate authentication
        if sessionCookie == "" {
            return c.Redirect("/signin", fiber.StatusTemporaryRedirect)
        }

        // 3. Instantiate dependencies
        userStore := store.NewUserStore(sqlDB)
        authService := service.NewAuthServiceImpl(userStore, jwtSecret)

        // 4. Validate session
        ctx := c.Context()
        user, err := authService.GetUserFromSession(ctx, sessionCookie)
        if err != nil {
            return c.Redirect("/signin", fiber.StatusTemporaryRedirect)
        }

        // 5. Store user in context for handlers
        c.Locals("user", user)

        // 6. Continue to next handler
        return c.Next()
    }
}
```

**Middleware Registration**:
```go
// Global middleware
app.Use(logger.New())

// Route-specific middleware
app.Get("/protected", middleware.RequireAuth(db), handler.HandleProtected(db))

// Group middleware
admin := app.Group("/admin")
admin.Use(basicauth.New(basicauth.Config{...}))
```

## Dependency Injection

The application uses **constructor-based dependency injection** throughout.

### Pattern

**Store Instantiation**:
```go
// In handler
resourceStore := store.NewResourceStore(dbWrapper)
```

**Service Instantiation**:
```go
// Single dependency
resourceService := service.NewResourceServiceImpl(resourceStore)

// Multiple dependencies
complexService := service.NewComplexServiceImpl(
    primaryStore,
    secondaryStore,
    externalService,
    configValue,
)
```

### Benefits

1. **Testability**: Dependencies can be mocked/stubbed
2. **Flexibility**: Implementation can be swapped without changing consumers
3. **Clarity**: Dependencies are explicit in constructor signature
4. **Lifecycle Control**: Instantiation happens at the right layer

### Dependency Flow

```
Handler
  ├─ Instantiates Store(s)
  ├─ Instantiates Service(s) with Store(s)
  └─ Calls Service method(s)

Service
  ├─ Uses injected Store(s)
  └─ Calls Store method(s)

Store
  ├─ Uses injected DB connection
  └─ Executes SQL queries
```

## Error Handling

### Patterns

**Handler Layer**:
```go
result, err := service.PerformOperation(ctx, params)
if err != nil {
    return c.Status(fiber.StatusBadRequest).SendString(err.Error())
}
```

**Service Layer**:
```go
// Validation errors
if input < 0 {
    return fmt.Errorf("input must be non-negative")
}

// Wrapped errors
result, err := s.store.Query(ctx, id)
if err != nil {
    return fmt.Errorf("failed to query resource: %w", err)
}
```

**Store Layer**:
```go
// Database errors bubble up
err := s.db.QueryRow(ctx, stmt, id).Scan(&resource)
if err != nil {
    return err
}

// Not found checks
if rowsAffected == 0 {
    return fmt.Errorf("no resource found with id %s", id)
}
```

### HTTP Status Codes

- `200 OK`: Successful request
- `302 Found`: Redirect
- `400 Bad Request`: Validation error, malformed input
- `401 Unauthorized`: Authentication required
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Unexpected server error

## Database Patterns

### Connection Management

**Initialization** (in `cmd/serve.go`):
```go
db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
if err != nil {
    panic(err)
}
defer db.Close()

if err = db.Ping(); err != nil {
    panic(err)
}
```

**Passing to Handlers**:
```go
app.Get("/resource", handler.HandleResource(db))
```

### Context Usage

All database operations accept `context.Context` for:
- Timeout handling
- Cancellation propagation
- Request tracing

```go
func (s *Store) FindByID(ctx context.Context, id string) (*model.Resource, error) {
    return s.db.QueryRow(ctx, stmt, id).Scan(...)
}
```

### Migration Pattern

- Migrations stored in `db/migrations/`
- SQL-based migrations
- Version controlled in git
- Run manually or via deployment process

## Webhook Handling

### Stripe Webhook Pattern

**Handler**:
```go
func HandleStripeWebhook(db *sql.DB) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // 1. Verify webhook signature
        payload := c.Body()
        signatureHeader := string(c.Request().Header.Peek("Stripe-Signature"))
        event, err := webhook.ConstructEvent(payload, signatureHeader, endpointSecret)

        // 2. Route based on event type
        switch event.Type {
        case "checkout.session.completed":
            var session stripe.CheckoutSession
            json.Unmarshal(event.Data.Raw, &session)

            // 3. Route based on metadata
            if product := session.Metadata["product"]; product == "type_a" {
                return processTypeA(db, &session)
            } else if _, hasTypeB := session.Metadata["type_b_id"]; hasTypeB {
                return processTypeB(db, &session)
            }
        }

        return c.SendStatus(fiber.StatusOK)
    }
}
```

**Processing Functions**:
```go
func processPaymentType(sqlDB *sql.DB, session *stripe.CheckoutSession) error {
    // 1. Wrap DB
    dbWrapper := &db.DB{DB: sqlDB}

    // 2. Instantiate dependencies
    store := store.NewResourceStore(dbWrapper)
    service := service.NewResourceServiceImpl(store)

    // 3. Process webhook
    ctx := context.Background()
    err := service.ProcessWebhook(ctx, session.ID, session.Metadata)
    if err != nil {
        log.Printf("Failed to process webhook: %v", err)
        return err
    }

    return nil
}
```

### Metadata Routing

Stripe checkout sessions use metadata for routing:
```go
Metadata: map[string]string{
    "product": "resource_type",
    "resource_id": resourceID,
    "user_id": userID,
}
```

Webhook handler checks metadata to determine processing path:
```go
if product := session.Metadata["product"]; product == "type_a" {
    processTypeA(db, &session)
} else if _, hasTypeB := session.Metadata["type_b_id"]; hasTypeB {
    processTypeB(db, session.ID, session.Metadata)
}
```

## Template Integration

### Templ Templates

**Location**: `component/`

**Handler Pattern**:
```go
page := component.TemplateName(data1, data2)
handler := adaptor.HTTPHandler(templ.Handler(page))
return handler(c)
```

**Type Conversions**:
- Integers need `strconv.Itoa(value)` in templates
- URLs need `templ.SafeURL("/path/" + id)` wrapper
- Add imports to template files as needed

**Layout Wrappers**:
- Use consistent layout components
- Pass data through layout to child templates

## Best Practices

### Handler Layer
1. Keep handlers thin - delegate to services
2. Instantiate dependencies at handler scope
3. Validate input early
4. Return appropriate HTTP status codes
5. Use descriptive error messages

### Service Layer
1. Validate business rules before database operations
2. Return descriptive errors with context
3. Handle external API failures gracefully
4. Log important events and errors
5. Use environment-aware behavior (dev vs prod)

### Store Layer
1. Use parameterized queries to prevent SQL injection
2. Always defer `rows.Close()`
3. Check `rows.Err()` after iteration
4. Return descriptive errors for not found cases
5. Use transactions for multi-step operations

### Model Layer
1. Use struct tags for JSON and DB mapping
2. Use nullable types appropriately
3. Provide helper methods for common operations
4. Handle timezones consistently
5. Validate data with explicit error messages

### General
1. Use `context.Context` for timeout/cancellation
2. Log with appropriate detail (include IDs, amounts, etc.)
3. Fail fast with clear error messages
4. Don't swallow errors - wrap and return them
5. Use environment variables for configuration

## Example: Complete Feature Implementation

Here's how all layers work together for a complete feature:

### 1. Define Model
```go
// model/resource.go
type Resource struct {
    ID          string
    Name        string
    Status      string
    Amount      int
    CreatedAt   time.Time
}

func (r *Resource) FormatAmount() string {
    return fmt.Sprintf("$%d.%02d", r.Amount/100, r.Amount%100)
}
```

### 2. Create Store
```go
// store/resource_store.go
type ResourceStore struct {
    db *db.DB
}

func NewResourceStore(database *db.DB) *ResourceStore {
    return &ResourceStore{db: database}
}

func (s *ResourceStore) Create(ctx context.Context, resource *model.Resource) error {
    stmt := `INSERT INTO resources (...) VALUES (...) RETURNING id, created_at`
    return s.db.QueryRow(ctx, stmt, ...).Scan(&resource.ID, &resource.CreatedAt)
}
```

### 3. Build Service
```go
// service/resource_service.go
type ResourceServiceImpl struct {
    resourceStore *store.ResourceStore
}

func NewResourceServiceImpl(resourceStore *store.ResourceStore) *ResourceServiceImpl {
    return &ResourceServiceImpl{resourceStore: resourceStore}
}

func (s *ResourceServiceImpl) CreateResource(ctx context.Context, resource *model.Resource) error {
    // Validate business rules
    if resource.Amount < 0 {
        return fmt.Errorf("amount cannot be negative")
    }

    return s.resourceStore.Create(ctx, resource)
}
```

### 4. Create Handler
```go
// handler/resource_handler.go
func HandleCreateResource(sqlDB *sql.DB) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Instantiate dependencies
        dbWrapper := &db.DB{DB: sqlDB}
        resourceStore := store.NewResourceStore(dbWrapper)
        resourceService := service.NewResourceServiceImpl(resourceStore)

        // Parse form
        resource, err := parseResourceForm(c)
        if err != nil {
            return c.Status(fiber.StatusBadRequest).SendString(err.Error())
        }

        // Call service
        ctx := context.Background()
        err = resourceService.CreateResource(ctx, resource)
        if err != nil {
            return c.Status(fiber.StatusBadRequest).SendString(err.Error())
        }

        // Redirect to list
        return c.Redirect("/admin/resources")
    }
}
```

### 5. Register Route
```go
// cmd/serve.go
admin.Post("/resources", handler.HandleCreateResource(db))
```

This layered approach provides:
- Clear separation of concerns
- Testability at each layer
- Flexibility to change implementations
- Maintainable and scalable codebase
