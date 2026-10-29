//go:build !windows

package ipc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"
)

const dialTimeout = 2 * time.Second

func candidateDirs() []string {
	var bases []string
	for _, env := range []string{"XDG_RUNTIME_DIR", "TMPDIR", "TMP", "TEMP"} {
		if v, ok := os.LookupEnv(env); ok && v != "" {
			bases = append(bases, v)
		}
	}
	bases = append(bases, "/tmp")

	dirs := make([]string, 0, 4*len(bases))
	for _, b := range bases {
		dirs = append(dirs,
			b,
			filepath.Join(b, "snap.discord"),
			filepath.Join(b, "app", "com.discordapp.Discord"),
			filepath.Join(b, ".flatpak", "com.discordapp.Discord", "xdg-run"),
		)
	}
	return dirs
}

func dialDiscord() (net.Conn, error) {
	dialer := net.Dialer{Timeout: dialTimeout}
	ctx := context.Background()
	var lastErr error
	for _, dir := range candidateDirs() {
		for i := 0; i < 10; i++ {
			path := filepath.Join(dir, fmt.Sprintf("discord-ipc-%d", i))
			conn, err := dialer.DialContext(ctx, "unix", path)
			if err == nil {
				return conn, nil
			}
			lastErr = err
		}
	}
	if lastErr == nil {
		lastErr = errors.New("no discord-ipc socket found (is Discord running?)")
	}
	return nil, lastErr
}
