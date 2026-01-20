package plex

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

const (
	gdmMulticastAddr = "239.0.0.250:32414"
	gdmRequest       = "M-SEARCH * HTTP/1.0\r\n\r\n"
)

// GDMScanner performs Plex server discovery using GDM protocol
type GDMScanner struct {
	timeout time.Duration
}

// NewGDMScanner creates a new GDM scanner with specified timeout
func NewGDMScanner(timeout time.Duration) *GDMScanner {
	return &GDMScanner{timeout: timeout}
}

// Scan sends GDM discovery packet and listens for responses
func (s *GDMScanner) Scan() ([]Server, error) {
	// Resolve multicast address
	addr, err := net.ResolveUDPAddr("udp4", gdmMulticastAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve GDM address: %w", err)
	}

	// Create UDP connection
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		return nil, fmt.Errorf("failed to create UDP connection: %w", err)
	}
	defer conn.Close()

	// Send GDM discovery packet
	_, err = conn.WriteToUDP([]byte(gdmRequest), addr)
	if err != nil {
		return nil, fmt.Errorf("failed to send GDM packet: %w", err)
	}

	// Set read timeout
	conn.SetReadDeadline(time.Now().Add(s.timeout))

	// Collect responses
	servers := make([]Server, 0)
	buf := make([]byte, 1024)

	for {
		n, remoteAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			// Timeout reached, stop listening
			break
		}

		// Parse GDM response
		server, err := parseGDMResponse(buf[:n], remoteAddr.IP.String())
		if err != nil {
			continue // Skip malformed responses
		}

		servers = append(servers, server)
	}

	return deduplicateServers(servers), nil
}

// parseGDMResponse parses a GDM response packet
func parseGDMResponse(data []byte, ip string) (Server, error) {
	server := Server{
		Address: ip,
		IsLocal: isLocalIP(ip),
	}

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Name: ") {
			server.Name = strings.TrimPrefix(line, "Name: ")
		} else if strings.HasPrefix(line, "Port: ") {
			server.Port = strings.TrimPrefix(line, "Port: ")
		} else if strings.HasPrefix(line, "Resource-Identifier: ") {
			server.ID = strings.TrimPrefix(line, "Resource-Identifier: ")
		}
	}

	if server.Name == "" {
		return server, fmt.Errorf("invalid GDM response: missing Name")
	}

	return server, nil
}

// isLocalIP checks if an IP address is on a local network
func isLocalIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// Check for private IP ranges
	return parsedIP.IsPrivate() || parsedIP.IsLoopback()
}

// deduplicateServers removes duplicate servers based on ID
func deduplicateServers(servers []Server) []Server {
	seen := make(map[string]bool)
	result := make([]Server, 0)

	for _, server := range servers {
		if server.ID == "" {
			result = append(result, server)
			continue
		}

		if !seen[server.ID] {
			seen[server.ID] = true
			result = append(result, server)
		}
	}

	return result
}
