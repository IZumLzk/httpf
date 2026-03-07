# AGENTS.md - GoSpaceX Framework Development Guide

This document provides essential information for AI coding agents working in the GoSpaceX repository.

## Build, Lint, and Test Commands

### Build
```bash
# Build the main application
go build -o bin/gospacex ./cmd/gospacex

# Build for all platforms
GOOS=linux GOARCH=amd64 go build -o bin/gospacex-linux ./cmd/gospacex
GOOS=windows GOARCH=amd64 go build -o bin/gospacex.exe ./cmd/gospacex
```

### Test Commands
```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for a specific package
go test -v ./core/protocol/http

# Run a specific test by name
go test -v ./core/protocol/http -run ^TestGinAdapter_AddRoute_GET$

# Run tests with coverage
go test -cover ./...

# Run tests with detailed coverage report
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out

# Run benchmark tests
go test -bench=. ./...

# Run a specific benchmark
go test -bench=BenchmarkMiddlewareFunc_Process ./core/protocol/http
```

### Lint Commands
```bash
# Format code
go fmt ./...

# Run go vet (static analysis)
go vet ./...

# Run staticcheck (if installed)
staticcheck ./...

# Run golangci-lint (if installed)
golangci-lint run
```

### Dependency Management
```bash
# Download dependencies
go mod download

# Tidy dependencies
go mod tidy

# Verify dependencies
go mod verify
```

## Project Structure

```
GoSpaceX/
├── cmd/gospacex/           # Application entry points
├── core/                   # Core framework components
│   ├── common/             # Shared interfaces and types
│   ├── config/             # Configuration loading
│   ├── protocol/
│   │   ├── http/           # HTTP abstraction layer
│   │   │   ├── *_adapter.go    # Framework adapters (Gin, Iris, Echo, Fiber, Chi, Hertz)
│   │   │   ├── factory.go      # Framework factory
│   │   │   ├── interface.go    # Type aliases
│   │   │   ├── server.go       # HTTP launcher
│   │   │   └── middleware/     # Middleware implementations
│   │   ├── rpc/            # RPC protocols
│   │   └── websocket/      # WebSocket support
│   ├── registry/           # Service registry
│   └── storage/            # Storage abstractions
├── examples/               # Usage examples
├── internal/               # Internal packages
│   ├── bootstrapper/       # Application bootstrap
│   └── container/          # Dependency injection
├── Documentation/          # Architecture docs
├── go.mod                  # Module: gospacex, Go 1.24.10
└── README.md               # Project documentation
```

## Core Interfaces

### HTTPProtocol Interface
```go
type HTTPProtocol interface {
    Start() error
    Stop() error
    AddRoute(method, path string, handler HTTPHandlerFunc)
    UseMiddlewares(mids ...HTTPMiddleware) HTTPProtocol
    AddGlobalMiddleware(middleware HTTPMiddleware)
}
```

### HTTPContext Interface
```go
type HTTPContext interface {
    JSON(code int, obj interface{})
    Param(key string) string
    Query(key string) string
    Bind(obj interface{}) error
    Request() *http.Request
    ResponseWriter() http.ResponseWriter
}
```

### HTTPMiddleware Interface
```go
type HTTPMiddleware interface {
    Process(ctx HTTPContext) bool  // Returns false to abort chain
}

type HTTPMiddlewareFunc func(ctx HTTPContext) bool
```

### HTTPHandlerFunc
```go
type HTTPHandlerFunc func(HTTPContext)
```

## Code Style Guidelines

### Imports Organization
```go
import (
    // 1. Standard library
    "context"
    "fmt"
    "net/http"
    
    // 2. This project's packages
    "gospacex/core/common"
    "gospacex/core/protocol/http/middleware"
    
    // 3. Third-party packages (alphabetically)
    "github.com/gin-gonic/gin"
    jwt "github.com/golang-jwt/jwt/v4"  // Use aliases when needed
    "github.com/kataras/iris/v12"
)
```

### Naming Conventions

- **Packages**: lowercase, single word preferred (e.g., `http`, `middleware`, `config`)
- **Interfaces**: noun or verb + "er" suffix (e.g., `HTTPProtocol`, `HTTPContext`)
- **Structs**: PascalCase for exported, camelCase for private
- **Functions**: `New<Component>()` for constructors, `Get<Thing>()` for factories
- **Constants**: PascalCase for exported (e.g., `GinFramework`, `HTTPMethod`)
- **Test functions**: `Test<Component>_<Behavior>(t *testing.T)` (e.g., `TestGinAdapter_AddRoute_GET`)

### Type Aliases Pattern
Use type aliases for clean package exports:
```go
// In interface.go
type HTTPProtocol = common.HTTPProtocol
type HandlerFunc = common.HTTPHandlerFunc
type Context = common.HTTPContext
```

### Middleware Pattern
```go
// Constructor returns HTTPMiddlewareFunc
func NewAuthJWTMW() common.HTTPMiddlewareFunc {
    return (&AuthJWTMiddleware{secret: "secret"}).Process
}

// Process method returns bool (false aborts chain)
func (am *AuthJWTMiddleware) Process(ctx common.HTTPContext) bool {
    // Middleware logic
    if !authorized {
        ctx.JSON(401, map[string]interface{}{"error": "unauthorized"})
        return false  // Abort middleware chain
    }
    return true  // Continue to next middleware/handler
}
```

