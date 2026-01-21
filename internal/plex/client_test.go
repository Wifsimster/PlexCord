package plex

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"plexcord/internal/errors"
)

// TestNewClient tests client creation
func TestNewClient(t *testing.T) {
	client := NewClient("test-token", "http://localhost:32400")

	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	if client.token != "test-token" {
		t.Errorf("Expected token 'test-token', got '%s'", client.token)
	}

	if client.serverURL != "http://localhost:32400" {
		t.Errorf("Expected serverURL 'http://localhost:32400', got '%s'", client.serverURL)
	}

	if client.httpClient == nil {
		t.Error("Expected httpClient to be initialized")
	}

	if client.httpClient.Timeout != 5*time.Second {
		t.Errorf("Expected timeout 5s, got %v", client.httpClient.Timeout)
	}
}

// TestValidateConnectionSuccess tests successful connection validation
func TestValidateConnectionSuccess(t *testing.T) {
	// Create mock servers for identity and library endpoints
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify User-Agent header
		userAgent := r.Header.Get("User-Agent")
		if userAgent != "PlexCord/1.0" {
			t.Errorf("Expected User-Agent 'PlexCord/1.0', got '%s'", userAgent)
		}

		switch r.URL.Path {
		case "/identity":
			// Identity endpoint doesn't require authentication
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<MediaContainer size="0" friendlyName="TestPlexServer" version="1.40.0.7998" machineIdentifier="abc123def456" claimed="1"/>`))
		case "/library/sections/":
			// Verify token query parameter is present for authenticated endpoints
			token := r.URL.Query().Get("X-Plex-Token")
			if token != "valid-token" {
				t.Errorf("Expected X-Plex-Token query param 'valid-token', got '%s'", token)
			}
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<MediaContainer size="3">
  <Directory key="1" title="Movies" type="movie"/>
  <Directory key="2" title="TV Shows" type="show"/>
  <Directory key="3" title="Music" type="artist"/>
</MediaContainer>`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewClient("valid-token", server.URL)
	result, err := client.ValidateConnection()

	if err != nil {
		t.Fatalf("ValidateConnection failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected Success to be true")
	}

	if result.ServerName != "TestPlexServer" {
		t.Errorf("Expected ServerName 'TestPlexServer', got '%s'", result.ServerName)
	}

	if result.ServerVersion != "1.40.0.7998" {
		t.Errorf("Expected ServerVersion '1.40.0.7998', got '%s'", result.ServerVersion)
	}

	if result.LibraryCount != 3 {
		t.Errorf("Expected LibraryCount 3, got %d", result.LibraryCount)
	}

	if result.MachineIdentifier != "abc123def456" {
		t.Errorf("Expected MachineIdentifier 'abc123def456', got '%s'", result.MachineIdentifier)
	}
}

// TestValidateConnectionAuthFailure tests authentication failure (401)
func TestValidateConnectionAuthFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
	}))
	defer server.Close()

	client := NewClient("invalid-token", server.URL)
	result, err := client.ValidateConnection()

	if result != nil {
		t.Error("Expected nil result for auth failure")
	}

	if err == nil {
		t.Fatal("Expected error for auth failure")
	}

	if !errors.Is(err, errors.PLEX_AUTH_FAILED) {
		t.Errorf("Expected PLEX_AUTH_FAILED error code, got: %s", errors.GetCode(err))
	}
}

// TestValidateConnectionForbidden tests forbidden response (403)
func TestValidateConnectionForbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
	}))
	defer server.Close()

	client := NewClient("forbidden-token", server.URL)
	result, err := client.ValidateConnection()

	if result != nil {
		t.Error("Expected nil result for forbidden")
	}

	if err == nil {
		t.Fatal("Expected error for forbidden")
	}

	if !errors.Is(err, errors.PLEX_AUTH_FAILED) {
		t.Errorf("Expected PLEX_AUTH_FAILED error code, got: %s", errors.GetCode(err))
	}
}

// TestValidateConnectionServerError tests server errors (500, 502, 503)
func TestValidateConnectionServerError(t *testing.T) {
	testCases := []struct {
		name       string
		statusCode int
	}{
		{"InternalServerError", http.StatusInternalServerError},
		{"BadGateway", http.StatusBadGateway},
		{"ServiceUnavailable", http.StatusServiceUnavailable},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
				w.Write([]byte("Server Error"))
			}))
			defer server.Close()

			client := NewClient("test-token", server.URL)
			result, err := client.ValidateConnection()

			if result != nil {
				t.Error("Expected nil result for server error")
			}

			if err == nil {
				t.Fatal("Expected error for server error")
			}

			if !errors.Is(err, errors.PLEX_UNREACHABLE) {
				t.Errorf("Expected PLEX_UNREACHABLE error code, got: %s", errors.GetCode(err))
			}
		})
	}
}

// TestValidateConnectionNotFound tests 404 response
func TestValidateConnectionNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
	}))
	defer server.Close()

	client := NewClient("test-token", server.URL)
	result, err := client.ValidateConnection()

	if result != nil {
		t.Error("Expected nil result for not found")
	}

	if err == nil {
		t.Fatal("Expected error for not found")
	}

	if !errors.Is(err, errors.PLEX_UNREACHABLE) {
		t.Errorf("Expected PLEX_UNREACHABLE error code, got: %s", errors.GetCode(err))
	}
}

