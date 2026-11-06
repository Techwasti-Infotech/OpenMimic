package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// SCIMUser represents an RFC 7643 SCIM User resource
type SCIMUser struct {
	Schemas  []string       `json:"schemas"`
	ID       string         `json:"id"`
	UserName string         `json:"userName"`
	Name     SCIMName       `json:"name"`
	Emails   []SCIMEmail    `json:"emails"`
	Active   bool           `json:"active"`
	Roles    []SCIMRole     `json:"roles,omitempty"`
	Groups   []SCIMGroupRef `json:"groups,omitempty"`
	Meta     SCIMMeta       `json:"meta"`
}

type SCIMName struct {
	Formatted  string `json:"formatted"`
	FamilyName string `json:"familyName"`
	GivenName  string `json:"givenName"`
}

type SCIMEmail struct {
	Value   string `json:"value"`
	Primary bool   `json:"primary"`
	Type    string `json:"type"`
}

type SCIMRole struct {
	Value string `json:"value"` // e.g. "super_admin", "tenant_admin", "viewer"
}

type SCIMGroupRef struct {
	Value   string `json:"value"` // group ID
	Display string `json:"display"`
}

type SCIMMeta struct {
	ResourceType string    `json:"resourceType"`
	Created      time.Time `json:"created"`
	LastModified time.Time `json:"lastModified"`
	Location     string    `json:"location"`
}

// SCIMGroup represents an RFC 7643 SCIM Group resource
type SCIMGroup struct {
	Schemas     []string          `json:"schemas"`
	ID          string            `json:"id"`
	DisplayName string            `json:"displayName"`
	Members     []SCIMGroupMember `json:"members"`
	Meta        SCIMMeta          `json:"meta"`
}

type SCIMGroupMember struct {
	Value   string `json:"value"` // user ID
	Display string `json:"display"`
}

// SCIMListResponse represents a paginated RFC 7644 list response
type SCIMListResponse struct {
	Schemas      []string `json:"schemas"`
	TotalResults int      `json:"totalResults"`
	StartIndex   int      `json:"startIndex"`
	ItemsPerPage int      `json:"itemsPerPage"`
	Resources    []any    `json:"Resources"`
}

// SCIMService handles SCIM 2.0 provisioning for Okta, Entra ID, and other IdPs
type SCIMService struct {
	mu          sync.RWMutex
	users       map[string]*SCIMUser
	groups      map[string]*SCIMGroup
	bearerToken string
}

// NewSCIMService instantiates an in-memory RFC 7644 SCIM 2.0 service
func NewSCIMService(bearerToken string) *SCIMService {
	return &SCIMService{
		users:       make(map[string]*SCIMUser),
		groups:      make(map[string]*SCIMGroup),
		bearerToken: bearerToken,
	}
}

// Handler returns the HTTP handler serving /scim/v2/...
func (s *SCIMService) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/scim/v2/Users", s.handleUsers)
	mux.HandleFunc("/scim/v2/Users/", s.handleUserByID)
	mux.HandleFunc("/scim/v2/Groups", s.handleGroups)
	mux.HandleFunc("/scim/v2/Groups/", s.handleGroupByID)
	mux.HandleFunc("/scim/v2/ServiceProviderConfig", s.handleServiceProviderConfig)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Authorization check
		if s.bearerToken != "" {
			authHeader := r.Header.Get("Authorization")
			expected := "Bearer " + s.bearerToken
			if authHeader != expected {
				w.Header().Set("Content-Type", "application/scim+json")
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"schemas": []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
					"detail":  "Unauthorized SCIM request",
					"status":  "401",
				})
				return
			}
		}
		mux.ServeHTTP(w, r)
	})
}

