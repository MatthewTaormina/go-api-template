// Package store defines the data store interface and an in-memory implementation.
// Swap the MemoryStore for a real database implementation without changing any handler code.
package store

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Todo is the domain model. Tags are used for JSON serialisation in handlers.
type Todo struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TodoStore defines the operations any backing store must support.
// Handlers depend on this interface, not on MemoryStore directly.
type TodoStore interface {
	List() []Todo
	Get(id string) (Todo, bool)
	Create(title string) Todo
	Update(id, title string, done bool) (Todo, bool)
	Delete(id string) bool
}

// MemoryStore is a thread-safe, in-memory TodoStore backed by a map.
// It is suitable for local development and testing.
// Replace with a database-backed implementation for production use.
type MemoryStore struct {
	mu     sync.RWMutex
	todos  map[string]Todo
	nextID atomic.Uint64
}

// NewMemoryStore returns an initialised, empty MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		todos: make(map[string]Todo),
	}
}

// List returns all todos in insertion order (deterministic for tests).
func (s *MemoryStore) List() []Todo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]Todo, 0, len(s.todos))
	for _, t := range s.todos {
		out = append(out, t)
	}
	return out
}

// Get returns the todo with the given ID and a boolean indicating whether it was found.
func (s *MemoryStore) Get(id string) (Todo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	t, ok := s.todos[id]
	return t, ok
}

// Create inserts a new todo with an auto-generated ID and returns it.
func (s *MemoryStore) Create(title string) Todo {
	id := fmt.Sprintf("%d", s.nextID.Add(1))
	now := time.Now().UTC()

	t := Todo{
		ID:        id,
		Title:     title,
		Done:      false,
		CreatedAt: now,
		UpdatedAt: now,
	}

	s.mu.Lock()
	s.todos[id] = t
	s.mu.Unlock()

	return t
}

// Update replaces the title and done flag of an existing todo.
// Returns the updated todo and true if found; zero value and false otherwise.
func (s *MemoryStore) Update(id, title string, done bool) (Todo, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, ok := s.todos[id]
	if !ok {
		return Todo{}, false
	}

	t.Title = title
	t.Done = done
	t.UpdatedAt = time.Now().UTC()
	s.todos[id] = t

	return t, true
}

// Delete removes the todo with the given ID.
// Returns true if the todo existed and was deleted; false if not found.
func (s *MemoryStore) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.todos[id]; !ok {
		return false
	}
	delete(s.todos, id)
	return true
}
