package ipc

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"time"
)

// responseTimeout bounds how long we wait for Discord to acknowledge a frame.
// Discord replies within milliseconds locally; this only guards against a hung
// socket so a presence update can never block a caller indefinitely.
const responseTimeout = 10 * time.Second

// errNotConnected is returned when an operation is attempted before Login.
var errNotConnected = errors.New("ipc: not connected")

// ClosedError indicates Discord sent a CLOSE (op 2) frame or the socket was
// torn down. Callers use it to detect a lost connection without string matching.
type ClosedError struct {
	Code    int
	Message string
}

func (e *ClosedError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("ipc: connection closed (%d): %s", e.Code, e.Message)
	}
	return "ipc: connection closed"
}

// ActivityError is an `evt: ERROR` response to a command (e.g. an invalid
// client id or malformed activity). The connection remains usable.
type ActivityError struct {
	Code    int
	Message string
}

func (e *ActivityError) Error() string {
	return fmt.Sprintf("ipc: activity rejected (%d): %s", e.Code, e.Message)
}

// Client is a single Discord IPC connection. It is not safe for concurrent use;
// callers (PresenceManager) serialize access with their own mutex.
type Client struct {
	conn net.Conn
	// dial opens the platform socket. Overridable in tests with a fake conn.
	dial func() (net.Conn, error)
}

// New returns a Client that connects to the local Discord IPC socket.
func New() *Client {
	return &Client{dial: dialDiscord}
}

// Login opens the IPC socket and performs the handshake for clientID. It
// returns an error if the socket cannot be opened or Discord rejects the
// handshake (e.g. an unknown client id).
func (c *Client) Login(clientID string) error {
	conn, err := c.dial()
	if err != nil {
		return err
	}
	c.conn = conn

	payload, err := json.Marshal(handshake{V: "1", ClientID: clientID})
	if err != nil {
		c.closeConn()
		return err
	}
	if err := c.send(opHandshake, payload); err != nil {
		c.closeConn()
		return err
	}
	// Discord replies with a DISPATCH/READY frame; surface a CLOSE (e.g. an
	// invalid client id) as an error so callers do not believe they connected.
	if err := c.readResponse(); err != nil {
		c.closeConn()
		return err
	}
	return nil
}

// SetActivity sends a SET_ACTIVITY command and parses the response, returning
// an *ActivityError if Discord rejected the payload or a *ClosedError if the
// connection dropped.
func (c *Client) SetActivity(a Activity) error {
	if c.conn == nil {
		return errNotConnected
	}
	payload, err := json.Marshal(frame{
		Cmd:   "SET_ACTIVITY",
		Args:  args{Pid: os.Getpid(), Activity: a.toPayload()},
		Nonce: nonce(),
	})
	if err != nil {
		return err
	}
	if err := c.send(opFrame, payload); err != nil {
		return err
	}
	return c.readResponse()
}

// Close tears down the connection. It is safe to call when not connected.
func (c *Client) Close() error {
	return c.closeConn()
}

func (c *Client) closeConn() error {
	if c.conn == nil {
		return nil
	}
	err := c.conn.Close()
	c.conn = nil
	return err
}

func (c *Client) send(op opcode, payload []byte) error {
	if c.conn == nil {
		return errNotConnected
	}
	_, err := c.conn.Write(encodeFrame(op, payload))
	return err
}

// readResponse reads and interprets a single response frame. A CLOSE frame
// yields a *ClosedError; an `evt: ERROR` frame yields an *ActivityError; any
// other frame (READY, a normal SET_ACTIVITY ack) is treated as success.
func (c *Client) readResponse() error {
	if c.conn == nil {
		return errNotConnected
	}
	_ = c.conn.SetReadDeadline(time.Now().Add(responseTimeout))
	defer func() { _ = c.conn.SetReadDeadline(time.Time{}) }()

	op, payload, err := readFrame(c.conn)
	if err != nil {
		return err
	}

	switch op {
	case opClose:
		var cl closePayload
		_ = json.Unmarshal(payload, &cl)
		return &ClosedError{Code: cl.Code, Message: cl.Message}
	case opFrame:
		var resp responseFrame
		if err := json.Unmarshal(payload, &resp); err != nil {
			// A frame we cannot parse is not fatal; Discord accepted the command.
			return nil
		}
		if resp.Evt == "ERROR" {
			return &ActivityError{Code: resp.Data.Code, Message: resp.Data.Message}
		}
		return nil
	default:
		return nil
	}
}