func (s *SCIMService) handleUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/scim+json")
	switch r.Method {
	case http.MethodGet:
		s.mu.RLock()
		defer s.mu.RUnlock()
		var res []any
		for _, u := range s.users {
			res = append(res, u)
		}
		_ = json.NewEncoder(w).Encode(SCIMListResponse{
			Schemas:      []string{"urn:ietf:params:scim:api:messages:2.0:ListResponse"},
			TotalResults: len(res),
			StartIndex:   1,
			ItemsPerPage: len(res),
			Resources:    res,
		})

	case http.MethodPost:
		var u SCIMUser
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			s.writeError(w, http.StatusBadRequest, "Invalid JSON body: "+err.Error())
			return
		}
		if u.UserName == "" {
			s.writeError(w, http.StatusBadRequest, "userName is required")
			return
		}
		s.mu.Lock()
		defer s.mu.Unlock()

		id := fmt.Sprintf("usr-%d", time.Now().UnixNano())
		u.ID = id
		u.Schemas = []string{"urn:ietf:params:scim:schemas:core:2.0:User"}
		now := time.Now().UTC()
		u.Meta = SCIMMeta{
			ResourceType: "User",
			Created:      now,
			LastModified: now,
			Location:     "/scim/v2/Users/" + id,
		}
		s.users[id] = &u

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(u)

	default:
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (s *SCIMService) handleUserByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/scim+json")
	id := strings.TrimPrefix(r.URL.Path, "/scim/v2/Users/")
	if id == "" {
		s.writeError(w, http.StatusBadRequest, "Missing user ID")
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.mu.RLock()
		user, exists := s.users[id]
		s.mu.RUnlock()
		if !exists {
			s.writeError(w, http.StatusNotFound, "User not found")
			return
		}
		_ = json.NewEncoder(w).Encode(user)

	case http.MethodPut:
		var u SCIMUser
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			s.writeError(w, http.StatusBadRequest, "Invalid JSON body")
			return
		}
		s.mu.Lock()
		defer s.mu.Unlock()
		existing, exists := s.users[id]
		if !exists {
			s.writeError(w, http.StatusNotFound, "User not found")
			return
		}
		u.ID = id
		u.Meta.Created = existing.Meta.Created
		u.Meta.LastModified = time.Now().UTC()
		u.Meta.ResourceType = "User"
		u.Meta.Location = "/scim/v2/Users/" + id
		s.users[id] = &u
		_ = json.NewEncoder(w).Encode(u)

	case http.MethodDelete:
		s.mu.Lock()
		defer s.mu.Unlock()
		if _, exists := s.users[id]; !exists {
			s.writeError(w, http.StatusNotFound, "User not found")
			return
		}
		delete(s.users, id)
		w.WriteHeader(http.StatusNoContent)

	default:
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (s *SCIMService) handleGroups(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/scim+json")
	switch r.Method {
	case http.MethodGet:
		s.mu.RLock()
		defer s.mu.RUnlock()
		var res []any
		for _, g := range s.groups {
			res = append(res, g)
		}
		_ = json.NewEncoder(w).Encode(SCIMListResponse{
			Schemas:      []string{"urn:ietf:params:scim:api:messages:2.0:ListResponse"},
			TotalResults: len(res),
			StartIndex:   1,
			ItemsPerPage: len(res),
			Resources:    res,
		})

	case http.MethodPost:
		var g SCIMGroup
		if err := json.NewDecoder(r.Body).Decode(&g); err != nil {
			s.writeError(w, http.StatusBadRequest, "Invalid JSON body")
			return
		}
		if g.DisplayName == "" {
			s.writeError(w, http.StatusBadRequest, "displayName is required")
			return
		}
		s.mu.Lock()
		defer s.mu.Unlock()
		id := fmt.Sprintf("grp-%d", time.Now().UnixNano())
		g.ID = id
		g.Schemas = []string{"urn:ietf:params:scim:schemas:core:2.0:Group"}
		now := time.Now().UTC()
		g.Meta = SCIMMeta{
			ResourceType: "Group",
			Created:      now,
			LastModified: now,
			Location:     "/scim/v2/Groups/" + id,
		}
		s.groups[id] = &g
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(g)

	default:
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (s *SCIMService) handleGroupByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/scim+json")
	id := strings.TrimPrefix(r.URL.Path, "/scim/v2/Groups/")
	if id == "" {
		s.writeError(w, http.StatusBadRequest, "Missing group ID")
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.mu.RLock()
		group, exists := s.groups[id]
		s.mu.RUnlock()
		if !exists {
			s.writeError(w, http.StatusNotFound, "Group not found")
			return
		}
		_ = json.NewEncoder(w).Encode(group)

	case http.MethodDelete:
		s.mu.Lock()
		defer s.mu.Unlock()
		if _, exists := s.groups[id]; !exists {
			s.writeError(w, http.StatusNotFound, "Group not found")
			return
		}
		delete(s.groups, id)
		w.WriteHeader(http.StatusNoContent)

	default:
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (s *SCIMService) handleServiceProviderConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/scim+json")
	config := map[string]any{
		"schemas":        []string{"urn:ietf:params:scim:schemas:core:2.0:ServiceProviderConfig"},
		"patch":          map[string]bool{"supported": false},
		"bulk":           map[string]any{"supported": false, "maxOperations": 0, "maxPayloadSize": 0},
		"filter":         map[string]any{"supported": false, "maxResults": 100},
		"changePassword": map[string]bool{"supported": false},
		"sort":           map[string]bool{"supported": false},
		"etag":           map[string]bool{"supported": false},
		"authenticationSchemes": []map[string]any{
			{
				"name":        "OAuth Bearer Token",
				"description": "Authentication scheme using the OAuth Bearer Token Standard",
				"specUri":     "http://www.rfc-editor.org/info/rfc6750",
				"type":        "oauthbearertoken",
				"primary":     true,
			},
		},
	}
	_ = json.NewEncoder(w).Encode(config)
}

func (s *SCIMService) writeError(w http.ResponseWriter, code int, detail string) {
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"schemas": []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
		"detail":  detail,
		"status":  fmt.Sprintf("%d", code),
	})
}
