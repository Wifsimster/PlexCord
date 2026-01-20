package plex

import (
	"time"

	"plexcord/internal/errors"
)

// DiscoverServers performs Plex server discovery using GDM protocol
func DiscoverServers(timeout time.Duration) ([]Server, error) {
	scanner := NewGDMScanner(timeout)

	servers, err := scanner.Scan()
	if err != nil {
		return nil, errors.Wrap(err, errors.PLEX_UNREACHABLE, "GDM discovery failed")
	}

	return servers, nil
}
