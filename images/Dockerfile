# Development/Build image
FROM golang:1.23-alpine

# Install build dependencies
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Install development tools
RUN go install github.com/a-h/templ/cmd/templ@latest
RUN go install github.com/air-verse/air@v1.61.7

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Generate templ files initially (air will regenerate on changes)
RUN templ generate

EXPOSE 8080

# Default command uses air for hot-reload
# For production build, override with: docker build --target builder -o ./bin .
CMD ["air", "-c", ".air.toml"]
