package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSCIMUserProvisioningLifecycle(t *testing.T) {
	bearerToken := "super-secure-scim-token-999"
	service := NewSCIMService(bearerToken)
	handler := service.Handler()

	// 1. Unauthorized request
	reqUnauth := httptest.NewRequest(http.MethodGet, "/scim/v2/Users", nil)
	rrUnauth := httptest.NewRecorder()
	handler.ServeHTTP(rrUnauth, reqUnauth)
	if rrUnauth.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized for missing bearer token, got %d", rrUnauth.Code)
	}

	// 2. Create User (POST /scim/v2/Users)
	newUser := map[string]any{
		"userName": "alice.engineer@techwasti.com",
		"name": map[string]string{
			"formatted":  "Alice Engineer",
			"familyName": "Engineer",
			"givenName":  "Alice",
		},
		"emails": []map[string]any{
			{"value": "alice.engineer@techwasti.com", "primary": true, "type": "work"},
		},
		"active": true,
		"roles": []map[string]string{
			{"value": "tenant_admin"},
		},
	}
	body, _ := json.Marshal(newUser)
	reqCreate := httptest.NewRequest(http.MethodPost, "/scim/v2/Users", bytes.NewReader(body))
	reqCreate.Header.Set("Authorization", "Bearer "+bearerToken)
	reqCreate.Header.Set("Content-Type", "application/scim+json")
	rrCreate := httptest.NewRecorder()
	handler.ServeHTTP(rrCreate, reqCreate)

	if rrCreate.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d body=%s", rrCreate.Code, rrCreate.Body.String())
	}

	var created SCIMUser
	if err := json.Unmarshal(rrCreate.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to decode created SCIM user: %v", err)
	}
	if created.ID == "" || created.UserName != "alice.engineer@techwasti.com" {
		t.Fatalf("unexpected created user details: %+v", created)
	}

	// 3. Get User By ID (GET /scim/v2/Users/{id})
	reqGet := httptest.NewRequest(http.MethodGet, "/scim/v2/Users/"+created.ID, nil)
	reqGet.Header.Set("Authorization", "Bearer "+bearerToken)
	rrGet := httptest.NewRecorder()
	handler.ServeHTTP(rrGet, reqGet)

	if rrGet.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for GET user, got %d", rrGet.Code)
	}

	// 4. List Users (GET /scim/v2/Users)
	reqList := httptest.NewRequest(http.MethodGet, "/scim/v2/Users", nil)
	reqList.Header.Set("Authorization", "Bearer "+bearerToken)
	rrList := httptest.NewRecorder()
	handler.ServeHTTP(rrList, reqList)

	var listResp SCIMListResponse
	if err := json.Unmarshal(rrList.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("failed to decode SCIM list: %v", err)
	}
	if listResp.TotalResults != 1 {
		t.Fatalf("expected 1 total result, got %d", listResp.TotalResults)
	}

	// 5. Update User (PUT /scim/v2/Users/{id})
	created.Active = false // Deactivate user
	updateBody, _ := json.Marshal(created)
	reqPut := httptest.NewRequest(http.MethodPut, "/scim/v2/Users/"+created.ID, bytes.NewReader(updateBody))
	reqPut.Header.Set("Authorization", "Bearer "+bearerToken)
	rrPut := httptest.NewRecorder()
	handler.ServeHTTP(rrPut, reqPut)

	if rrPut.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for update, got %d", rrPut.Code)
	}

	// 6. Delete User (DELETE /scim/v2/Users/{id})
	reqDel := httptest.NewRequest(http.MethodDelete, "/scim/v2/Users/"+created.ID, nil)
	reqDel.Header.Set("Authorization", "Bearer "+bearerToken)
	rrDel := httptest.NewRecorder()
	handler.ServeHTTP(rrDel, reqDel)

	if rrDel.Code != http.StatusNoContent {
		t.Fatalf("expected 204 No Content for delete, got %d", rrDel.Code)
	}

	// 7. Verify Deleted User returns 404
	reqGetDeleted := httptest.NewRequest(http.MethodGet, "/scim/v2/Users/"+created.ID, nil)
	reqGetDeleted.Header.Set("Authorization", "Bearer "+bearerToken)
	rrGetDeleted := httptest.NewRecorder()
	handler.ServeHTTP(rrGetDeleted, reqGetDeleted)
	if rrGetDeleted.Code != http.StatusNotFound {
		t.Fatalf("expected 404 Not Found, got %d", rrGetDeleted.Code)
	}
}

func TestSCIMGroupManagement(t *testing.T) {
	bearerToken := "scim-token"
	service := NewSCIMService(bearerToken)
	handler := service.Handler()

	// 1. Create Group (POST /scim/v2/Groups)
	newGroup := map[string]any{
		"displayName": "DataPlatform-Admins",
		"members": []map[string]string{
			{"value": "usr-12345", "display": "Lead Architect"},
		},
	}
	body, _ := json.Marshal(newGroup)
	req := httptest.NewRequest(http.MethodPost, "/scim/v2/Groups", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+bearerToken)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for group, got %d body=%s", rr.Code, rr.Body.String())
	}

	var group SCIMGroup
	_ = json.Unmarshal(rr.Body.Bytes(), &group)
	if group.DisplayName != "DataPlatform-Admins" {
		t.Errorf("expected displayName DataPlatform-Admins, got %s", group.DisplayName)
	}

	// 2. ServiceProviderConfig endpoint
	reqCfg := httptest.NewRequest(http.MethodGet, "/scim/v2/ServiceProviderConfig", nil)
	reqCfg.Header.Set("Authorization", "Bearer "+bearerToken)
	rrCfg := httptest.NewRecorder()
	handler.ServeHTTP(rrCfg, reqCfg)
	if rrCfg.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for ServiceProviderConfig, got %d", rrCfg.Code)
	}
}