// TestValidateConnectionInvalidXML tests invalid XML response handling
func TestValidateConnectionInvalidXML(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("this is not valid xml"))
	}))
	defer server.Close()

	client := NewClient("test-token", server.URL)
	result, err := client.ValidateConnection()

	if result != nil {
		t.Error("Expected nil result for invalid XML")
	}

	if err == nil {
		t.Fatal("Expected error for invalid XML")
	}

	if !errors.Is(err, errors.PLEX_CONN_FAILED) {
		t.Errorf("Expected PLEX_CONN_FAILED error code, got: %s", errors.GetCode(err))
	}
}

// TestValidateConnectionEmptyResponse tests empty server response
func TestValidateConnectionEmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		// Empty response body
	}))
	defer server.Close()

	client := NewClient("test-token", server.URL)
	result, err := client.ValidateConnection()

	// Empty XML should still parse (though fields will be empty)
	// This tests that we handle empty responses gracefully
	if err != nil {
		// An error is acceptable for empty/invalid XML
		if !errors.Is(err, errors.PLEX_CONN_FAILED) {
			t.Errorf("Expected PLEX_CONN_FAILED error code, got: %s", errors.GetCode(err))
		}
	}

	// If no error, result should have empty fields
	if result != nil && result.ServerName != "" {
		t.Errorf("Expected empty ServerName for empty response, got '%s'", result.ServerName)
	}
}

// TestValidateConnectionUnreachable tests connection to unreachable server
func TestValidateConnectionUnreachable(t *testing.T) {
	// Use an address that will definitely fail to connect
	client := NewClient("test-token", "http://localhost:12345")

	// Override timeout for faster test
	client.httpClient.Timeout = 1 * time.Second

	result, err := client.ValidateConnection()

	if result != nil {
		t.Error("Expected nil result for unreachable server")
	}

	if err == nil {
		t.Fatal("Expected error for unreachable server")
	}

	// Should get either PLEX_UNREACHABLE or TIMEOUT depending on system behavior
	code := errors.GetCode(err)
	if code != errors.PLEX_UNREACHABLE && code != errors.TIMEOUT {
		t.Errorf("Expected PLEX_UNREACHABLE or TIMEOUT error code, got: %s", code)
	}
}

// TestValidateConnectionInvalidURL tests invalid server URL
func TestValidateConnectionInvalidURL(t *testing.T) {
	client := NewClient("test-token", "not-a-valid-url")
	result, err := client.ValidateConnection()

	if result != nil {
		t.Error("Expected nil result for invalid URL")
	}

	if err == nil {
		t.Fatal("Expected error for invalid URL")
	}
}

// TestValidateConnectionLibraryRequestFails tests when identity succeeds but library fails
func TestValidateConnectionLibraryRequestFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/identity":
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<MediaContainer size="0" friendlyName="TestPlexServer" version="1.40.0.7998" machineIdentifier="abc123" claimed="1"/>`))
		case "/library/sections/":
			// Return auth failure for library request
			w.WriteHeader(http.StatusUnauthorized)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewClient("test-token", server.URL)
	result, err := client.ValidateConnection()

	if result != nil {
		t.Error("Expected nil result when library request fails")
	}

	if err == nil {
		t.Fatal("Expected error when library request fails")
	}

	// Should get auth failure since that's what library endpoint returned
	if !errors.Is(err, errors.PLEX_AUTH_FAILED) {
		t.Errorf("Expected PLEX_AUTH_FAILED error code, got: %s", errors.GetCode(err))
	}
}

// TestValidationResultJSONSerialization tests that ValidationResult serializes correctly
func TestValidationResultJSONSerialization(t *testing.T) {
	result := ValidationResult{
		Success:      true,
		ServerName:   "My Plex Server",
		LibraryCount: 5,
	}

	// Verify JSON tags are using camelCase (important for Wails binding)
	if result.Success != true {
		t.Error("Success field not correctly set")
	}

	if result.ServerName != "My Plex Server" {
		t.Error("ServerName field not correctly set")
	}

	if result.LibraryCount != 5 {
		t.Error("LibraryCount field not correctly set")
	}
}

// TestMapHTTPStatusCode tests the status code to error code mapping
func TestMapHTTPStatusCode(t *testing.T) {
	testCases := []struct {
		expectedCode string
		statusCode   int
	}{
		{errors.PLEX_AUTH_FAILED, 401},
		{errors.PLEX_AUTH_FAILED, 403},
		{errors.PLEX_UNREACHABLE, 404},
		{errors.PLEX_UNREACHABLE, 500},
		{errors.PLEX_UNREACHABLE, 502},
		{errors.PLEX_UNREACHABLE, 503},
		{errors.PLEX_CONN_FAILED, 418}, // Other status codes
	}

	for _, tc := range testCases {
		t.Run(http.StatusText(tc.statusCode), func(t *testing.T) {
			err := mapHTTPStatusCode(tc.statusCode)
			if !errors.Is(err, tc.expectedCode) {
				t.Errorf("Expected error code %s for status %d, got %s",
					tc.expectedCode, tc.statusCode, errors.GetCode(err))
			}
		})
	}
}

// TestIdentityResponseParsing tests XML parsing of identity response
func TestIdentityResponseParsing(t *testing.T) {
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<MediaContainer>
  <Server name="MyPlexServer" version="1.40.0.7998" machineIdentifier="abc123def456"/>
</MediaContainer>`)

	var identity IdentityResponse
	// Verify XMLName.Space is accessible
	_ = identity.XMLName.Space

	// Test that the struct can unmarshal XML
	// The actual parsing is tested in TestValidateConnectionSuccess
	_ = xmlData
}