### Adapter Pattern
Each HTTP framework adapter implements `HTTPProtocol`:
```go
type GinAdapter struct {
    Engine      *gin.Engine
    Config      HTTPConfig
    Middlewares []Middleware
}

func NewGinAdapter(config HTTPConfig) *GinAdapter {
    gin.SetMode(gin.ReleaseMode)
    return &GinAdapter{
        Engine:      gin.New(),
        Config:      config,
        Middlewares: make([]Middleware, 0),
    }
}
```

### Factory Pattern
```go
func GetHTTPFramework(framework HTTPFrameworkType) HTTPFactory {
    switch framework {
    case GinFramework:
        return &ginHTTPFactory{}
    // ... other frameworks
    default:
        return &ginHTTPFactory{}  // Default fallback
    }
}
```

### Error Handling
```go
// Return errors from functions
func (ga *GinAdapter) Start() error {
    addr := ga.Config.Host + ":" + fmt.Sprintf("%d", ga.Config.Port)
    return ga.Engine.Run(addr)
}

// Check errors in handlers
func handler(ctx Context) {
    var req Request
    if err := ctx.Bind(&req); err != nil {
        ctx.JSON(400, map[string]string{"error": "invalid request"})
        return
    }
    // ... process request
}
```

## Testing Guidelines

### Test File Structure
```go
package http

import (
    "net/http/httptest"
    "testing"
    
    "github.com/gin-gonic/gin"
)

// Set test mode for frameworks
func init() {
    gin.SetMode(gin.TestMode)
}

func TestGinAdapter_AddRoute_GET(t *testing.T) {
    // Setup
    adapter := NewGinAdapter(HTTPConfig{Framework: GinFramework})
    
    // Test
    adapter.AddRoute("GET", "/test", handler)
    
    // Verify with httptest
    req := httptest.NewRequest("GET", "/test", nil)
    recorder := httptest.NewRecorder()
    adapter.Engine.ServeHTTP(recorder, req)
    
    // Assert
    if recorder.Code != 200 {
        t.Errorf("Expected 200, got %d", recorder.Code)
    }
}

// Use subtests for related tests
func TestComponent(t *testing.T) {
    t.Run("case1", func(t *testing.T) { ... })
    t.Run("case2", func(t *testing.T) { ... })
}
```

### Table-Driven Tests
```go
func TestParseNodeAddress(t *testing.T) {
    tests := []struct {
        name         string
        address      string
        expectedHost string
        expectedPort int
    }{
        {"standard", "localhost:8080", "localhost", 8080},
        {"IP address", "192.168.1.10:9090", "192.168.1.10", 9090},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            config := parseNodeAddress(tt.address)
            if config.Host != tt.expectedHost {
                t.Errorf("Expected %s, got %s", tt.expectedHost, config.Host)
            }
        })
    }
}
```

### Mock Context for Testing
```go
type MockContext struct {
    responseWriter http.ResponseWriter
    request        *http.Request
    jsonCode       int
    jsonData       interface{}
}

func (mc *MockContext) JSON(code int, obj interface{}) {
    mc.jsonCode = code
    mc.jsonData = obj
}
// ... implement other Context methods
```

## Supported HTTP Frameworks

| Framework | Constant | Import |
|-----------|----------|--------|
| Gin | `GinFramework` | `github.com/gin-gonic/gin` |
| Iris | `IrisFramework` | `github.com/kataras/iris/v12` |
| Echo | `EchoFramework` | (simplified implementation) |
| Fiber | `FiberFramework` | `github.com/gofiber/fiber/v2` |
| Chi | `ChiFramework` | `github.com/go-chi/chi/v5` |
| Hertz | `HertzFramework` | (simplified implementation) |

## Key Architecture Patterns

1. **Adapter Pattern** - Each HTTP framework has an adapter implementing `HTTPProtocol`
2. **Factory Pattern** - `GetHTTPFramework()` returns framework-specific factory
3. **Strategy Pattern** - Different runtime behaviors via framework selection
4. **Middleware Chain** - Responsibility chain via `HTTPMiddleware.Process()`

## Common Tasks

### Add a New HTTP Framework Adapter
1. Create `xxx_adapter.go` in `core/protocol/http/`
2. Implement `HTTPProtocol` interface
3. Create `XXXContext` implementing `HTTPContext`
4. Add factory in `factory.go`
5. Add constant in `factory.go`
6. Create comprehensive `xxx_adapter_test.go`

### Add New Middleware
1. Create `xxx_middleware.go` in `core/protocol/http/middleware/`
2. Define struct with configuration
3. Implement `Process(ctx HTTPContext) bool` method
4. Create `NewXXXMW()` constructor
5. Create `xxx_middleware_test.go`

### Add Route Handler
```go
func MyHandler(ctx http.Context) {
    // 1. Get parameters
    id := ctx.Param("id")
    page := ctx.Query("page")
    
    // 2. Bind request body
    var req MyRequest
    if err := ctx.Bind(&req); err != nil {
        ctx.JSON(400, map[string]string{"error": "invalid request"})
        return
    }
    
    // 3. Process business logic
    result, err := process(id, req)
    if err != nil {
        ctx.JSON(500, map[string]string{"error": err.Error()})
        return
    }
    
    // 4. Return response
    ctx.JSON(200, result)
}
```

## Important Notes

- **Go Version**: 1.24.10
- **Module Name**: `gospacex`
- **Never use type assertions** (`as any`, `@ts-ignore`) - not applicable in Go
- **Always check errors** - Go's error handling is explicit
- **Use context for cancellation** - Pass `context.Context` for long-running operations
- **Keep adapters consistent** - All adapters should follow the same patterns
- **Write comprehensive tests** - Use `httptest` for HTTP testing, aim for high coverage
- **Follow existing patterns** - Look at existing adapters/middleware for reference