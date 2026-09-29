# go-api-template

A production-ready Go REST API template built **entirely on the standard library** — no third-party HTTP framework, no third-party router, no external config library. Zero entries in `go.mod`.

## Quick Start

```bash
# Clone / use as template
git clone <your-fork> && cd go-api-template

# Run (no go get, no setup)
go run ./cmd/api

# Or build and run
go build -o bin/api ./cmd/api && ./bin/api
```

The server starts on `http://localhost:8080` by default.

## Routes

| Method   | Path          | Description             |
|----------|---------------|-------------------------|
| `GET`    | `/health`     | Health probe            |
| `GET`    | `/todos`      | List all todos          |
| `POST`   | `/todos`      | Create a todo           |
| `GET`    | `/todos/{id}` | Get a todo by ID        |
| `PUT`    | `/todos/{id}` | Replace a todo          |
| `DELETE` | `/todos/{id}` | Delete a todo           |

### Example requests

```bash
# Health check
curl http://localhost:8080/health

# Create
curl -X POST http://localhost:8080/todos \
  -H "Content-Type: application/json" \
  -d '{"title":"Buy milk"}'

# List
curl http://localhost:8080/todos

# Update
curl -X PUT http://localhost:8080/todos/1 \
  -H "Content-Type: application/json" \
  -d '{"title":"Buy oat milk","done":true}'

# Delete
curl -X DELETE http://localhost:8080/todos/1
```

All responses are wrapped in a consistent envelope:

```json
{ "data": { ... } }
{ "error": "not found" }
{ "error": "validation failed", "details": { "title": "required" } }
```

## Configuration

All settings are read from environment variables. Defaults are chosen to work out of the box for local development.

| Variable           | Default | Description                                                       |
|--------------------|---------|-------------------------------------------------------------------|
| `HOST`             | `""`    | Interface to bind (`""` = all interfaces, `"127.0.0.1"` = local) |
| `PORT`             | `8080`  | TCP port                                                          |
| `READ_TIMEOUT`     | `5s`    | Max duration to read the full request (including body)            |
| `WRITE_TIMEOUT`    | `10s`   | Max duration to write the full response                           |
| `IDLE_TIMEOUT`     | `120s`  | Max idle time on a keep-alive connection                          |
| `SHUTDOWN_TIMEOUT` | `15s`   | How long graceful shutdown waits for in-flight requests           |
| `TRUSTED_PROXY`    | `false` | Trust `X-Forwarded-For` / `X-Real-IP` headers from upstream proxy |

### Running behind a proxy

If you deploy behind nginx, Caddy, or a cloud load balancer, set `TRUSTED_PROXY=true` so the logged `remote_ip` reflects the real client IP rather than the proxy's address.

```bash
TRUSTED_PROXY=true PORT=3000 go run ./cmd/api
```

> **Security**: Only enable `TRUSTED_PROXY` when the server is not directly reachable by untrusted clients. Enabling it on a public server allows IP spoofing.

## Project Layout

```
go-api-template/
├── cmd/
│   └── api/
│       └── main.go              # Entry point: wire everything, run server
├── internal/
│   ├── api/
│   │   ├── router.go            # ServeMux setup, route registration
│   │   ├── middleware.go        # RequestID, RealIP, Logger, Recoverer, CORS
│   │   ├── respond.go           # Shared JSON response helpers
│   │   └── handlers/
│   │       ├── health.go        # GET /health
│   │       └── todos.go         # CRUD handlers for /todos
│   ├── store/
│   │   └── memory.go            # In-memory TodoStore (sync.RWMutex)
│   └── config/
│       └── config.go            # Env-based config
├── go.mod                       # Zero external dependencies
├── Makefile
└── README.md
```

`internal/` enforces Go's package visibility rules — nothing outside this module can import these packages.

## What's Included

- **`net/http` ServeMux** with Go 1.22+ method-prefixed patterns (`GET /todos/{id}`) — no third-party router needed
- **`log/slog`** structured JSON logging (Go 1.21+) — threads request ID through context
- **Middleware chain**: RequestID injection, RealIP (proxy-aware), request/response logging, panic recovery, CORS
- **Graceful shutdown** on `SIGINT` / `SIGTERM` with configurable timeout
- **Consistent JSON envelope** (`{"data":...}` / `{"error":...}` / `{"error":"validation failed","details":{...}}`)
- **In-memory store** behind a `TodoStore` interface — swap for a real DB without touching handlers
- **Tests** using only `testing` + `net/http/httptest`

## Adding a Database

The `store.TodoStore` interface is your contract. To swap to PostgreSQL:

1. `go get github.com/jackc/pgx/v5`
2. Create `internal/store/postgres.go` implementing `TodoStore`
3. In `cmd/api/main.go`, replace `store.NewMemoryStore()` with your new constructor

Handlers and middleware are completely unaffected.

## Development

```bash
make run      # Run the server
make test     # Run all tests with race detector
make lint     # go vet + gofmt check
make build    # Compile to bin/api
```

## License

MIT — see [LICENSE](LICENSE).