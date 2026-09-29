package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/user/go-api-template/internal/respond"
	"github.com/user/go-api-template/internal/store"
)

// TodoHandler groups the HTTP handlers for the /todos resource.
// It holds a reference to the store via the TodoStore interface so that
// tests can inject a fake implementation without touching the real store.
type TodoHandler struct {
	store store.TodoStore
}

// NewTodoHandler constructs a TodoHandler with the given store.
func NewTodoHandler(s store.TodoStore) *TodoHandler {
	return &TodoHandler{store: s}
}

// List handles GET /todos
// Returns all todos as a JSON array. Returns an empty array (not null) when
// the store is empty so clients do not need to guard against a null response.
func (h *TodoHandler) List(w http.ResponseWriter, r *http.Request) {
	todos := h.store.List()

	// Guarantee a JSON array, never null.
	if todos == nil {
		todos = []store.Todo{}
	}

	respond.JSON(w, http.StatusOK, todos)
}

// Get handles GET /todos/{id}
// Returns the todo with the given ID or 404 if not found.
func (h *TodoHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	todo, ok := h.store.Get(id)
	if !ok {
		respond.Error(w, http.StatusNotFound, "todo not found")
		return
	}

	respond.JSON(w, http.StatusOK, todo)
}

// createTodoRequest is the expected JSON body for POST /todos.
type createTodoRequest struct {
	Title string `json:"title"`
}

// Create handles POST /todos
// Expects a JSON body with a "title" field.
// Returns 201 Created with the new todo on success.
func (h *TodoHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createTodoRequest
	if err := decodeJSON(r, &req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	// Validate.
	errs := make(map[string]string)
	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		errs["title"] = "required"
	}
	if len(errs) > 0 {
		respond.ValidationError(w, errs)
		return
	}

	todo := h.store.Create(req.Title)
	respond.JSON(w, http.StatusCreated, todo)
}

// updateTodoRequest is the expected JSON body for PUT /todos/{id}.
type updateTodoRequest struct {
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

// Update handles PUT /todos/{id}
// Replaces the title and done flag of an existing todo.
// Returns 200 with the updated todo, or 404 if not found.
func (h *TodoHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req updateTodoRequest
	if err := decodeJSON(r, &req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	// Validate.
	errs := make(map[string]string)
	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		errs["title"] = "required"
	}
	if len(errs) > 0 {
		respond.ValidationError(w, errs)
		return
	}

	todo, ok := h.store.Update(id, req.Title, req.Done)
	if !ok {
		respond.Error(w, http.StatusNotFound, "todo not found")
		return
	}

	respond.JSON(w, http.StatusOK, todo)
}

// Delete handles DELETE /todos/{id}
// Returns 204 No Content on success, or 404 if not found.
func (h *TodoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if ok := h.store.Delete(id); !ok {
		respond.Error(w, http.StatusNotFound, "todo not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// decodeJSON decodes the request body into v, enforcing that the body is valid
// JSON and disallowing unknown fields. The request body is always closed after
// decoding regardless of error, satisfying the http.Handler contract.
func decodeJSON(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}
