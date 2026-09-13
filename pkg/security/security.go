package security

import (
	"strings"
	"sync"
)

// User represents an authenticated identity with assigned roles and permissions.
type User struct {
	ID          string
	Username    string
	Roles       []string
	Permissions []string
	Metadata    map[string]any
}

// HasRole checks if the user possesses the specified role (case-insensitive).
func (u *User) HasRole(role string) bool {
	if u == nil {
		return false
	}
	target := strings.ToLower(role)
	for _, r := range u.Roles {
		if strings.ToLower(r) == target {
			return true
		}
	}
	return false
}

// HasPermission checks if the user possesses the specified permission string.
func (u *User) HasPermission(perm string) bool {
	if u == nil {
		return false
	}
	for _, p := range u.Permissions {
		if p == "*" || p == perm {
			return true
		}
	}
	return false
}

// SecurityManager coordinates authentication state, sessions, and active credentials.
type SecurityManager struct {
	mu          sync.RWMutex
	currentUser *User
	token       string
}

var (
	defaultSecurityOnce sync.Once
	defaultSecurity     *SecurityManager
)

// Default returns the global SecurityManager instance.
func Default() *SecurityManager {
	defaultSecurityOnce.Do(func() {
		defaultSecurity = NewSecurityManager()
	})
	return defaultSecurity
}

// NewSecurityManager creates an unauthenticated security manager.
func NewSecurityManager() *SecurityManager {
	return &SecurityManager{}
}

// Login authenticates a user and optionally sets a security token.
func (sm *SecurityManager) Login(user *User, token string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.currentUser = user
	sm.token = token
}

// Logout clears the current authentication state.
func (sm *SecurityManager) Logout() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.currentUser = nil
	sm.token = ""
}

// CurrentUser returns the active user, or nil if unauthenticated.
func (sm *SecurityManager) CurrentUser() *User {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.currentUser
}

// IsAuthenticated returns true if a user is currently logged in.
func (sm *SecurityManager) IsAuthenticated() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.currentUser != nil
}

// Token returns the active session/bearer token.
func (sm *SecurityManager) Token() string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.token
}
