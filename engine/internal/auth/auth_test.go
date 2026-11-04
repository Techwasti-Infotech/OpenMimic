package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTokenValidatorRBAC(t *testing.T) {
	validator := NewTokenValidator()

	// Handler requiring super_admin
	superHandler := validator.Middleware(RoleSuperAdmin)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok_super"))
	}))

	// 1. Missing header -> 401
	reqMissing := httptest.NewRequest("GET", "/api/admin", nil)
	rrMissing := httptest.NewRecorder()
	superHandler.ServeHTTP(rrMissing, reqMissing)
	if rrMissing.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized for missing header, got %d", rrMissing.Code)
	}

	// 2. Tenant admin key accessing super_admin endpoint -> 403 Forbidden
	reqTenant := httptest.NewRequest("GET", "/api/admin", nil)
	reqTenant.Header.Set("Authorization", "Bearer key_tenant_alpha")
	rrTenant := httptest.NewRecorder()
	superHandler.ServeHTTP(rrTenant, reqTenant)
	if rrTenant.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for tenant accessing super_admin, got %d", rrTenant.Code)
	}

	// 3. Super admin key accessing super_admin endpoint -> 200 OK
	reqSuper := httptest.NewRequest("GET", "/api/admin", nil)
	reqSuper.Header.Set("Authorization", "Bearer key_super_secret")
	rrSuper := httptest.NewRecorder()
	superHandler.ServeHTTP(rrSuper, reqSuper)
	if rrSuper.Code != http.StatusOK {
		t.Errorf("expected 200 OK for valid super_admin, got %d", rrSuper.Code)
	}
}

func TestTokenValidatorTenantAdminAccess(t *testing.T) {
	validator := NewTokenValidator()

	tenantHandler := validator.Middleware(RoleTenantAdmin)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := GetClaims(r.Context())
		if !ok || claims == nil {
			http.Error(w, "missing claims", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(claims.TenantID))
	}))

	// Viewer accessing tenant_admin route -> 403
	reqViewer := httptest.NewRequest("POST", "/api/generator", nil)
	reqViewer.Header.Set("X-API-Key", "key_viewer_demo")
	rrViewer := httptest.NewRecorder()
	tenantHandler.ServeHTTP(rrViewer, reqViewer)
	if rrViewer.Code != http.StatusForbidden {
		t.Errorf("expected 403 for viewer accessing tenant admin, got %d", rrViewer.Code)
	}

	// Tenant admin accessing tenant_admin route -> 200
	reqAdmin := httptest.NewRequest("POST", "/api/generator", nil)
	reqAdmin.Header.Set("X-API-Key", "key_tenant_alpha")
	rrAdmin := httptest.NewRecorder()
	tenantHandler.ServeHTTP(rrAdmin, reqAdmin)
	if rrAdmin.Code != http.StatusOK || rrAdmin.Body.String() != "tenant_alpha" {
		t.Errorf("expected 200 OK with body tenant_alpha, got %d body=%s", rrAdmin.Code, rrAdmin.Body.String())
	}
}

