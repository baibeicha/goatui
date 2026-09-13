package router

import (
	"strings"
	"sync"
)

// HandlerFunc handles a matched route and typically instantiates a Screen.
type HandlerFunc func(ctx *RouteContext) any

// Middleware wraps a HandlerFunc to perform pre/post-processing (e.g. auth, logging).
type Middleware func(next HandlerFunc) HandlerFunc

// RedirectMsg requests navigation to another URL from within a handler or middleware.
type RedirectMsg struct {
	URL string
}

// Redirect returns a RedirectMsg for URL redirection.
func Redirect(url string) RedirectMsg {
	return RedirectMsg{URL: url}
}

type trieNode struct {
	part       string
	children   []*trieNode
	isParam    bool
	paramName  string
	isWildcard bool
	handler    HandlerFunc
	pattern    string
}

func (n *trieNode) insert(segments []string, pattern string, handler HandlerFunc) {
	if len(segments) == 0 {
		n.handler = handler
		n.pattern = pattern
		return
	}

	seg := segments[0]
	var child *trieNode

	for _, c := range n.children {
		if c.part == seg {
			child = c
			break
		}
	}

	if child == nil {
		isParam := strings.HasPrefix(seg, ":")
		paramName := ""
		if isParam {
			paramName = seg[1:]
		}
		isWildcard := (seg == "*")

		child = &trieNode{
			part:       seg,
			isParam:    isParam,
			paramName:  paramName,
			isWildcard: isWildcard,
		}
		n.children = append(n.children, child)
	}

	child.insert(segments[1:], pattern, handler)
}

func (n *trieNode) search(segments []string, params map[string]string) *trieNode {
	if len(segments) == 0 {
		if n.handler != nil {
			return n
		}
		// Check if there is a wildcard child
		for _, c := range n.children {
			if c.isWildcard && c.handler != nil {
				return c
			}
		}
		return nil
	}

	seg := segments[0]

	// 1. Try exact match first
	for _, c := range n.children {
		if !c.isParam && !c.isWildcard && c.part == seg {
			if matched := c.search(segments[1:], params); matched != nil {
				return matched
			}
		}
	}

	// 2. Try parameter match
	for _, c := range n.children {
		if c.isParam {
			params[c.paramName] = seg
			if matched := c.search(segments[1:], params); matched != nil {
				return matched
			}
			delete(params, c.paramName)
		}
	}

	// 3. Try wildcard match
	for _, c := range n.children {
		if c.isWildcard {
			params["*"] = strings.Join(segments, "/")
			return c
		}
	}

	return nil
}

// Router provides ultra-fast Trie-based URL route matching with path parameters, query strings, and middlewares.
type Router struct {
	mu          sync.RWMutex
	root        *trieNode
	middlewares []Middleware
	session     *SessionStore
	patterns    []string
}

// NewRouter creates a new URL router.
func NewRouter() *Router {
	return &Router{
		root:    &trieNode{},
		session: NewSessionStore(),
	}
}

// Session returns the router's session storage.
func (r *Router) Session() *SessionStore {
	return r.session
}

// Use registers global router middlewares.
func (r *Router) Use(middlewares ...Middleware) *Router {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.middlewares = append(r.middlewares, middlewares...)
	return r
}

// Handle registers a handler for the given URL pattern with optional route-specific middlewares.
func (r *Router) Handle(pattern string, handler HandlerFunc, middlewares ...Middleware) *Router {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Chain route-specific middlewares and global middlewares
	finalHandler := handler
	for i := len(middlewares) - 1; i >= 0; i-- {
		finalHandler = middlewares[i](finalHandler)
	}
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		finalHandler = r.middlewares[i](finalHandler)
	}

	segments := splitPath(pattern)
	r.root.insert(segments, pattern, finalHandler)
	r.patterns = append(r.patterns, pattern)
	return r
}

// Match resolves a URL path and query string to its registered handler and extracted context.
func (r *Router) Match(rawURL string) (HandlerFunc, *RouteContext, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	path, query := ParseURL(rawURL)
	segments := splitPath(path)

	ctx := NewRouteContext(path, r.session)
	ctx.Query = query

	node := r.root.search(segments, ctx.Params)
	if node == nil || node.handler == nil {
		return nil, ctx, false
	}

	ctx.Pattern = node.pattern
	return node.handler, ctx, true
}

// Routes returns all registered route patterns in definition order.
func (r *Router) Routes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	res := make([]string, len(r.patterns))
	copy(res, r.patterns)
	return res
}

func splitPath(path string) []string {
	if schemeIdx := strings.Index(path, "://"); schemeIdx != -1 {
		scheme := path[:schemeIdx+3]
		rest := path[schemeIdx+3:]
		rest = strings.Trim(rest, "/")
		parts := []string{scheme}
		if rest != "" {
			parts = append(parts, strings.Split(rest, "/")...)
		}
		return parts
	}

	path = strings.Trim(path, "/")
	if path == "" {
		return nil
	}
	return strings.Split(path, "/")
}
