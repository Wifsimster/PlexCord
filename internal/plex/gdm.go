package plex

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

const (
	gdmMulticastAddr = "239.0.0.250:32414"
	gdmRequest       = "M-SEARCH * HTTP/1.0\r\n\r\n"
)

type GDMScanner struct {
	timeout time.Duration
}

func NewGDMScanner(timeout time.Duration) *GDMScanner {
	return &GDMScanner{timeout: timeout}
}

func (s *GDMScanner) Scan() ([]Server, error) {
	addr, err := net.ResolveUDPAddr("udp4", gdmMulticastAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve GDM address: %w", err)
	}

	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		return nil, fmt.Errorf("failed to create UDP connection: %w", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Warning: Failed to close UDP connection: %v", err)
		}
	}()

	_, err = conn.WriteToUDP([]byte(gdmRequest), addr)
	if err != nil {
		return nil, fmt.Errorf("failed to send GDM packet: %w", err)
	}

	if err := conn.SetReadDeadline(time.Now().Add(s.timeout)); err != nil {
		return nil, fmt.Errorf("failed to set read deadline: %w", err)
	}

	servers := make([]Server, 0)
	buf := make([]byte, 1024)

	for {
		n, remoteAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			break
		}

		server, err := parseGDMResponse(buf[:n], remoteAddr.IP.String())
		if err != nil {
			continue
		}

		servers = append(servers, server)
	}

	return deduplicateServers(servers), nil
}

func parseGDMResponse(data []byte, ip string) (Server, error) {
	server := Server{
		Address: ip,
		IsLocal: isLocalIP(ip),
	}

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "Name: "):
			server.Name = strings.TrimPrefix(line, "Name: ")
		case strings.HasPrefix(line, "Port: "):
			server.Port = strings.TrimPrefix(line, "Port: ")
		case strings.HasPrefix(line, "Resource-Identifier: "):
			server.ID = strings.TrimPrefix(line, "Resource-Identifier: ")
		}
	}

	if server.Name == "" {
		return server, fmt.Errorf("invalid GDM response: missing Name")
	}

	if server.Port == "" {
		server.Port = "32400"
	}

	return server, nil
}

func isLocalIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	return parsedIP.IsPrivate() || parsedIP.IsLoopback()
}

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