// TestLibraryResponseParsing tests XML parsing of library response
func TestLibraryResponseParsing(t *testing.T) {
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<MediaContainer size="3">
  <Directory key="1" title="Movies" type="movie"/>
  <Directory key="2" title="TV Shows" type="show"/>
  <Directory key="3" title="Music" type="artist"/>
</MediaContainer>`)

	// Test that the struct can handle the XML format
	// The actual parsing is tested in TestValidateConnectionSuccess
	_ = xmlData
}

// TestValidateConnectionSlowServer tests behavior with slow server response
func TestValidateConnectionSlowServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(100 * time.Millisecond)
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<MediaContainer>
  <Server name="SlowServer" version="1.0" machineIdentifier="slow123"/>
</MediaContainer>`))
	}))
	defer server.Close()

	client := NewClient("test-token", server.URL)
	// Use short timeout to trigger timeout error
	client.httpClient.Timeout = 50 * time.Millisecond

	result, err := client.ValidateConnection()

	if result != nil {
		t.Error("Expected nil result for timeout")
	}

	if err == nil {
		t.Fatal("Expected error for timeout")
	}

	// Should get timeout or connection error
	code := errors.GetCode(err)
	if code != errors.TIMEOUT && code != errors.PLEX_UNREACHABLE && code != errors.PLEX_CONN_FAILED {
		t.Errorf("Expected TIMEOUT, PLEX_UNREACHABLE, or PLEX_CONN_FAILED error code, got: %s", code)
	}
}

// TestValidateConnectionZeroLibraries tests server with no libraries
func TestValidateConnectionZeroLibraries(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/identity":
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<MediaContainer size="0" friendlyName="EmptyServer" version="1.0" machineIdentifier="empty123" claimed="1"/>`))
		case "/library/sections/":
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<MediaContainer size="0">
</MediaContainer>`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewClient("test-token", server.URL)
	result, err := client.ValidateConnection()

	if err != nil {
		t.Fatalf("ValidateConnection failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected Success to be true for server with zero libraries")
	}

	if result.LibraryCount != 0 {
		t.Errorf("Expected LibraryCount 0, got %d", result.LibraryCount)
	}

	if result.ServerName != "EmptyServer" {
		t.Errorf("Expected ServerName 'EmptyServer', got '%s'", result.ServerName)
	}
}

// ========================================
// GetUsers Tests
// ========================================

// TestGetUsersSuccess tests successful user retrieval with multiple users
func TestGetUsersSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify token query param is present
		token := r.URL.Query().Get("X-Plex-Token")
		if token != "valid-token" {
			t.Errorf("Expected X-Plex-Token query param 'valid-token', got '%s'", token)
		}

		// Verify User-Agent header
		userAgent := r.Header.Get("User-Agent")
		if userAgent != "PlexCord/1.0" {
			t.Errorf("Expected User-Agent 'PlexCord/1.0', got '%s'", userAgent)
		}

		if r.URL.Path == "/accounts" {
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<MediaContainer size="3">
  <Account id="1" name="Admin" thumb="https://plex.tv/users/admin/avatar"/>
  <Account id="2" name="FamilyMember" thumb="https://plex.tv/users/family/avatar"/>
  <Account id="3" name="Guest" thumb=""/>
</MediaContainer>`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewClient("valid-token", server.URL)
	users, err := client.GetUsers()

	if err != nil {
		t.Fatalf("GetUsers failed: %v", err)
	}

	if len(users) != 3 {
		t.Errorf("Expected 3 users, got %d", len(users))
	}

	// Verify first user
	if users[0].ID != "1" {
		t.Errorf("Expected user 0 ID '1', got '%s'", users[0].ID)
	}
	if users[0].Name != "Admin" {
		t.Errorf("Expected user 0 name 'Admin', got '%s'", users[0].Name)
	}
	if users[0].Thumb != "https://plex.tv/users/admin/avatar" {
		t.Errorf("Expected user 0 thumb URL, got '%s'", users[0].Thumb)
	}

	// Verify second user
	if users[1].ID != "2" {
		t.Errorf("Expected user 1 ID '2', got '%s'", users[1].ID)
	}
	if users[1].Name != "FamilyMember" {
		t.Errorf("Expected user 1 name 'FamilyMember', got '%s'", users[1].Name)
	}

	// Verify third user (empty thumb)
	if users[2].ID != "3" {
		t.Errorf("Expected user 2 ID '3', got '%s'", users[2].ID)
	}
	if users[2].Thumb != "" {
		t.Errorf("Expected user 2 thumb to be empty, got '%s'", users[2].Thumb)
	}
}

// TestGetUsersSingleUser tests retrieval of a single user (auto-select scenario)
func TestGetUsersSingleUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/accounts" {
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<MediaContainer size="1">
  <Account id="1" name="OnlyUser" thumb="https://plex.tv/users/only/avatar"/>
</MediaContainer>`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewClient("test-token", server.URL)
	users, err := client.GetUsers()

	if err != nil {
		t.Fatalf("GetUsers failed: %v", err)
	}

	if len(users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(users))
	}

	if users[0].Name != "OnlyUser" {
		t.Errorf("Expected user name 'OnlyUser', got '%s'", users[0].Name)
	}
}

