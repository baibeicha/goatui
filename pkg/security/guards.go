package security

import (
	"fmt"

	"github.com/baibeicha/goatui/pkg/router"
	"github.com/baibeicha/goatui/pkg/window"
)

func resolveSecurityManager(smOpt []*SecurityManager) *SecurityManager {
	if len(smOpt) > 0 && smOpt[0] != nil {
		return smOpt[0]
	}
	return Default()
}

// RequireAuth enforces that a user must be logged in before viewing the route.
// If unauthenticated, it either redirects to fallbackURL (if non-empty) or renders an AccessDeniedScreen.
func RequireAuth(fallbackURL string, smOpt ...*SecurityManager) router.Middleware {
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(ctx *router.RouteContext) any {
			sm := resolveSecurityManager(smOpt)
			if !sm.IsAuthenticated() {
				if fallbackURL != "" {
					return router.Redirect(fallbackURL)
				}
				return window.NewAccessDeniedScreen("Authentication required to access this resource", ctx.Path)
			}
			return next(ctx)
		}
	}
}

// RequireRole enforces that the active user possesses the specified role.
// If unauthorized, it either redirects to fallbackURL (if non-empty) or renders an AccessDeniedScreen.
func RequireRole(role string, fallbackURL string, smOpt ...*SecurityManager) router.Middleware {
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(ctx *router.RouteContext) any {
			sm := resolveSecurityManager(smOpt)
			user := sm.CurrentUser()
			if user == nil || !user.HasRole(role) {
				if fallbackURL != "" {
					return router.Redirect(fallbackURL)
				}
				reason := fmt.Sprintf("Role '%s' is required to access this screen", role)
				return window.NewAccessDeniedScreen(reason, ctx.Path)
			}
			return next(ctx)
		}
	}
}

// RequirePermission enforces that the active user possesses the specified permission string.
// If unauthorized, it either redirects to fallbackURL (if non-empty) or renders an AccessDeniedScreen.
func RequirePermission(perm string, fallbackURL string, smOpt ...*SecurityManager) router.Middleware {
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(ctx *router.RouteContext) any {
			sm := resolveSecurityManager(smOpt)
			user := sm.CurrentUser()
			if user == nil || !user.HasPermission(perm) {
				if fallbackURL != "" {
					return router.Redirect(fallbackURL)
				}
				reason := fmt.Sprintf("Permission '%s' is required to access this screen", perm)
				return window.NewAccessDeniedScreen(reason, ctx.Path)
			}
			return next(ctx)
		}
	}
}

// Require enforces a custom authorization predicate.
func Require(predicate func(user *User, ctx *router.RouteContext) bool, fallbackURL, reason string, smOpt ...*SecurityManager) router.Middleware {
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(ctx *router.RouteContext) any {
			sm := resolveSecurityManager(smOpt)
			user := sm.CurrentUser()
			if !predicate(user, ctx) {
				if fallbackURL != "" {
					return router.Redirect(fallbackURL)
				}
				if reason == "" {
					reason = "Access to this resource was denied"
				}
				return window.NewAccessDeniedScreen(reason, ctx.Path)
			}
			return next(ctx)
		}
	}
}
