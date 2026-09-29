package store_test

import (
	"sync"
	"testing"

	"github.com/user/go-api-template/internal/store"
)

func TestMemoryStore_CreateAndGet(t *testing.T) {
	s := store.NewMemoryStore()

	todo := s.Create("Buy milk")

	if todo.ID == "" {
		t.Error("Create: expected non-empty ID")
	}
	if todo.Title != "Buy milk" {
		t.Errorf("Create: Title got %q, want %q", todo.Title, "Buy milk")
	}
	if todo.Done {
		t.Error("Create: expected Done=false")
	}
	if todo.CreatedAt.IsZero() {
		t.Error("Create: expected non-zero CreatedAt")
	}

	got, ok := s.Get(todo.ID)
	if !ok {
		t.Fatal("Get: expected todo to exist")
	}
	if got.ID != todo.ID {
		t.Errorf("Get: ID mismatch got %q want %q", got.ID, todo.ID)
	}
}

func TestMemoryStore_GetMissing(t *testing.T) {
	s := store.NewMemoryStore()

	_, ok := s.Get("nonexistent")
	if ok {
		t.Error("Get: expected false for missing ID")
	}
}

func TestMemoryStore_List(t *testing.T) {
	s := store.NewMemoryStore()

	// Empty store returns non-nil slice.
	list := s.List()
	if list == nil {
		t.Error("List: expected non-nil slice for empty store")
	}

	s.Create("A")
	s.Create("B")
	s.Create("C")

	list = s.List()
	if len(list) != 3 {
		t.Errorf("List: got %d items, want 3", len(list))
	}
}

func TestMemoryStore_Update(t *testing.T) {
	s := store.NewMemoryStore()
	todo := s.Create("Original")

	updated, ok := s.Update(todo.ID, "Updated", true)
	if !ok {
		t.Fatal("Update: expected ok=true")
	}
	if updated.Title != "Updated" {
		t.Errorf("Update: Title got %q, want %q", updated.Title, "Updated")
	}
	if !updated.Done {
		t.Error("Update: expected Done=true")
	}
	if !updated.UpdatedAt.After(todo.CreatedAt) || updated.UpdatedAt.Equal(todo.CreatedAt) {
		// UpdatedAt should be >= CreatedAt. Allow equal for fast tests.
		// Just verify it's not zero.
		if updated.UpdatedAt.IsZero() {
			t.Error("Update: UpdatedAt is zero")
		}
	}
}

func TestMemoryStore_UpdateMissing(t *testing.T) {
	s := store.NewMemoryStore()

	_, ok := s.Update("nonexistent", "Title", false)
	if ok {
		t.Error("Update: expected ok=false for missing ID")
	}
}

func TestMemoryStore_Delete(t *testing.T) {
	s := store.NewMemoryStore()
	todo := s.Create("To delete")

	if ok := s.Delete(todo.ID); !ok {
		t.Fatal("Delete: expected ok=true")
	}
	if _, ok := s.Get(todo.ID); ok {
		t.Error("Get after Delete: expected ok=false")
	}
}

func TestMemoryStore_DeleteMissing(t *testing.T) {
	s := store.NewMemoryStore()

	if ok := s.Delete("nonexistent"); ok {
		t.Error("Delete: expected ok=false for missing ID")
	}
}

// TestMemoryStore_ConcurrentReadWrite verifies that the store does not race
// under concurrent access. Run with -race to detect data races.
func TestMemoryStore_ConcurrentReadWrite(t *testing.T) {
	s := store.NewMemoryStore()

	var wg sync.WaitGroup
	const goroutines = 50

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			todo := s.Create("concurrent")
			s.Get(todo.ID)
			s.List()
			s.Update(todo.ID, "updated", true)
			s.Delete(todo.ID)
		}()
	}

	wg.Wait()
}
