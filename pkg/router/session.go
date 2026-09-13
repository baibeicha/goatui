package router

import (
	"sync"
)

// SessionStore provides thread-safe in-memory key-value storage across screen navigations.
type SessionStore struct {
	mu   sync.RWMutex
	data map[string]any
}

// NewSessionStore creates an empty session storage.
func NewSessionStore() *SessionStore {
	return &SessionStore{
		data: make(map[string]any),
	}
}

// Set stores a value under the specified key.
func (s *SessionStore) Set(key string, val any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = val
}

// Get retrieves a raw value by key.
func (s *SessionStore) Get(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[key]
	return v, ok
}

// GetString retrieves a string value or fallback if not found or wrong type.
func (s *SessionStore) GetString(key string, fallback string) string {
	if v, ok := s.Get(key); ok {
		if str, ok := v.(string); ok {
			return str
		}
	}
	return fallback
}

// GetInt retrieves an integer value or fallback if not found or wrong type.
func (s *SessionStore) GetInt(key string, fallback int) int {
	if v, ok := s.Get(key); ok {
		if i, ok := v.(int); ok {
			return i
		}
	}
	return fallback
}

// GetBool retrieves a boolean value or fallback if not found or wrong type.
func (s *SessionStore) GetBool(key string, fallback bool) bool {
	if v, ok := s.Get(key); ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return fallback
}

// Delete removes a key from the session.
func (s *SessionStore) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
}

// Clear removes all stored session values.
func (s *SessionStore) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = make(map[string]any)
}
