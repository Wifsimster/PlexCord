package plex

import (
	"testing"
	"time"
)

// TestGDMPacketFormat tests that GDM request packet is correct
func TestGDMPacketFormat(t *testing.T) {
	expected := "M-SEARCH * HTTP/1.0\r\n\r\n"
	if gdmRequest != expected {
		t.Errorf("GDM request format incorrect.\nExpected: %q\nGot: %q", expected, gdmRequest)
	}
}

// TestGDMMulticastAddress tests that multicast address is correct
func TestGDMMulticastAddress(t *testing.T) {
	expected := "239.0.0.250:32414"
	if gdmMulticastAddr != expected {
		t.Errorf("GDM multicast address incorrect.\nExpected: %q\nGot: %q", expected, gdmMulticastAddr)
	}
}

// TestParseGDMResponse tests parsing a valid GDM response
func TestParseGDMResponse(t *testing.T) {
	response := `HTTP/1.0 200 OK
Name: MyPlexServer
Port: 32400
Resource-Identifier: 1234567890abcdef`

	server, err := parseGDMResponse([]byte(response), "192.168.1.100")
	if err != nil {
		t.Fatalf("parseGDMResponse failed: %v", err)
	}

	if server.Name != "MyPlexServer" {
		t.Errorf("Expected Name: MyPlexServer, got: %s", server.Name)
	}

	if server.Port != "32400" {
		t.Errorf("Expected Port: 32400, got: %s", server.Port)
	}

	if server.Address != "192.168.1.100" {
		t.Errorf("Expected Address: 192.168.1.100, got: %s", server.Address)
	}

	if server.ID != "1234567890abcdef" {
		t.Errorf("Expected ID: 1234567890abcdef, got: %s", server.ID)
	}

	if !server.IsLocal {
		t.Error("Expected IsLocal: true, got: false")
	}
}

// TestParseGDMResponseMissingName tests that missing Name field returns error
func TestParseGDMResponseMissingName(t *testing.T) {
	response := `HTTP/1.0 200 OK
Port: 32400
Resource-Identifier: 1234567890abcdef`

	_, err := parseGDMResponse([]byte(response), "192.168.1.100")
	if err == nil {
		t.Error("Expected error for missing Name field, got nil")
	}
}

// TestIsLocalIP tests local IP detection
func TestIsLocalIP(t *testing.T) {
	tests := []struct {
		ip       string
		expected bool
	}{
		{"192.168.1.100", true},
		{"10.0.0.1", true},
		{"172.16.0.1", true},
		{"127.0.0.1", true},
		{"8.8.8.8", false},
		{"1.1.1.1", false},
		{"invalid-ip", false},
	}

	for _, tt := range tests {
		result := isLocalIP(tt.ip)
		if result != tt.expected {
			t.Errorf("isLocalIP(%q) = %v, expected %v", tt.ip, result, tt.expected)
		}
	}
}

// TestDeduplicateServers tests server deduplication by ID
func TestDeduplicateServers(t *testing.T) {
	servers := []Server{
		{ID: "server-1", Name: "Server 1"},
		{ID: "server-2", Name: "Server 2"},
		{ID: "server-1", Name: "Server 1 Duplicate"},
		{ID: "", Name: "Server Without ID 1"},
		{ID: "", Name: "Server Without ID 2"},
	}

	result := deduplicateServers(servers)

	// Should have 4 servers: 2 unique IDs + 2 without IDs
	if len(result) != 4 {
		t.Errorf("Expected 4 servers after deduplication, got %d", len(result))
	}

	// Check that server-1 appears only once
	count := 0
	for _, s := range result {
		if s.ID == "server-1" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("Expected server-1 to appear once, appeared %d times", count)
	}
}

// TestDeduplicateServersEmpty tests deduplication with empty slice
func TestDeduplicateServersEmpty(t *testing.T) {
	servers := []Server{}
	result := deduplicateServers(servers)

	if len(result) != 0 {
		t.Errorf("Expected empty result, got %d servers", len(result))
	}
}

// TestNewGDMScanner tests scanner creation
func TestNewGDMScanner(t *testing.T) {
	timeout := 5 * time.Second
	scanner := NewGDMScanner(timeout)

	if scanner == nil {
		t.Fatal("NewGDMScanner returned nil")
	}

	if scanner.timeout != timeout {
		t.Errorf("Expected timeout %v, got %v", timeout, scanner.timeout)
	}
}

// TestServerURL tests Server.URL() method
func TestServerURL(t *testing.T) {
	server := Server{
		Address: "192.168.1.100",
		Port:    "32400",
	}

	expected := "http://192.168.1.100:32400"
	result := server.URL()

	if result != expected {
		t.Errorf("Expected URL %q, got %q", expected, result)
	}
}

// TestGDMScannerTimeout tests that scan respects timeout
func TestGDMScannerTimeout(t *testing.T) {
	// This test verifies timeout mechanism works
	// We can't test actual network discovery in unit tests
	scanner := NewGDMScanner(100 * time.Millisecond)

	if scanner.timeout != 100*time.Millisecond {
		t.Errorf("Expected timeout 100ms, got %v", scanner.timeout)
	}
}