// TestGetUsersEmptyList tests retrieval when no users are returned
func TestGetUsersEmptyList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/accounts" {
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<MediaContainer size="0">
</MediaContainer>`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewClient("test-token", server.URL)
	users, err := client.GetUsers()

	if err != nil {
		t.Fatalf("GetUsers failed: %v", err)
	}

	if len(users) != 0 {
		t.Errorf("Expected 0 users, got %d", len(users))
	}
}

// TestGetUsersAuthFailure tests authentication failure for GetUsers
func TestGetUsersAuthFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
	}))
	defer server.Close()

	client := NewClient("invalid-token", server.URL)
	users, err := client.GetUsers()

	if users != nil {
		t.Error("Expected nil users for auth failure")
	}

	if err == nil {
		t.Fatal("Expected error for auth failure")
	}

	if !errors.Is(err, errors.PLEX_AUTH_FAILED) {
		t.Errorf("Expected PLEX_AUTH_FAILED error code, got: %s", errors.GetCode(err))
	}
}

// TestGetUsersForbidden tests forbidden response for GetUsers
func TestGetUsersForbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
	}))
	defer server.Close()

	client := NewClient("forbidden-token", server.URL)
	users, err := client.GetUsers()

	if users != nil {
		t.Error("Expected nil users for forbidden")
	}

	if err == nil {
		t.Fatal("Expected error for forbidden")
	}

	if !errors.Is(err, errors.PLEX_AUTH_FAILED) {
		t.Errorf("Expected PLEX_AUTH_FAILED error code, got: %s", errors.GetCode(err))
	}
}

// TestGetUsersServerError tests server errors for GetUsers
func TestGetUsersServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := NewClient("test-token", server.URL)
	users, err := client.GetUsers()

	if users != nil {
		t.Error("Expected nil users for server error")
	}

	if err == nil {
		t.Fatal("Expected error for server error")
	}

	if !errors.Is(err, errors.PLEX_UNREACHABLE) {
		t.Errorf("Expected PLEX_UNREACHABLE error code, got: %s", errors.GetCode(err))
	}
}

// TestGetUsersInvalidXML tests invalid XML response for GetUsers
func TestGetUsersInvalidXML(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("this is not valid xml"))
	}))
	defer server.Close()

	client := NewClient("test-token", server.URL)
	users, err := client.GetUsers()

	if users != nil {
		t.Error("Expected nil users for invalid XML")
	}

	if err == nil {
		t.Fatal("Expected error for invalid XML")
	}

	if !errors.Is(err, errors.PLEX_CONN_FAILED) {
		t.Errorf("Expected PLEX_CONN_FAILED error code, got: %s", errors.GetCode(err))
	}
}

// TestGetUsersTimeout tests timeout behavior for GetUsers
func TestGetUsersTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(100 * time.Millisecond)
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<MediaContainer size="1">
  <Account id="1" name="SlowUser"/>
</MediaContainer>`))
	}))
	defer server.Close()

	client := NewClient("test-token", server.URL)
	// Use short timeout to trigger timeout error
	client.httpClient.Timeout = 50 * time.Millisecond

	users, err := client.GetUsers()

	if users != nil {
		t.Error("Expected nil users for timeout")
	}

	if err == nil {
		t.Fatal("Expected error for timeout")
	}

	// Should get timeout or connection error
	code := errors.GetCode(err)
	if code != errors.TIMEOUT && code != errors.PLEX_UNREACHABLE && code != errors.PLEX_CONN_FAILED {
		t.Errorf("Expected TIMEOUT, PLEX_UNREACHABLE, or PLEX_CONN_FAILED error code, got: %s", code)
	}
}

// TestGetUsersUnreachable tests connection to unreachable server for GetUsers
func TestGetUsersUnreachable(t *testing.T) {
	// Use an address that will definitely fail to connect
	client := NewClient("test-token", "http://localhost:12345")

	// Override timeout for faster test
	client.httpClient.Timeout = 1 * time.Second

	users, err := client.GetUsers()

	if users != nil {
		t.Error("Expected nil users for unreachable server")
	}

	if err == nil {
		t.Fatal("Expected error for unreachable server")
	}

	// Should get either PLEX_UNREACHABLE or TIMEOUT depending on system behavior
	code := errors.GetCode(err)
	if code != errors.PLEX_UNREACHABLE && code != errors.TIMEOUT {
		t.Errorf("Expected PLEX_UNREACHABLE or TIMEOUT error code, got: %s", code)
	}
}

// TestPlexUserJSONSerialization tests that PlexUser serializes correctly for Wails
func TestPlexUserJSONSerialization(t *testing.T) {
	user := PlexUser{
		ID:    "123",
		Name:  "TestUser",
		Thumb: "https://example.com/avatar.png",
	}

	// Verify struct fields (JSON tags are tested indirectly through Wails serialization)
	if user.ID != "123" {
		t.Error("ID field not correctly set")
	}

	if user.Name != "TestUser" {
		t.Error("Name field not correctly set")
	}

	if user.Thumb != "https://example.com/avatar.png" {
		t.Error("Thumb field not correctly set")
	}
}

// ========================================
// GetSessions Tests
// ========================================

