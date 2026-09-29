// Package api — middleware.go
// Each middleware is a plain func(http.Handler) http.Handler.
// They compose without a framework: wrap the innermost handler last.
//
// Applied order (outermost → innermost):
//
//	CORS → RequestID → RealIP → Logger → Recoverer → mux
package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/user/go-api-template/internal/config"
	"github.com/user/go-api-template/internal/respond"
)

// contextKey is an unexported type for context keys in this package,
// preventing collisions with keys defined in other packages.
type contextKey string

const (
	// RequestIDKey is the context key that stores the request ID string.
	RequestIDKey contextKey = "request_id"

	// RequestIDHeader is the HTTP header used to propagate the request ID.
	RequestIDHeader = "X-Request-ID"
)

// RequestIDFromContext retrieves the request ID stored in ctx, or "" if absent.
func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(RequestIDKey).(string)
	return id
}

// RequestID is a middleware that attaches a unique ID to every request.
// It first checks whether the client sent an X-Request-ID header (useful for
// tracing across services). If not, it generates a new random 16-byte hex ID.
// The ID is stored in the request context and echoed in the response header.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(RequestIDHeader)
		if id == "" {
			id = newRequestID()
		}
		w.Header().Set(RequestIDHeader, id)
		ctx := context.WithValue(r.Context(), RequestIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// newRequestID generates a cryptographically random 16-byte hex string.
func newRequestID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// RealIP returns a middleware that overwrites r.RemoteAddr with the first address
// found in X-Forwarded-For (or X-Real-IP) when cfg.TrustedProxy is true.
//
// SECURITY: Only enable TrustedProxy when you control the upstream proxy.
// Enabling it on a public-facing server allows clients to spoof their IP.
func RealIP(cfg config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.TrustedProxy {
				if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
					// X-Forwarded-For may be a comma-separated list; take the first (original client).
					parts := strings.SplitN(xff, ",", 2)
					ip := strings.TrimSpace(parts[0])
					if net.ParseIP(ip) != nil {
						// Preserve the original port if present; replace the host portion.
						_, port, err := net.SplitHostPort(r.RemoteAddr)
						if err != nil {
							port = "0"
						}
						r.RemoteAddr = net.JoinHostPort(ip, port)
					}
				} else if xri := r.Header.Get("X-Real-IP"); xri != "" {
					if net.ParseIP(xri) != nil {
						_, port, err := net.SplitHostPort(r.RemoteAddr)
						if err != nil {
							port = "0"
						}
						r.RemoteAddr = net.JoinHostPort(xri, port)
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// statusRecorder wraps http.ResponseWriter to capture the written status code
// after the handler has finished, for use in the Logger middleware.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.status = code
	sr.ResponseWriter.WriteHeader(code)
}

// Logger is a middleware that logs each completed HTTP request using slog.
// Fields logged: method, path, status, latency, remote_ip, request_id.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rec, r)

		ip, _, _ := net.SplitHostPort(r.RemoteAddr)

		slog.InfoContext(r.Context(), "request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"latency_ms", time.Since(start).Milliseconds(),
			"remote_ip", ip,
			"request_id", RequestIDFromContext(r.Context()),
		)
	})
}

// Recoverer is a middleware that catches panics, logs the stack trace,
// and returns a 500 Internal Server Error instead of crashing the server.
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.ErrorContext(r.Context(), "panic recovered",
					"err", err,
					"stack", string(debug.Stack()),
					"request_id", RequestIDFromContext(r.Context()),
				)
				respond.Error(w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// CORS returns middleware that adds permissive CORS headers suitable for local
// development. For production, replace the wildcard origin with a specific
// allowed-origins list and restrict the methods/headers as needed.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")

		// Respond immediately to preflight requests.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
