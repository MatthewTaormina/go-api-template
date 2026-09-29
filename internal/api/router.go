// Package api — router.go
// NewRouter builds the ServeMux and wraps it in the middleware chain.
// Add new routes here; add new middleware in middleware.go.
package api

import (
	"net/http"

	"github.com/user/go-api-template/internal/api/handlers"
	"github.com/user/go-api-template/internal/config"
	"github.com/user/go-api-template/internal/store"
)

// NewRouter constructs the fully-wired http.Handler for the application.
// Route patterns use the Go 1.22+ ServeMux syntax:
//
//	"METHOD /path"         — exact match, specific method
//	"METHOD /path/{id}"   — path parameter captured as r.PathValue("id")
//
// Middleware is applied from outermost to innermost:
//
//	CORS → RequestID → RealIP → Logger → Recoverer → mux
func NewRouter(cfg config.Config, s store.TodoStore) http.Handler {
	mux := http.NewServeMux()

	// ── Health ──────────────────────────────────────────────────────────────
	mux.HandleFunc("GET /health", handlers.Health)

	// ── Todos ───────────────────────────────────────────────────────────────
	todos := handlers.NewTodoHandler(s)
	mux.HandleFunc("GET /todos", todos.List)
	mux.HandleFunc("POST /todos", todos.Create)
	mux.HandleFunc("GET /todos/{id}", todos.Get)
	mux.HandleFunc("PUT /todos/{id}", todos.Update)
	mux.HandleFunc("DELETE /todos/{id}", todos.Delete)

	// ── Middleware chain (outermost wraps last) ──────────────────────────────
	var h http.Handler = mux
	h = Recoverer(h)
	h = Logger(h)
	h = RealIP(cfg)(h)
	h = RequestID(h)
	h = CORS(h)

	return h
}
