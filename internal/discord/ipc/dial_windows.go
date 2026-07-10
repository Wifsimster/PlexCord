//go:build windows

package ipc

import (
	"errors"
	"fmt"
	"net"
	"time"

	npipe "gopkg.in/natefinch/npipe.v2"
)

// dialTimeout bounds a single pipe dial attempt. npipe.DialTimeout is used
// because a plain Dial blocks for a very long time when Discord is not running.
const dialTimeout = 2 * time.Second

// dialDiscord connects to the first available discord-ipc-{0..9} named pipe.
func dialDiscord() (net.Conn, error) {
	var lastErr error
	for i := 0; i < 10; i++ {
		path := fmt.Sprintf(`\\.\pipe\discord-ipc-%d`, i)
		conn, err := npipe.DialTimeout(path, dialTimeout)
		if err == nil {
			return conn, nil
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = errors.New("no discord-ipc pipe found (is Discord running?)")
	}
	return nil, lastErr
}
