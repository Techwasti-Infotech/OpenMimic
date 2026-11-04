package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Role represents an authorization tier in OpenMimic.
type Role string

const (
	RoleSuperAdmin Role = "super_admin"
	RoleTenantAdmin Role = "tenant_admin"
	RoleViewer     Role = "viewer"
)

// Claims represents authenticated principal information.
type Claims struct {
	Subject  string `json:"sub"`
	TenantID string `json:"tenant_id"`
	Role     Role   `json:"role"`
	Expires  int64  `json:"exp"`
}

type contextKey string

const ClaimsContextKey contextKey = "openmimic_auth_claims"

// TokenValidator checks JWT or API key strings and returns authenticated claims.
type TokenValidator struct {
	apiKeys map[string]Claims
}

// NewTokenValidator creates a validator with predefined API keys or mock JWTs.
func NewTokenValidator() *TokenValidator {
	return &TokenValidator{
		apiKeys: map[string]Claims{
			"key_super_secret": {
				Subject:  "super-admin-user",
				TenantID: "*",
				Role:     RoleSuperAdmin,
				Expires:  time.Now().Add(365 * 24 * time.Hour).Unix(),
			},
			"key_tenant_alpha": {
				Subject:  "alpha-admin",
				TenantID: "tenant_alpha",
				Role:     RoleTenantAdmin,
				Expires:  time.Now().Add(365 * 24 * time.Hour).Unix(),
			},
			"key_viewer_demo": {
				Subject:  "demo-viewer",
				TenantID: "tenant_alpha",
				Role:     RoleViewer,
				Expires:  time.Now().Add(365 * 24 * time.Hour).Unix(),
			},
		},
	}
}

// RegisterKey adds a token/api-key pair.
func (v *TokenValidator) RegisterKey(key string, claims Claims) {
	v.apiKeys[key] = claims
}

// ValidateToken verifies an Authorization header token or API key.
func (v *TokenValidator) ValidateToken(tokenStr string) (*Claims, error) {
	clean := strings.TrimPrefix(tokenStr, "Bearer ")
	clean = strings.TrimSpace(clean)

	claims, found := v.apiKeys[clean]
	if !found {
		return nil, fmt.Errorf("invalid or unrecognized authorization token")
	}

	if claims.Expires > 0 && time.Now().Unix() > claims.Expires {
		return nil, fmt.Errorf("authorization token has expired")
	}

	return &claims, nil
}

// Middleware creates an HTTP authentication & RBAC middleware.
func (v *TokenValidator) Middleware(requiredRole Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				authHeader = r.Header.Get("X-API-Key")
			}

			if authHeader == "" {
				http.Error(w, `{"error":"unauthorized: missing Authorization or X-API-Key header"}`, http.StatusUnauthorized)
				return
			}

			claims, err := v.ValidateToken(authHeader)
			if err != nil {
				http.Error(w, fmt.Sprintf(`{"error":"unauthorized: %s"}`, err.Error()), http.StatusUnauthorized)
				return
			}

			// RBAC hierarchy check: SuperAdmin can do anything; TenantAdmin can view/edit; Viewer is read-only
			if requiredRole == RoleSuperAdmin && claims.Role != RoleSuperAdmin {
				http.Error(w, `{"error":"forbidden: super_admin role required"}`, http.StatusForbidden)
				return
			}
			if requiredRole == RoleTenantAdmin && (claims.Role != RoleSuperAdmin && claims.Role != RoleTenantAdmin) {
				http.Error(w, `{"error":"forbidden: tenant_admin role required"}`, http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetClaims retrieves Claims from request context.
func GetClaims(ctx context.Context) (*Claims, bool) {
	c, ok := ctx.Value(ClaimsContextKey).(*Claims)
	return c, ok
}

