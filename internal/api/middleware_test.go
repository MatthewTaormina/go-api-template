package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/go-api-template/internal/api"
	"github.com/user/go-api-template/internal/config"
)

// ── RequestID middleware ────────────────────────────────────────────────────

func TestRequestID_GeneratesID(t *testing.T) {
	handler := api.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := api.RequestIDFromContext(r.Context())
		if id == "" {
			t.Error("RequestID: context has empty request ID")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Header().Get(api.RequestIDHeader) == "" {
		t.Error("RequestID: response header X-Request-ID is empty")
	}
}

func TestRequestID_PropagatesClientID(t *testing.T) {
	const clientID = "my-trace-id"

	handler := api.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := api.RequestIDFromContext(r.Context())
		if id != clientID {
			t.Errorf("RequestID: context got %q, want %q", id, clientID)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(api.RequestIDHeader, clientID)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get(api.RequestIDHeader); got != clientID {
		t.Errorf("RequestID response header: got %q, want %q", got, clientID)
	}
}

// ── RealIP middleware ───────────────────────────────────────────────────────

func TestRealIP_TrustedProxy_XForwardedFor(t *testing.T) {
	cfg := config.Config{TrustedProxy: true}

	capturedAddr := ""
	handler := api.RealIP(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAddr = r.RemoteAddr
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	req.Header.Set("X-Forwarded-For", "1.2.3.4, 5.6.7.8")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if capturedAddr != "1.2.3.4:12345" {
		t.Errorf("RealIP: RemoteAddr got %q, want %q", capturedAddr, "1.2.3.4:12345")
	}
}

func TestRealIP_Untrusted_DoesNotOverwrite(t *testing.T) {
	cfg := config.Config{TrustedProxy: false}

	capturedAddr := ""
	handler := api.RealIP(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAddr = r.RemoteAddr
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	req.Header.Set("X-Forwarded-For", "1.2.3.4")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if capturedAddr != "10.0.0.1:12345" {
		t.Errorf("RealIP (untrusted): RemoteAddr got %q, want original %q", capturedAddr, "10.0.0.1:12345")
	}
}

// ── Recoverer middleware ────────────────────────────────────────────────────

func TestRecoverer_CatchesPanic(t *testing.T) {
	handler := api.Recoverer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("something went wrong")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	// Should not panic; Recoverer must catch it and return 500.
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Recoverer: status got %d, want 500", rec.Code)
	}
}

// ── CORS middleware ─────────────────────────────────────────────────────────

func TestCORS_PreflightReturns204(t *testing.T) {
	handler := api.CORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/todos", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("CORS preflight: status got %d, want 204", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Error("CORS: Access-Control-Allow-Origin header missing")
	}
}

func TestCORS_SetsHeadersOnNormalRequest(t *testing.T) {
	handler := api.CORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/todos", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("CORS normal: status got %d, want 200", rec.Code)
	}
	if h := rec.Header().Get("Access-Control-Allow-Origin"); h != "*" {
		t.Errorf("CORS: Access-Control-Allow-Origin got %q, want %q", h, "*")
	}
}
