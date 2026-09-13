package router

import (
	"testing"
)

func TestRouterMatching(t *testing.T) {
	r := NewRouter()

	r.Handle("/dashboard", func(ctx *RouteContext) any {
		return "dashboard"
	})

	r.Handle("/users/:id", func(ctx *RouteContext) any {
		return "user-" + ctx.Param("id")
	})

	r.Handle("/processes/:pid/threads/:tid", func(ctx *RouteContext) any {
		return "proc-" + ctx.Param("pid") + "-thread-" + ctx.Param("tid")
	})

	r.Handle("/admin/*", func(ctx *RouteContext) any {
		return "admin"
	})

	// Test static match
	handler, ctx, ok := r.Match("/dashboard")
	if !ok || handler(ctx) != "dashboard" {
		t.Fatalf("Failed to match /dashboard")
	}

	// Test param match with query string
	handler, ctx, ok = r.Match("/users/123?tab=profile&active=true")
	if !ok {
		t.Fatalf("Failed to match /users/123")
	}
	if ctx.Param("id") != "123" {
		t.Errorf("Expected id=123, got %q", ctx.Param("id"))
	}
	if ctx.QueryParam("tab") != "profile" || ctx.QueryParam("active") != "true" {
		t.Errorf("Query params mismatch: %+v", ctx.Query)
	}
	if res := handler(ctx); res != "user-123" {
		t.Errorf("Handler return mismatch: got %v", res)
	}

	// Test multi-param match
	handler, ctx, ok = r.Match("/processes/999/threads/42")
	if !ok {
		t.Fatalf("Failed to match multi-param")
	}
	if ctx.Param("pid") != "999" || ctx.Param("tid") != "42" {
		t.Errorf("Param mismatch: pid=%s tid=%s", ctx.Param("pid"), ctx.Param("tid"))
	}

	// Test wildcard match
	handler, ctx, ok = r.Match("/admin/secret/delete")
	if !ok || handler(ctx) != "admin" {
		t.Fatalf("Failed to match wildcard /admin/*")
	}

	// Test URL scheme routes (file://* and http://*)
	r.Handle("file://*", func(ctx *RouteContext) any {
		return "file-viewer"
	})
	r.Handle("http://*", func(ctx *RouteContext) any {
		return "http-viewer"
	})

	handler, ctx, ok = r.Match("file:///home/user/document.txt")
	if !ok || handler(ctx) != "file-viewer" {
		t.Fatalf("Failed to match file://* route, ok=%v", ok)
	}

	handler, ctx, ok = r.Match("http://api.example.com/v1/health")
	if !ok || handler(ctx) != "http-viewer" {
		t.Fatalf("Failed to match http://* route, ok=%v", ok)
	}

	// Test 404
	_, _, ok = r.Match("/nonexistent")
	if ok {
		t.Errorf("Expected false for non-existent route")
	}
}

func TestRouterMiddleware(t *testing.T) {
	r := NewRouter()
	calls := []string{}

	globalMw := func(next HandlerFunc) HandlerFunc {
		return func(ctx *RouteContext) any {
			calls = append(calls, "global_before")
			res := next(ctx)
			calls = append(calls, "global_after")
			return res
		}
	}

	routeMw := func(next HandlerFunc) HandlerFunc {
		return func(ctx *RouteContext) any {
			calls = append(calls, "route_before")
			res := next(ctx)
			calls = append(calls, "route_after")
			return res
		}
	}

	r.Use(globalMw)
	r.Handle("/protected", func(ctx *RouteContext) any {
		calls = append(calls, "handler")
		return "ok"
	}, routeMw)

	handler, ctx, ok := r.Match("/protected")
	if !ok {
		t.Fatalf("Route not matched")
	}
	res := handler(ctx)
	if res != "ok" {
		t.Errorf("Expected 'ok', got %v", res)
	}

	expected := []string{"global_before", "route_before", "handler", "route_after", "global_after"}
	if len(calls) != len(expected) {
		t.Fatalf("Middleware sequence mismatch: got %v, want %v", calls, expected)
	}
	for i := range expected {
		if calls[i] != expected[i] {
			t.Errorf("Step %d: got %s, want %s", i, calls[i], expected[i])
		}
	}
}

func TestSessionStore(t *testing.T) {
	s := NewSessionStore()
	s.Set("user", "alice")
	s.Set("counter", 42)
	s.Set("auth", true)

	if s.GetString("user", "") != "alice" {
		t.Errorf("Expected alice")
	}
	if s.GetInt("counter", 0) != 42 {
		t.Errorf("Expected 42")
	}
	if !s.GetBool("auth", false) {
		t.Errorf("Expected true")
	}

	s.Delete("user")
	if s.GetString("user", "none") != "none" {
		t.Errorf("Expected fallback after delete")
	}
}

func BenchmarkRouterMatch(b *testing.B) {
	r := NewRouter()
	r.Handle("/api/v1/users/:id/posts/:post_id/comments", func(ctx *RouteContext) any {
		return "matched"
	})
	path := "/api/v1/users/42/posts/100/comments?sort=desc"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Match(path)
	}
}
