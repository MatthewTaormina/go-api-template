// Package handlers contains the HTTP handler functions for the API.
// Each handler follows the same pattern:
//  1. Parse input (path params, query, body).
//  2. Validate inputs and return 400/422 on failure.
//  3. Call the store.
//  4. Write JSON response.
//
// Handlers depend only on the store.TodoStore interface, not on MemoryStore,
// making them trivially testable with a mock.
package handlers