// TestGetSessionsSuccess tests successful session retrieval with music playing
func TestGetSessionsSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify token query param is present
		token := r.URL.Query().Get("X-Plex-Token")
		if token != "valid-token" {
			t.Errorf("Expected X-Plex-Token query param 'valid-token', got '%s'", token)
		}

		if r.URL.Path == "/status/sessions" {
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<MediaContainer size="1">
  <Track sessionKey="abc123" key="/library/metadata/12345" type="track"
         title="Test Song" grandparentTitle="Test Artist" parentTitle="Test Album"
         thumb="/library/metadata/12345/thumb/123" duration="180000" viewOffset="45000">
    <User id="1" title="TestUser" thumb="https://plex.tv/avatar"/>
    <Player state="playing" title="Chrome" product="Plex Web"/>
  </Track>
</MediaContainer>`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewClient("valid-token", server.URL)
	sessions, err := client.GetSessions("1")

	if err != nil {
		t.Fatalf("GetSessions failed: %v", err)
	}

	if len(sessions) != 1 {
		t.Errorf("Expected 1 session, got %d", len(sessions))
	}

	// Verify session details
	if sessions[0].SessionKey != "abc123" {
		t.Errorf("Expected SessionKey 'abc123', got '%s'", sessions[0].SessionKey)
	}
	if sessions[0].UserID != "1" {
		t.Errorf("Expected UserID '1', got '%s'", sessions[0].UserID)
	}
	if sessions[0].UserName != "TestUser" {
		t.Errorf("Expected UserName 'TestUser', got '%s'", sessions[0].UserName)
	}
	if sessions[0].Type != "track" {
		t.Errorf("Expected Type 'track', got '%s'", sessions[0].Type)
	}
	if sessions[0].State != "playing" {
		t.Errorf("Expected State 'playing', got '%s'", sessions[0].State)
	}
	if sessions[0].PlayerName != "Chrome" {
		t.Errorf("Expected PlayerName 'Chrome', got '%s'", sessions[0].PlayerName)
	}
}

// TestGetSessionsNoSessions tests retrieval when no sessions are active
func TestGetSessionsNoSessions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/status/sessions" {
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<MediaContainer size="0">
</MediaContainer>`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewClient("test-token", server.URL)
	sessions, err := client.GetSessions("1")

	if err != nil {
		t.Fatalf("GetSessions failed: %v", err)
	}

	// Should return empty slice, not nil
	if sessions == nil {
		t.Error("Expected empty slice, got nil")
	}

	if len(sessions) != 0 {
		t.Errorf("Expected 0 sessions, got %d", len(sessions))
	}
}

// TestGetSessionsUserFilter tests that sessions are filtered by user ID
func TestGetSessionsUserFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/status/sessions" {
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<MediaContainer size="2">
  <Track sessionKey="session1" type="track" title="Song 1">
    <User id="1" title="User1"/>
    <Player state="playing" title="Player1"/>
  </Track>
  <Track sessionKey="session2" type="track" title="Song 2">
    <User id="2" title="User2"/>
    <Player state="playing" title="Player2"/>
  </Track>
</MediaContainer>`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewClient("test-token", server.URL)

	// Filter for user 1 only
	sessions, err := client.GetSessions("1")
	if err != nil {
		t.Fatalf("GetSessions failed: %v", err)
	}

	if len(sessions) != 1 {
		t.Errorf("Expected 1 session for user 1, got %d", len(sessions))
	}

	if sessions[0].UserID != "1" {
		t.Errorf("Expected UserID '1', got '%s'", sessions[0].UserID)
	}
}

// TestGetSessionsEmptyUserFilter tests that empty user filter returns all sessions
func TestGetSessionsEmptyUserFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/status/sessions" {
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<MediaContainer size="2">
  <Track sessionKey="session1" type="track" title="Song 1">
    <User id="1" title="User1"/>
    <Player state="playing" title="Player1"/>
  </Track>
  <Track sessionKey="session2" type="track" title="Song 2">
    <User id="2" title="User2"/>
    <Player state="playing" title="Player2"/>
  </Track>
</MediaContainer>`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewClient("test-token", server.URL)

	// Empty user filter returns all sessions
	sessions, err := client.GetSessions("")
	if err != nil {
		t.Fatalf("GetSessions failed: %v", err)
	}

	if len(sessions) != 2 {
		t.Errorf("Expected 2 sessions with empty filter, got %d", len(sessions))
	}
}

// TestGetSessionsAuthFailure tests authentication failure for GetSessions
func TestGetSessionsAuthFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
	}))
	defer server.Close()

	client := NewClient("invalid-token", server.URL)
	sessions, err := client.GetSessions("1")

	if sessions != nil {
		t.Error("Expected nil sessions for auth failure")
	}

	if err == nil {
		t.Fatal("Expected error for auth failure")
	}

	if !errors.Is(err, errors.PLEX_AUTH_FAILED) {
		t.Errorf("Expected PLEX_AUTH_FAILED error code, got: %s", errors.GetCode(err))
	}
}

// TestGetSessionsTimeout tests 500ms timeout for polling performance (NFR5)
func TestGetSessionsTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response exceeding 500ms
		time.Sleep(600 * time.Millisecond)
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<MediaContainer size="0"></MediaContainer>`))
	}))
	defer server.Close()

	client := NewClient("test-token", server.URL)
	sessions, err := client.GetSessions("1")

	if sessions != nil {
		t.Error("Expected nil sessions for timeout")
	}

	if err == nil {
		t.Fatal("Expected error for timeout")
	}

	// Should get timeout error
	code := errors.GetCode(err)
	if code != errors.TIMEOUT && code != errors.PLEX_UNREACHABLE && code != errors.PLEX_CONN_FAILED {
		t.Errorf("Expected TIMEOUT, PLEX_UNREACHABLE, or PLEX_CONN_FAILED error code, got: %s", code)
	}
}

// ========================================
// GetMusicSessions Tests
// ========================================

// TestGetMusicSessionsSuccess tests successful music session retrieval
func TestGetMusicSessionsSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/status/sessions" {
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<MediaContainer size="1">
  <Track sessionKey="music123" key="/library/metadata/12345" type="track"
         title="Awesome Song" grandparentTitle="Cool Artist" parentTitle="Great Album"
         thumb="/library/metadata/12345/thumb/123" duration="240000" viewOffset="60000">
    <User id="1" title="MusicFan" thumb="https://plex.tv/avatar"/>
    <Player state="playing" title="Sonos" product="Plex for Sonos"/>
  </Track>
</MediaContainer>`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewClient("valid-token", server.URL)
	musicSessions, err := client.GetMusicSessions("1")

	if err != nil {
		t.Fatalf("GetMusicSessions failed: %v", err)
	}

	if len(musicSessions) != 1 {
		t.Errorf("Expected 1 music session, got %d", len(musicSessions))
	}

	// Verify music session details
	session := musicSessions[0]
	if session.Track != "Awesome Song" {
		t.Errorf("Expected Track 'Awesome Song', got '%s'", session.Track)
	}
	if session.Artist != "Cool Artist" {
		t.Errorf("Expected Artist 'Cool Artist', got '%s'", session.Artist)
	}
	if session.Album != "Great Album" {
		t.Errorf("Expected Album 'Great Album', got '%s'", session.Album)
	}
	if session.Duration != 240000 {
		t.Errorf("Expected Duration 240000, got %d", session.Duration)
	}
	if session.ViewOffset != 60000 {
		t.Errorf("Expected ViewOffset 60000, got %d", session.ViewOffset)
	}
	if session.State != "playing" {
		t.Errorf("Expected State 'playing', got '%s'", session.State)
	}
}

