package auth

import (
	"net/http"
)

// RequireAnyRole returns a middleware that allows access if the user has
// at least one of the required roles.
func RequireAnyRole(requiredRoles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(requiredRoles))
	for _, r := range requiredRoles {
		allowed[r] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			roles := RolesFromContext(r.Context())
			if len(roles) == 0 {
				http.Error(w, "forbidden (no roles)", http.StatusForbidden)
				return
			}

			for _, ur := range roles {
				if _, ok := allowed[ur]; ok {
					next.ServeHTTP(w, r)
					return
				}
			}

			http.Error(w, "forbidden (insufficient role)", http.StatusForbidden)
		})
	}
}

// RequireSuperadmin is a convenience wrapper for RequireAnyRole("superadmin").
func RequireSuperadmin() func(http.Handler) http.Handler {
	return RequireAnyRole("superadmin")
}
