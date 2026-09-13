package router

import (
	"net/url"
	"strings"
)

// RouteContext encapsulates the URL path, extracted parameters, query string values, and session data.
type RouteContext struct {
	Path     string
	Pattern  string
	Params   map[string]string
	Query    map[string]string
	Session  *SessionStore
	UserData map[string]any
}

// NewRouteContext creates a new route context.
func NewRouteContext(path string, session *SessionStore) *RouteContext {
	return &RouteContext{
		Path:     path,
		Params:   make(map[string]string),
		Query:    make(map[string]string),
		Session:  session,
		UserData: make(map[string]any),
	}
}

// Param returns a path parameter value by key, or empty string if not found.
func (c *RouteContext) Param(key string) string {
	if c.Params == nil {
		return ""
	}
	return c.Params[key]
}

// QueryParam returns a query parameter value by key, or empty string if not found.
func (c *RouteContext) QueryParam(key string) string {
	if c.Query == nil {
		return ""
	}
	return c.Query[key]
}

// HasParam returns true if the specified path parameter exists.
func (c *RouteContext) HasParam(key string) bool {
	if c.Params == nil {
		return false
	}
	_, ok := c.Params[key]
	return ok
}

// HasQuery returns true if the specified query parameter exists.
func (c *RouteContext) HasQuery(key string) bool {
	if c.Query == nil {
		return false
	}
	_, ok := c.Query[key]
	return ok
}

// ParseURL decomposes a URL into path and query parameter map.
func ParseURL(rawURL string) (string, map[string]string) {
	if idx := strings.IndexByte(rawURL, '#'); idx != -1 {
		rawURL = rawURL[:idx]
	}
	parts := strings.SplitN(rawURL, "?", 2)
	path := parts[0]
	if path == "" {
		path = "/"
	}
	hasScheme := strings.Contains(path, "://")
	if !hasScheme && !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	queryParams := make(map[string]string)
	if len(parts) > 1 && parts[1] != "" {
		queryStr := parts[1]
		pairs := strings.Split(queryStr, "&")
		for _, pair := range pairs {
			if pair == "" {
				continue
			}
			kv := strings.SplitN(pair, "=", 2)
			if len(kv) == 2 {
				if unescaped, err := url.QueryUnescape(kv[1]); err == nil {
					queryParams[kv[0]] = unescaped
				} else {
					queryParams[kv[0]] = kv[1]
				}
			} else {
				queryParams[kv[0]] = "true"
			}
		}
	}

	return path, queryParams
}