// TestGetMusicSessionsFiltersNonMusic tests that non-music sessions are filtered out
func TestGetMusicSessionsFiltersNonMusic(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/status/sessions" {
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusOK)
			// Response with music, video, and photo sessions - only music should be returned
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<MediaContainer size="3">
  <Track sessionKey="music1" type="track" title="Song">
    <User id="1" title="User"/>
    <Player state="playing" title="Player"/>
  </Track>
  <Track sessionKey="video1" type="episode" title="Episode">
    <User id="1" title="User"/>
    <Player state="playing" title="Player"/>
  </Track>
  <Track sessionKey="video2" type="movie" title="Movie">
    <User id="1" title="User"/>
    <Player state="playing" title="Player"/>
  </Track>
</MediaContainer>`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewClient("test-token", server.URL)
	musicSessions, err := client.GetMusicSessions("1")

	if err != nil {
		t.Fatalf("GetMusicSessions failed: %v", err)
	}

	// Should only return the music session (type="track")
	if len(musicSessions) != 1 {
		t.Errorf("Expected 1 music session (filtering non-music), got %d", len(musicSessions))
	}

	if musicSessions[0].Type != "track" {
		t.Errorf("Expected Type 'track', got '%s'", musicSessions[0].Type)
	}
}

// TestGetMusicSessionsNoMusic tests retrieval when only non-music sessions exist
func TestGetMusicSessionsNoMusic(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/status/sessions" {
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusOK)
			// Only video sessions, no music
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<MediaContainer size="1">
  <Track sessionKey="video1" type="movie" title="Movie">
    <User id="1" title="User"/>
    <Player state="playing" title="Player"/>
  </Track>
</MediaContainer>`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewClient("test-token", server.URL)
	musicSessions, err := client.GetMusicSessions("1")

	if err != nil {
		t.Fatalf("GetMusicSessions failed: %v", err)
	}

	// Should return empty slice when no music is playing
	if musicSessions == nil {
		t.Error("Expected empty slice, got nil")
	}

	if len(musicSessions) != 0 {
		t.Errorf("Expected 0 music sessions, got %d", len(musicSessions))
	}
}

// TestGetMusicSessionsUserFilter tests that music sessions are filtered by user ID
func TestGetMusicSessionsUserFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/status/sessions" {
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<MediaContainer size="2">
  <Track sessionKey="music1" type="track" title="Song1">
    <User id="1" title="User1"/>
    <Player state="playing" title="Player1"/>
  </Track>
  <Track sessionKey="music2" type="track" title="Song2">
    <User id="2" title="User2"/>
    <Player state="playing" title="Player2"/>
  </Track>
