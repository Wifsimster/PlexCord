package ipc

import (
	"encoding/json"
	"errors"
	"net"
	"testing"
	"time"
)

// fakeDiscord runs an in-process Discord IPC server on one end of a net.Pipe.
// handle receives each decoded command frame and returns the (opcode, payload)
// to reply with. Returning a nil payload replies with an empty SET_ACTIVITY ack.
type fakeDiscord struct {
	server net.Conn
	client net.Conn
}

func newFakeDiscord(t *testing.T, handle func(op opcode, payload []byte) (opcode, []byte)) *fakeDiscord {
	t.Helper()
	server, client := net.Pipe()
	f := &fakeDiscord{server: server, client: client}

	go func() {
		for {
			op, payload, err := readFrame(server)
			if err != nil {
				return
			}
			respOp, respPayload := handle(op, payload)
			if respPayload == nil {
				respPayload = []byte(`{"cmd":"SET_ACTIVITY","evt":null}`)
			}
			if _, err := server.Write(encodeFrame(respOp, respPayload)); err != nil {
				return
			}
		}
	}()

	t.Cleanup(func() {
		_ = server.Close()
		_ = client.Close()
	})
	return f
}

// newClient wires a Client to the fake server's pipe end.
func (f *fakeDiscord) newClient() *Client {
	return &Client{dial: func() (net.Conn, error) { return f.client, nil }}
}

func readyFrame() []byte {
	return []byte(`{"cmd":"DISPATCH","evt":"READY","data":{}}`)
}

func TestClient_LoginHandshakeAndSetActivity(t *testing.T) {
	var gotHandshake handshake
	var gotFrame frame

	f := newFakeDiscord(t, func(op opcode, payload []byte) (opcode, []byte) {
		switch op {
		case opHandshake:
			_ = json.Unmarshal(payload, &gotHandshake)
			return opFrame, readyFrame()
		default:
			_ = json.Unmarshal(payload, &gotFrame)
			return opFrame, nil // default ack
		}
	})

	c := f.newClient()
	if err := c.Login("123456789012345678"); err != nil {
		t.Fatalf("Login: %v", err)
	}
	if gotHandshake.ClientID != "123456789012345678" {
		t.Errorf("handshake client id = %q", gotHandshake.ClientID)
	}
	if gotHandshake.V != "1" {
		t.Errorf("handshake version = %q, want 1", gotHandshake.V)
	}

	sd := StatusDisplayDetails
	err := c.SetActivity(Activity{
		Type:              ActivityListening,
		StatusDisplayType: &sd,
		Details:           "Bohemian Rhapsody",
		State:             "by Queen",
	})
	if err != nil {
		t.Fatalf("SetActivity: %v", err)
	}
	if gotFrame.Cmd != "SET_ACTIVITY" {
		t.Errorf("cmd = %q, want SET_ACTIVITY", gotFrame.Cmd)
	}
	if gotFrame.Args.Activity == nil {
		t.Fatal("activity payload missing")
	}
	if gotFrame.Args.Activity.Type != 2 {
		t.Errorf("activity type = %d, want 2", gotFrame.Args.Activity.Type)
	}
	if gotFrame.Args.Activity.StatusDisplayType == nil || *gotFrame.Args.Activity.StatusDisplayType != 2 {
		t.Errorf("status_display_type not encoded as 2")
	}
	if gotFrame.Nonce == "" {
		t.Error("expected a nonce on the command frame")
	}
}

func TestClient_SetActivity_SurfacesErrorEvt(t *testing.T) {
	f := newFakeDiscord(t, func(op opcode, payload []byte) (opcode, []byte) {
		if op == opHandshake {
			return opFrame, readyFrame()
		}
		return opFrame, []byte(`{"cmd":"SET_ACTIVITY","evt":"ERROR","data":{"code":4000,"message":"invalid activity"}}`)
	})

	c := f.newClient()
	if err := c.Login("123456789012345678"); err != nil {
		t.Fatalf("Login: %v", err)
	}

	err := c.SetActivity(Activity{Details: "x"})
	var actErr *ActivityError
	if !errors.As(err, &actErr) {
		t.Fatalf("expected *ActivityError, got %v", err)
	}
	if actErr.Code != 4000 || actErr.Message != "invalid activity" {
		t.Errorf("unexpected activity error: %+v", actErr)
	}
}

func TestClient_Login_SurfacesCloseFrame(t *testing.T) {
	f := newFakeDiscord(t, func(op opcode, payload []byte) (opcode, []byte) {
		// Reject the handshake with a CLOSE frame (e.g. invalid client id).
		return opClose, []byte(`{"code":4000,"message":"Invalid Client ID"}`)
	})

	c := f.newClient()
	err := c.Login("999999999999999999")
	var closeErr *ClosedError
	if !errors.As(err, &closeErr) {
		t.Fatalf("expected *ClosedError, got %v", err)
	}
	if closeErr.Message != "Invalid Client ID" {
		t.Errorf("close message = %q", closeErr.Message)
	}
	if c.conn != nil {
		t.Error("connection should be closed after a failed login")
	}
}

func TestClient_SetActivity_NotConnected(t *testing.T) {
	c := New()
	if err := c.SetActivity(Activity{}); !errors.Is(err, errNotConnected) {
		t.Errorf("expected errNotConnected, got %v", err)
	}
}

func TestClient_Close_Idempotent(t *testing.T) {
	c := New()
	if err := c.Close(); err != nil {
		t.Errorf("Close on fresh client: %v", err)
	}
}

func TestClient_ReadResponse_HonorsDeadline(t *testing.T) {
	// A server that never replies must not hang readResponse forever: the
	// read deadline turns a silent socket into an error the caller can handle.
	server, client := net.Pipe()
	t.Cleanup(func() { _ = server.Close(); _ = client.Close() })

	_ = client.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
	if _, _, err := readFrame(client); err == nil {
		t.Error("expected a read timeout error from a silent socket")
	}
}
