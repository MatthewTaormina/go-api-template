package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/go-api-template/internal/api/handlers"
	"github.com/user/go-api-template/internal/store"
)

// newTodoServer spins up an httptest.Server wired to a fresh MemoryStore.
// It returns the server and its base URL. The server is cleaned up automatically.
func newTodoServer(t *testing.T) (*httptest.Server, store.TodoStore) {
	t.Helper()
	s := store.NewMemoryStore()
	h := handlers.NewTodoHandler(s)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /todos", h.List)
	mux.HandleFunc("POST /todos", h.Create)
	mux.HandleFunc("GET /todos/{id}", h.Get)
	mux.HandleFunc("PUT /todos/{id}", h.Update)
	mux.HandleFunc("DELETE /todos/{id}", h.Delete)

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, s
}

// jsonBody encodes v as JSON and returns a *bytes.Buffer suitable as an http.Request body.
func jsonBody(t *testing.T, v any) *bytes.Buffer {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("jsonBody: %v", err)
	}
	return bytes.NewBuffer(b)
}

// decode unmarshals a JSON response body's "data" field into v.
func decode(t *testing.T, resp *http.Response, v any) {
	t.Helper()
	var env map[string]json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if data, ok := env["data"]; ok {
		if err := json.Unmarshal(data, v); err != nil {
			t.Fatalf("decode data: %v", err)
		}
	}
}

// ── Health ─────────────────────────────────────────────────────────────────

func TestHealth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(handlers.Health))
	t.Cleanup(srv.Close)

	resp, err := srv.Client().Get(srv.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Health: status got %d, want 200", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Health: Content-Type got %q, want application/json", ct)
	}
}

// ── Todos ──────────────────────────────────────────────────────────────────

func TestTodoList_Empty(t *testing.T) {
	srv, _ := newTodoServer(t)

	resp, err := srv.Client().Get(srv.URL + "/todos")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("List: status got %d, want 200", resp.StatusCode)
	}

	var todos []store.Todo
	decode(t, resp, &todos)
	if len(todos) != 0 {
		t.Errorf("List: got %d items, want 0", len(todos))
	}
}

func TestTodoCreate(t *testing.T) {
	srv, _ := newTodoServer(t)

	body := jsonBody(t, map[string]string{"title": "Write tests"})
	resp, err := srv.Client().Post(srv.URL+"/todos", "application/json", body)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Create: status got %d, want 201", resp.StatusCode)
	}

	var todo store.Todo
	decode(t, resp, &todo)
	if todo.ID == "" {
		t.Error("Create: expected non-empty ID")
	}
	if todo.Title != "Write tests" {
		t.Errorf("Create: Title got %q, want %q", todo.Title, "Write tests")
	}
}

func TestTodoCreate_ValidationError(t *testing.T) {
	srv, _ := newTodoServer(t)

	body := jsonBody(t, map[string]string{"title": "   "})
	resp, err := srv.Client().Post(srv.URL+"/todos", "application/json", body)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("Create (empty title): status got %d, want 422", resp.StatusCode)
	}
}

func TestTodoCreate_BadJSON(t *testing.T) {
	srv, _ := newTodoServer(t)

	body := bytes.NewBufferString("not json")
	resp, err := srv.Client().Post(srv.URL+"/todos", "application/json", body)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Create (bad JSON): status got %d, want 400", resp.StatusCode)
	}
}

func TestTodoCreate_UnknownField(t *testing.T) {
	srv, _ := newTodoServer(t)

	body := jsonBody(t, map[string]string{"title": "Good", "unknown": "bad"})
	resp, err := srv.Client().Post(srv.URL+"/todos", "application/json", body)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// DisallowUnknownFields should cause a 400.
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Create (unknown field): status got %d, want 400", resp.StatusCode)
	}
}

func TestTodoGet(t *testing.T) {
	srv, s := newTodoServer(t)
	todo := s.Create("Find me")

	resp, err := srv.Client().Get(srv.URL + "/todos/" + todo.ID)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Get: status got %d, want 200", resp.StatusCode)
	}

	var got store.Todo
	decode(t, resp, &got)
	if got.ID != todo.ID {
		t.Errorf("Get: ID got %q, want %q", got.ID, todo.ID)
	}
}

func TestTodoGet_NotFound(t *testing.T) {
	srv, _ := newTodoServer(t)

	resp, err := srv.Client().Get(srv.URL + "/todos/999")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Get (missing): status got %d, want 404", resp.StatusCode)
	}
}

func TestTodoUpdate(t *testing.T) {
	srv, s := newTodoServer(t)
	todo := s.Create("Old title")

	body := jsonBody(t, map[string]any{"title": "New title", "done": true})
	req, _ := http.NewRequest(http.MethodPut, srv.URL+"/todos/"+todo.ID, body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Update: status got %d, want 200", resp.StatusCode)
	}

	var got store.Todo
	decode(t, resp, &got)
	if got.Title != "New title" {
		t.Errorf("Update: Title got %q, want %q", got.Title, "New title")
	}
	if !got.Done {
		t.Error("Update: expected Done=true")
	}
}

func TestTodoUpdate_NotFound(t *testing.T) {
	srv, _ := newTodoServer(t)

	body := jsonBody(t, map[string]any{"title": "X", "done": false})
	req, _ := http.NewRequest(http.MethodPut, srv.URL+"/todos/999", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Update (missing): status got %d, want 404", resp.StatusCode)
	}
}

func TestTodoDelete(t *testing.T) {
	srv, s := newTodoServer(t)
	todo := s.Create("Delete me")

	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/todos/"+todo.ID, nil)
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Delete: status got %d, want 204", resp.StatusCode)
	}

	// Confirm gone.
	resp2, _ := srv.Client().Get(srv.URL + "/todos/" + todo.ID)
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusNotFound {
		t.Errorf("Get after Delete: status got %d, want 404", resp2.StatusCode)
	}
}

func TestTodoDelete_NotFound(t *testing.T) {
	srv, _ := newTodoServer(t)

	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/todos/999", nil)
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Delete (missing): status got %d, want 404", resp.StatusCode)
	}
}