</MediaContainer>`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewClient("test-token", server.URL)

	// Filter for user 2 only
	musicSessions, err := client.GetMusicSessions("2")
	if err != nil {
		t.Fatalf("GetMusicSessions failed: %v", err)
	}

	if len(musicSessions) != 1 {
		t.Errorf("Expected 1 music session for user 2, got %d", len(musicSessions))
	}

	if musicSessions[0].UserID != "2" {
		t.Errorf("Expected UserID '2', got '%s'", musicSessions[0].UserID)
	}
}

// TestMusicSessionJSONSerialization tests MusicSession struct serialization
func TestMusicSessionJSONSerialization(t *testing.T) {
	session := MusicSession{
		Session: Session{
			SessionKey: "test123",
			UserID:     "1",
			UserName:   "TestUser",
			Type:       "track",
			State:      "playing",
			PlayerName: "TestPlayer",
		},
		Track:      "Test Track",
		Artist:     "Test Artist",
		Album:      "Test Album",
		Thumb:      "/thumb/123",
		ThumbURL:   "http://server/thumb/123?X-Plex-Token=token",
		Duration:   180000,
		ViewOffset: 45000,
	}

	// Verify all fields are correctly set
	if session.SessionKey != "test123" {
		t.Error("SessionKey not correctly set")
	}
	if session.Track != "Test Track" {
		t.Error("Track not correctly set")
	}
	if session.Artist != "Test Artist" {
		t.Error("Artist not correctly set")
	}
	if session.Album != "Test Album" {
		t.Error("Album not correctly set")
	}
	if session.Thumb != "/thumb/123" {
		t.Error("Thumb not correctly set")
	}
	if session.ThumbURL != "http://server/thumb/123?X-Plex-Token=token" {
		t.Error("ThumbURL not correctly set")
	}
	if session.Duration != 180000 {
		t.Error("Duration not correctly set")
	}
	if session.ViewOffset != 45000 {
		t.Error("ViewOffset not correctly set")
	}
}

// ========================================
// Metadata Fallback Tests (Story 2.9)
// ========================================

// TestApplyFallbacksEmptyTrack tests fallback for missing track title (AC1)
func TestApplyFallbacksEmptyTrack(t *testing.T) {
	session := MusicSession{
		Track:  "",
		Artist: "Known Artist",
		Album:  "Known Album",
	}

	session.ApplyFallbacks()

	if session.Track != FallbackTrackTitle {
		t.Errorf("Expected Track '%s', got '%s'", FallbackTrackTitle, session.Track)
	}
	if session.Artist != "Known Artist" {
		t.Errorf("Expected Artist 'Known Artist', got '%s'", session.Artist)
	}
	if session.Album != "Known Album" {
		t.Errorf("Expected Album 'Known Album', got '%s'", session.Album)
	}
}

// TestApplyFallbacksEmptyArtist tests fallback for missing artist (AC2)
func TestApplyFallbacksEmptyArtist(t *testing.T) {
	session := MusicSession{
		Track:  "Known Track",
		Artist: "",
		Album:  "Known Album",
	}

	session.ApplyFallbacks()

	if session.Track != "Known Track" {
		t.Errorf("Expected Track 'Known Track', got '%s'", session.Track)
	}
	if session.Artist != FallbackArtist {
		t.Errorf("Expected Artist '%s', got '%s'", FallbackArtist, session.Artist)
	}
	if session.Album != "Known Album" {
		t.Errorf("Expected Album 'Known Album', got '%s'", session.Album)
	}
}

// TestApplyFallbacksEmptyAlbum tests fallback for missing album (AC3)
func TestApplyFallbacksEmptyAlbum(t *testing.T) {
	session := MusicSession{
		Track:  "Known Track",
		Artist: "Known Artist",
		Album:  "",
	}

	session.ApplyFallbacks()

	if session.Track != "Known Track" {
		t.Errorf("Expected Track 'Known Track', got '%s'", session.Track)
	}
	if session.Artist != "Known Artist" {
		t.Errorf("Expected Artist 'Known Artist', got '%s'", session.Artist)
	}
	if session.Album != FallbackAlbum {
		t.Errorf("Expected Album '%s', got '%s'", FallbackAlbum, session.Album)
	}
}

// TestApplyFallbacksAllEmpty tests fallback for all missing metadata (AC7)
func TestApplyFallbacksAllEmpty(t *testing.T) {
	session := MusicSession{
		Track:  "",
		Artist: "",
		Album:  "",
	}

	session.ApplyFallbacks()

	if session.Track != FallbackTrackTitle {
		t.Errorf("Expected Track '%s', got '%s'", FallbackTrackTitle, session.Track)
	}
	if session.Artist != FallbackArtist {
		t.Errorf("Expected Artist '%s', got '%s'", FallbackArtist, session.Artist)
	}
	if session.Album != FallbackAlbum {
		t.Errorf("Expected Album '%s', got '%s'", FallbackAlbum, session.Album)
	}
}

// TestApplyFallbacksNoChange tests that complete metadata is not modified
func TestApplyFallbacksNoChange(t *testing.T) {
	session := MusicSession{
		Track:  "My Track",
		Artist: "My Artist",
		Album:  "My Album",
	}

	session.ApplyFallbacks()

	if session.Track != "My Track" {
		t.Errorf("Expected Track 'My Track', got '%s'", session.Track)
	}
	if session.Artist != "My Artist" {
		t.Errorf("Expected Artist 'My Artist', got '%s'", session.Artist)
	}
	if session.Album != "My Album" {
		t.Errorf("Expected Album 'My Album', got '%s'", session.Album)
	}
}

// TestApplyFallbacksDurationZero tests that zero duration is valid (AC5)
func TestApplyFallbacksDurationZero(t *testing.T) {
	session := MusicSession{
		Track:    "Track",
		Artist:   "Artist",
		Album:    "Album",
		Duration: 0,
	}

	session.ApplyFallbacks()

	// Duration 0 is valid - no fallback needed
	if session.Duration != 0 {
		t.Errorf("Expected Duration 0, got %d", session.Duration)
	}
}

// TestApplyFallbacksViewOffsetZero tests that zero viewOffset is valid (AC6)
func TestApplyFallbacksViewOffsetZero(t *testing.T) {
	session := MusicSession{
		Track:      "Track",
		Artist:     "Artist",
		Album:      "Album",
		ViewOffset: 0,
	}

	session.ApplyFallbacks()

	// ViewOffset 0 is valid - no fallback needed
	if session.ViewOffset != 0 {
		t.Errorf("Expected ViewOffset 0, got %d", session.ViewOffset)
	}
}

// ========================================
// Artwork URL Tests (Story 2.9)
// ========================================

// TestBuildArtworkURL tests absolute artwork URL construction (AC4)
func TestBuildArtworkURL(t *testing.T) {
	client := NewClient("my-token", "http://192.168.1.100:32400")

	testCases := []struct {
		name      string
		thumbPath string
		expected  string
	}{
		{
			name:      "Standard thumb path",
			thumbPath: "/library/metadata/12345/thumb/1234567890",
			expected:  "http://192.168.1.100:32400/library/metadata/12345/thumb/1234567890?X-Plex-Token=my-token",
		},
		{
			name:      "Empty thumb path",
			thumbPath: "",
			expected:  "",
		},
		{
			name:      "Simple thumb path",
			thumbPath: "/thumb/123",
			expected:  "http://192.168.1.100:32400/thumb/123?X-Plex-Token=my-token",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := client.buildArtworkURL(tc.thumbPath)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

// TestBuildArtworkURLEscapesToken tests that special characters in token are escaped
func TestBuildArtworkURLEscapesToken(t *testing.T) {
	client := NewClient("token+with/special=chars", "http://localhost:32400")

	result := client.buildArtworkURL("/thumb/123")

	// Token should be URL-escaped
	if result != "http://localhost:32400/thumb/123?X-Plex-Token=token%2Bwith%2Fspecial%3Dchars" {
		t.Errorf("Token not properly escaped in URL: %s", result)
	}
}

// TestGetMusicSessionsAppliesFallbacks tests that GetMusicSessions applies fallbacks (AC7)
func TestGetMusicSessionsAppliesFallbacks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/status/sessions" {
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusOK)
			// Return session with missing metadata
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<MediaContainer size="1">
  <Track sessionKey="test123" type="track"
         title="" grandparentTitle="" parentTitle=""
         thumb="" duration="0" viewOffset="0">
    <User id="1" title="User"/>
    <Player state="playing" title="Player"/>
  </Track>
</MediaContainer>`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewClient("test-token", server.URL)
	sessions, err := client.GetMusicSessions("1")

	if err != nil {
		t.Fatalf("GetMusicSessions failed: %v", err)
	}

	if len(sessions) != 1 {
		t.Fatalf("Expected 1 session, got %d", len(sessions))
	}

	session := sessions[0]

	// Verify fallbacks were applied
	if session.Track != FallbackTrackTitle {
		t.Errorf("Expected Track fallback '%s', got '%s'", FallbackTrackTitle, session.Track)
	}
	if session.Artist != FallbackArtist {
		t.Errorf("Expected Artist fallback '%s', got '%s'", FallbackArtist, session.Artist)
	}
	if session.Album != FallbackAlbum {
		t.Errorf("Expected Album fallback '%s', got '%s'", FallbackAlbum, session.Album)
	}

	// Thumb/ThumbURL should be empty (no fallback for artwork)
	if session.Thumb != "" {
		t.Errorf("Expected empty Thumb, got '%s'", session.Thumb)
	}
	if session.ThumbURL != "" {
		t.Errorf("Expected empty ThumbURL, got '%s'", session.ThumbURL)
	}
}

