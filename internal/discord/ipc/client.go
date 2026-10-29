package ipc

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"time"
)

const responseTimeout = 10 * time.Second

var errNotConnected = errors.New("ipc: not connected")

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

type ActivityError struct {
	Code    int
	Message string
}

func (e *ActivityError) Error() string {
	return fmt.Sprintf("ipc: activity rejected (%d): %s", e.Code, e.Message)
}

type Client struct {
	conn net.Conn
	dial func() (net.Conn, error)
}

func New() *Client {
	return &Client{dial: dialDiscord}
}

func (c *Client) Login(clientID string) error {
	conn, err := c.dial()
	if err != nil {
		return err
	}
	c.conn = conn

	if err := c.handshake(clientID); err != nil {
		return errors.Join(err, c.Close())
	}
	return nil
}

func (c *Client) handshake(clientID string) error {
	payload, err := json.Marshal(handshake{V: "1", ClientID: clientID})
	if err != nil {
		return err
	}
	if err := c.send(opHandshake, payload); err != nil {
		return err
	}
	return c.readResponse()
}

func (c *Client) SetActivity(a Activity) error {
	if c.conn == nil {
		return errNotConnected
	}
	var activity *payloadActivity
	if !a.isZero() {
		activity = a.toPayload()
	}
	payload, err := json.Marshal(frame{
		Cmd:   "SET_ACTIVITY",
		Args:  args{Pid: os.Getpid(), Activity: activity},
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

func (c *Client) Close() error {
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
	if len(payload) > maxFrameSize {
		return fmt.Errorf("ipc: payload too large (%d bytes)", len(payload))
	}
	_, err := c.conn.Write(encodeFrame(op, payload))
	return err
}

func (c *Client) readResponse() error {
	if c.conn == nil {
		return errNotConnected
	}
	if err := c.conn.SetReadDeadline(time.Now().Add(responseTimeout)); err != nil {
		return err
	}

	op, payload, err := readFrame(c.conn)
	if err != nil {
		return err
	}

	switch op {
	case opClose:
		var cl closePayload
		if err := json.Unmarshal(payload, &cl); err != nil {
			return &ClosedError{}
		}
		return &ClosedError{Code: cl.Code, Message: cl.Message}
	case opFrame:
		var resp responseFrame
		if err := json.Unmarshal(payload, &resp); err != nil {
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
