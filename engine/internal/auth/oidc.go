package auth

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	pb "openmimic/engine/pkg/proto/v1"
)

// JWK represents a cryptographic key in a JSON Web Key Set.
type JWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// JWKS represents a JSON Web Key Set response from an OIDC provider.
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// OIDCDiscoveryDoc represents standard OpenID Connect discovery metadata.
type OIDCDiscoveryDoc struct {
	Issuer                string `json:"issuer"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	UserinfoEndpoint      string `json:"userinfo_endpoint"`
	JwksURI               string `json:"jwks_uri"`
}

// EnterpriseOIDCValidator handles OIDC discovery, JWKS key caching, and JWT verification.
type EnterpriseOIDCValidator struct {
	mu          sync.RWMutex
	cfg         *pb.EnterpriseSSOConfig
	httpClient  *http.Client
	cachedKeys  map[string]*rsa.PublicKey
	lastRefresh time.Time
	discovery   *OIDCDiscoveryDoc
}

// NewEnterpriseOIDCValidator initializes an OIDC validator from configuration.
func NewEnterpriseOIDCValidator(cfg *pb.EnterpriseSSOConfig, client *http.Client) *EnterpriseOIDCValidator {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	return &EnterpriseOIDCValidator{
		cfg:        cfg,
		httpClient: client,
		cachedKeys: make(map[string]*rsa.PublicKey),
	}
}

// RegisterKey adds an RSA public key directly (useful for tests or air-gapped environments).
func (v *EnterpriseOIDCValidator) RegisterPublicKey(kid string, pubKey *rsa.PublicKey) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.cachedKeys[kid] = pubKey
}

// DiscoverMetadata queries the .well-known/openid-configuration endpoint.
func (v *EnterpriseOIDCValidator) DiscoverMetadata() (*OIDCDiscoveryDoc, error) {
	if v.cfg == nil || v.cfg.GetIssuerUrl() == "" {
		return nil, fmt.Errorf("issuer_url is required for OIDC discovery")
	}

	endpoint := strings.TrimRight(v.cfg.GetIssuerUrl(), "/") + "/.well-known/openid-configuration"
	resp, err := v.httpClient.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch openid configuration: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openid configuration endpoint returned status %d", resp.StatusCode)
	}

	var doc OIDCDiscoveryDoc
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, fmt.Errorf("failed to decode openid configuration: %w", err)
	}

	v.mu.Lock()
	v.discovery = &doc
	v.mu.Unlock()

	return &doc, nil
}

// RefreshKeys downloads and parses the JWKS public keys.
func (v *EnterpriseOIDCValidator) RefreshKeys() error {
	jwksURI := ""
	if v.cfg != nil && v.cfg.GetJwksUrl() != "" {
		jwksURI = v.cfg.GetJwksUrl()
	} else if v.discovery != nil && v.discovery.JwksURI != "" {
		jwksURI = v.discovery.JwksURI
	} else {
		// Discover if not known
		doc, err := v.DiscoverMetadata()
		if err != nil {
			return err
		}
		jwksURI = doc.JwksURI
	}

	resp, err := v.httpClient.Get(jwksURI)
	if err != nil {
		return fmt.Errorf("failed to fetch jwks: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var jwks JWKS
	if err := json.Unmarshal(body, &jwks); err != nil {
		return fmt.Errorf("failed to parse jwks json: %w", err)
	}

	v.mu.Lock()
	defer v.mu.Unlock()

	for _, k := range jwks.Keys {
		if k.Kty == "RSA" && k.N != "" && k.E != "" {
			pubKey, err := parseRSAPublicKey(k.N, k.E)
			if err == nil {
				v.cachedKeys[k.Kid] = pubKey
			}
		}
	}
	v.lastRefresh = time.Now()
	return nil
}

// TokenClaims represents decoded enterprise JWT claims.
type TokenClaims struct {
	Issuer    string   `json:"iss"`
	Audience  any      `json:"aud"`
	Subject   string   `json:"sub"`
	Email     string   `json:"email"`
	Name      string   `json:"name"`
	Groups    []string `json:"groups"`
	Roles     []string `json:"roles"`
	ExpiresAt int64    `json:"exp"`
	NotBefore int64    `json:"nbf"`
	IssuedAt  int64    `json:"iat"`
}

// ValidateJWT verifies signature, issuer, audience, and expiration, returning Claims with mapped Role.
func (v *EnterpriseOIDCValidator) ValidateJWT(tokenStr string) (*Claims, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid jwt format: expected 3 segments")
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid jwt header base64: %w", err)
	}

	var header struct {
		Alg string `json:"alg"`
		Kid string `json:"kid"`
	}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, fmt.Errorf("invalid jwt header json: %w", err)
	}

	if header.Alg != "RS256" {
		return nil, fmt.Errorf("unsupported jwt algorithm: %s (only RS256 supported)", header.Alg)
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid jwt payload base64: %w", err)
	}

	var tClaims TokenClaims
	if err := json.Unmarshal(payloadBytes, &tClaims); err != nil {
		return nil, fmt.Errorf("invalid jwt payload json: %w", err)
	}

	// Expiration check (with 30-second clock skew allowance)
	now := time.Now().Unix()
	if tClaims.ExpiresAt > 0 && now > (tClaims.ExpiresAt+30) {
		return nil, fmt.Errorf("jwt token has expired")
	}
	if tClaims.NotBefore > 0 && now < (tClaims.NotBefore-30) {
		return nil, fmt.Errorf("jwt token is not valid yet")
	}

	// Audience check if configured
	if v.cfg != nil && v.cfg.GetClientId() != "" {
		if !matchAudience(tClaims.Audience, v.cfg.GetClientId()) {
			return nil, fmt.Errorf("jwt audience mismatch: expected client_id %s", v.cfg.GetClientId())
		}
	}

	// Verify cryptographic RS256 signature
	v.mu.RLock()
	pubKey, exists := v.cachedKeys[header.Kid]
	v.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("unknown key id (kid: %s) - key not found in JWKS cache", header.Kid)
	}

	sigBytes, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("invalid jwt signature base64: %w", err)
	}

	signedContent := parts[0] + "." + parts[1]
	hasher := sha256.New()
	hasher.Write([]byte(signedContent))
	hashed := hasher.Sum(nil)

	if err := rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hashed, sigBytes); err != nil {
		return nil, fmt.Errorf("jwt signature verification failed: %w", err)
	}

	// Map enterprise groups/roles to OpenMimic Role
	resolvedRole := v.ResolveRole(tClaims.Groups, tClaims.Roles)

	tenant := "default"
	if v.cfg != nil && v.cfg.GetIssuerUrl() != "" {
		tenant = v.cfg.GetIssuerUrl()
	}

	return &Claims{
		Subject:  tClaims.Subject,
		TenantID: tenant,
		Role:     resolvedRole,
		Expires:  tClaims.ExpiresAt,
	}, nil
}

// ResolveRole maps IdP groups or roles according to GroupRoleMappings.
func (v *EnterpriseOIDCValidator) ResolveRole(groups, roles []string) Role {
	defaultRole := RoleViewer
	if v.cfg != nil && v.cfg.GetDefaultRole() != "" {
		defaultRole = Role(v.cfg.GetDefaultRole())
	}

	if v.cfg == nil || len(v.cfg.GetGroupRoleMappings()) == 0 {
		return defaultRole
	}

	mappings := v.cfg.GetGroupRoleMappings()

	// Check groups first
	for _, grp := range groups {
		if targetRole, found := mappings[grp]; found {
			return Role(targetRole)
		}
	}

	// Check roles
	for _, r := range roles {
		if targetRole, found := mappings[r]; found {
			return Role(targetRole)
		}
	}

	return defaultRole
}

func matchAudience(aud any, expected string) bool {
	switch v := aud.(type) {
	case string:
		return v == expected
	case []any:
		for _, item := range v {
			if s, ok := item.(string); ok && s == expected {
				return true
			}
		}
	case []string:
		for _, s := range v {
			if s == expected {
				return true
			}
		}
	}
	return false
}

func parseRSAPublicKey(nStr, eStr string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(nStr)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(eStr)
	if err != nil {
		return nil, err
	}

	n := new(big.Int).SetBytes(nBytes)
	var e int
	for _, b := range eBytes {
		e = (e << 8) | int(b)
	}

	return &rsa.PublicKey{
		N: n,
		E: e,
	}, nil
}

