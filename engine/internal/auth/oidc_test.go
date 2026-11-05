package auth

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	pb "openmimic/engine/pkg/proto/v1"
)

func TestEnterpriseOIDCValidationAndRoleMapping(t *testing.T) {
	// 1. Generate RSA Key Pair for mock IdP
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	kid := "mock-okta-key-1"
	pubKey := &privKey.PublicKey

	// 2. Mock OIDC discovery and JWKS server
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"issuer":"http://mock-idp","jwks_uri":"http://%s/jwks"}`, r.Host)
	})
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		nStr := base64.RawURLEncoding.EncodeToString(pubKey.N.Bytes())
		eBytes := []byte{byte(pubKey.E >> 16), byte(pubKey.E >> 8), byte(pubKey.E)}
		if pubKey.E == 65537 {
			eBytes = []byte{0x01, 0x00, 0x01}
		}
		eStr := base64.RawURLEncoding.EncodeToString(eBytes)
		fmt.Fprintf(w, `{"keys":[{"kty":"RSA","kid":"%s","use":"sig","alg":"RS256","n":"%s","e":"%s"}]}`, kid, nStr, eStr)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	// 3. Configure Enterprise SSO settings
	ssoCfg := &pb.EnterpriseSSOConfig{
		Enabled:      true,
		ProviderType: "OKTA",
		IssuerUrl:    server.URL,
		ClientId:     "openmimic-enterprise-client",
		GroupRoleMappings: map[string]string{
			"DataPlatform-Admins": "super_admin",
			"Data-Engineers":      "tenant_admin",
		},
		DefaultRole: "viewer",
	}

	validator := NewEnterpriseOIDCValidator(ssoCfg, server.Client())
	if err := validator.RefreshKeys(); err != nil {
		t.Fatalf("failed to refresh keys from JWKS: %v", err)
	}

	// 4. Generate Valid Signed JWT for an engineer in "Data-Engineers"
	tokenStr := createSignedJWT(t, privKey, kid, "user-01@corp.internal", "openmimic-enterprise-client", []string{"Data-Engineers"}, time.Now().Add(1*time.Hour).Unix())

	claims, err := validator.ValidateJWT(tokenStr)
	if err != nil {
		t.Fatalf("expected successful JWT validation, got: %v", err)
	}

	if claims.Role != RoleTenantAdmin {
		t.Errorf("expected mapped role %s, got %s", RoleTenantAdmin, claims.Role)
	}
	if claims.Subject != "user-01@corp.internal" {
		t.Errorf("expected subject user-01@corp.internal, got %s", claims.Subject)
	}

	// 5. Test Super Admin Group mapping
	adminToken := createSignedJWT(t, privKey, kid, "admin@corp.internal", "openmimic-enterprise-client", []string{"DataPlatform-Admins"}, time.Now().Add(1*time.Hour).Unix())
	adminClaims, err := validator.ValidateJWT(adminToken)
	if err != nil {
		t.Fatalf("expected admin token validation success: %v", err)
	}
	if adminClaims.Role != RoleSuperAdmin {
		t.Errorf("expected role %s, got %s", RoleSuperAdmin, adminClaims.Role)
	}

	// 6. Test Default Fallback Role for unmapped group
	viewerToken := createSignedJWT(t, privKey, kid, "intern@corp.internal", "openmimic-enterprise-client", []string{"Marketing-Guests"}, time.Now().Add(1*time.Hour).Unix())
	viewerClaims, err := validator.ValidateJWT(viewerToken)
	if err != nil {
		t.Fatalf("expected viewer token validation success: %v", err)
	}
	if viewerClaims.Role != RoleViewer {
		t.Errorf("expected fallback role %s, got %s", RoleViewer, viewerClaims.Role)
	}

	// 7. Test Expired JWT rejection
	expiredToken := createSignedJWT(t, privKey, kid, "expired@corp.internal", "openmimic-enterprise-client", []string{"Data-Engineers"}, time.Now().Add(-2*time.Hour).Unix())
	_, err = validator.ValidateJWT(expiredToken)
	if err == nil {
		t.Errorf("expected error for expired token, got nil")
	}

	// 8. Test Client ID / Audience Mismatch rejection
	wrongAudToken := createSignedJWT(t, privKey, kid, "wrong-aud@corp.internal", "another-client-id", []string{"Data-Engineers"}, time.Now().Add(1*time.Hour).Unix())
	_, err = validator.ValidateJWT(wrongAudToken)
	if err == nil {
		t.Errorf("expected error for audience mismatch, got nil")
	}
}

func createSignedJWT(t *testing.T, privKey *rsa.PrivateKey, kid, sub, aud string, groups []string, exp int64) string {
	header := map[string]string{
		"alg": "RS256",
		"typ": "JWT",
		"kid": kid,
	}
	headerBytes, _ := json.Marshal(header)
	hB64 := base64.RawURLEncoding.EncodeToString(headerBytes)

	payload := map[string]any{
		"sub":    sub,
		"iss":    "http://mock-idp",
		"aud":    aud,
		"groups": groups,
		"exp":    exp,
		"iat":    time.Now().Unix(),
	}
	payloadBytes, _ := json.Marshal(payload)
	pB64 := base64.RawURLEncoding.EncodeToString(payloadBytes)

	signedContent := hB64 + "." + pB64
	hasher := sha256.New()
	hasher.Write([]byte(signedContent))
	hashed := hasher.Sum(nil)

	sigBytes, err := rsa.SignPKCS1v15(rand.Reader, privKey, crypto.SHA256, hashed)
	if err != nil {
		t.Fatalf("failed to sign jwt: %v", err)
	}
	sB64 := base64.RawURLEncoding.EncodeToString(sigBytes)

	return signedContent + "." + sB64
}

