package security_test

import (
	"testing"

	"github.com/baibeicha/goatui/pkg/router"
	"github.com/baibeicha/goatui/pkg/security"
	"github.com/baibeicha/goatui/pkg/window"
)

type dummyScreen struct {
	window.BaseScreen
	name string
}

func TestUserRolesAndPermissions(t *testing.T) {
	u := &security.User{
		ID:          "usr_123",
		Username:    "alice",
		Roles:       []string{"Admin", "Developer"},
		Permissions: []string{"read:users", "write:*"},
	}

	if !u.HasRole("admin") || !u.HasRole("ADMIN") || !u.HasRole("developer") {
		t.Errorf("expected role matches, got false")
	}
	if u.HasRole("manager") {
		t.Errorf("expected manager role to be false")
	}

	if !u.HasPermission("read:users") {
		t.Errorf("expected permission read:users to match")
	}
	if u.HasPermission("delete:all") {
		t.Errorf("expected unlisted permission to not match")
	}

	superUser := &security.User{
		Permissions: []string{"*"},
	}
	if !superUser.HasPermission("anything") {
		t.Errorf("wildcard permission should match anything")
	}

	var nilUser *security.User
	if nilUser.HasRole("admin") || nilUser.HasPermission("read") {
		t.Errorf("nil user should have no roles or permissions")
	}
}

func TestSecurityManagerAuth(t *testing.T) {
	sm := security.NewSecurityManager()

	if sm.IsAuthenticated() {
		t.Fatalf("expected unauthenticated manager")
	}
	if sm.CurrentUser() != nil {
		t.Fatalf("expected nil current user")
	}

	alice := &security.User{ID: "1", Username: "alice", Roles: []string{"admin"}}
	sm.Login(alice, "secret-token-xyz")

	if !sm.IsAuthenticated() {
		t.Fatalf("expected authenticated state")
	}
	if sm.CurrentUser().Username != "alice" {
		t.Fatalf("expected user alice, got %v", sm.CurrentUser())
	}
	if sm.Token() != "secret-token-xyz" {
		t.Fatalf("expected secret-token-xyz, got %s", sm.Token())
	}

	sm.Logout()
	if sm.IsAuthenticated() {
		t.Fatalf("expected unauthenticated state after logout")
	}
	if sm.CurrentUser() != nil {
		t.Fatalf("expected nil user after logout")
	}
	if sm.Token() != "" {
		t.Fatalf("expected empty token after logout")
	}
}

func TestRouteGuards(t *testing.T) {
	sm := security.NewSecurityManager()
	r := router.NewRouter()

	r.Handle("/public", func(ctx *router.RouteContext) any {
		return &dummyScreen{name: "public"}
	})

	r.Handle("/dashboard", func(ctx *router.RouteContext) any {
		return &dummyScreen{name: "dashboard"}
	}, security.RequireAuth("/login", sm))

	r.Handle("/admin", func(ctx *router.RouteContext) any {
		return &dummyScreen{name: "admin"}
	}, security.RequireRole("admin", "", sm))

	r.Handle("/billing", func(ctx *router.RouteContext) any {
		return &dummyScreen{name: "billing"}
	}, security.RequirePermission("billing:write", "/dashboard", sm))

	// 1. Unauthenticated test on /public
	h, ctx, ok := r.Match("/public")
	if !ok {
		t.Fatalf("expected match for /public")
	}
	res := h(ctx)
	if scr, ok := res.(*dummyScreen); !ok || scr.name != "public" {
		t.Fatalf("expected public screen")
	}

	// 2. Unauthenticated test on /dashboard (should redirect to /login)
	h, ctx, ok = r.Match("/dashboard")
	if !ok {
		t.Fatalf("expected match for /dashboard")
	}
	res = h(ctx)
	red, isRed := res.(router.RedirectMsg)
	if !isRed || red.URL != "/login" {
		t.Fatalf("expected redirect to /login, got %#v", res)
	}

	// 3. Unauthenticated test on /admin (should return AccessDeniedScreen)
	h, ctx, ok = r.Match("/admin")
	if !ok {
		t.Fatalf("expected match for /admin")
	}
	res = h(ctx)
	if _, isDenied := res.(*window.AccessDeniedScreen); !isDenied {
		t.Fatalf("expected AccessDeniedScreen, got %#v", res)
	}

	// 4. Log in as regular user
	sm.Login(&security.User{Username: "bob", Roles: []string{"user"}, Permissions: []string{"read"}}, "tok")

	// 4a. /dashboard now accessible
	h, ctx, _ = r.Match("/dashboard")
	res = h(ctx)
	if scr, ok := res.(*dummyScreen); !ok || scr.name != "dashboard" {
		t.Fatalf("expected dashboard screen for logged in user")
	}

	// 4b. /admin still denied
	h, ctx, _ = r.Match("/admin")
	res = h(ctx)
	if _, isDenied := res.(*window.AccessDeniedScreen); !isDenied {
		t.Fatalf("expected AccessDeniedScreen for bob on /admin, got %#v", res)
	}

	// 4c. /billing redirected to /dashboard
	h, ctx, _ = r.Match("/billing")
	res = h(ctx)
	if red, ok := res.(router.RedirectMsg); !ok || red.URL != "/dashboard" {
		t.Fatalf("expected redirect to /dashboard for /billing, got %#v", res)
	}

	// 5. Elevate bob to admin with billing:write
	sm.Login(&security.User{Username: "bob", Roles: []string{"admin"}, Permissions: []string{"billing:write"}}, "tok")

	h, ctx, _ = r.Match("/admin")
	res = h(ctx)
	if scr, ok := res.(*dummyScreen); !ok || scr.name != "admin" {
		t.Fatalf("expected admin screen for bob")
	}

	h, ctx, _ = r.Match("/billing")
	res = h(ctx)
	if scr, ok := res.(*dummyScreen); !ok || scr.name != "billing" {
		t.Fatalf("expected billing screen for bob")
	}
}