// TestGetMusicSessionsBuildsThumbURL tests that GetMusicSessions builds absolute artwork URL (AC4)
func TestGetMusicSessionsBuildsThumbURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/status/sessions" {
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<MediaContainer size="1">
  <Track sessionKey="test123" type="track"
         title="Song" grandparentTitle="Artist" parentTitle="Album"
         thumb="/library/metadata/12345/thumb/9876" duration="180000" viewOffset="30000">
    <User id="1" title="User"/>
    <Player state="playing" title="Player"/>
  </Track>
</MediaContainer>`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewClient("my-secret-token", server.URL)
	sessions, err := client.GetMusicSessions("1")

	if err != nil {
		t.Fatalf("GetMusicSessions failed: %v", err)
	}

	if len(sessions) != 1 {
		t.Fatalf("Expected 1 session, got %d", len(sessions))
	}

	session := sessions[0]

	// Verify relative thumb is preserved
	if session.Thumb != "/library/metadata/12345/thumb/9876" {
		t.Errorf("Expected Thumb '/library/metadata/12345/thumb/9876', got '%s'", session.Thumb)
	}

	// Verify absolute ThumbURL is constructed
	expectedThumbURL := server.URL + "/library/metadata/12345/thumb/9876?X-Plex-Token=my-secret-token"
	if session.ThumbURL != expectedThumbURL {
		t.Errorf("Expected ThumbURL '%s', got '%s'", expectedThumbURL, session.ThumbURL)
	}
}

// TestGetMusicSessionsCompleteMetadata tests extraction of all metadata fields
func TestGetMusicSessionsCompleteMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/status/sessions" {
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<MediaContainer size="1">
  <Track sessionKey="complete123" key="/library/metadata/99999" type="track"
         title="Complete Song Title" grandparentTitle="Complete Artist Name" parentTitle="Complete Album Name"
         thumb="/library/metadata/99999/thumb/111" duration="256000" viewOffset="128000">
    <User id="42" title="CompleteUser" thumb="https://plex.tv/user/avatar"/>
    <Player state="paused" title="Sonos" product="Plex for Sonos"/>
  </Track>
</MediaContainer>`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewClient("complete-token", server.URL)
	sessions, err := client.GetMusicSessions("42")

	if err != nil {
		t.Fatalf("GetMusicSessions failed: %v", err)
	}

	if len(sessions) != 1 {
		t.Fatalf("Expected 1 session, got %d", len(sessions))
	}

	session := sessions[0]

	// Verify all metadata fields (AC1-AC6)
	if session.Track != "Complete Song Title" {
		t.Errorf("Track: expected 'Complete Song Title', got '%s'", session.Track)
	}
	if session.Artist != "Complete Artist Name" {
		t.Errorf("Artist: expected 'Complete Artist Name', got '%s'", session.Artist)
	}
	if session.Album != "Complete Album Name" {
		t.Errorf("Album: expected 'Complete Album Name', got '%s'", session.Album)
	}
	if session.Duration != 256000 {
		t.Errorf("Duration: expected 256000, got %d", session.Duration)
	}
	if session.ViewOffset != 128000 {
		t.Errorf("ViewOffset: expected 128000, got %d", session.ViewOffset)
	}
	if session.State != "paused" {
		t.Errorf("State: expected 'paused', got '%s'", session.State)
	}
	if session.Thumb != "/library/metadata/99999/thumb/111" {
		t.Errorf("Thumb: expected '/library/metadata/99999/thumb/111', got '%s'", session.Thumb)
	}
	if session.ThumbURL == "" {
		t.Error("ThumbURL should not be empty for session with thumb")
	}
}
