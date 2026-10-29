package discord

import (
	"errors"
	"testing"

	"plexcord/internal/discord/ipc"
)

type fakeConn struct {
	loginErr    error
	activityErr error
	logins      []string
	activities  []ipc.Activity
	closed      int
}

func (c *fakeConn) Login(clientID string) error {
	c.logins = append(c.logins, clientID)
	return c.loginErr
}

func (c *fakeConn) SetActivity(a ipc.Activity) error {
	c.activities = append(c.activities, a)
	return c.activityErr
}

func (c *fakeConn) Close() error {
	c.closed++
	return nil
}

func newFakeManager(conn *fakeConn) *PresenceManager {
	return NewPresenceManager(WithDialer(func() Conn { return conn }))
}

const testClientID = "1463211692656689172"

func TestPresenceManagerConnectsThroughDialer(t *testing.T) {
	conn := &fakeConn{}
	pm := newFakeManager(conn)

	if pm.IsConnected() {
		t.Fatal("a fresh manager reports itself connected")
	}
	if err := pm.Connect(testClientID); err != nil {
		t.Fatalf("Connect() error: %v", err)
	}
	if !pm.IsConnected() {
		t.Error("IsConnected() = false after a successful Connect")
	}
	if pm.GetClientID() != testClientID {
		t.Errorf("GetClientID() = %q, want %q", pm.GetClientID(), testClientID)
	}
	if len(conn.logins) != 1 || conn.logins[0] != testClientID {
		t.Errorf("logins = %v, want one login with the client ID", conn.logins)
	}
}

func TestPresenceManagerConnectFailureLeavesLinkDown(t *testing.T) {
	conn := &fakeConn{loginErr: errors.New("connection refused")}
	pm := newFakeManager(conn)

	if err := pm.Connect(testClientID); err == nil {
		t.Fatal("Connect() = nil despite a failing login")
	}
	if pm.IsConnected() {
		t.Error("IsConnected() = true after a failed login")
	}
}

func TestPresenceManagerUpdatePlaybackBuildsActivity(t *testing.T) {
	conn := &fakeConn{}
	pm := newFakeManager(conn)
	if err := pm.Connect(testClientID); err != nil {
		t.Fatalf("Connect() error: %v", err)
	}

	err := pm.UpdatePlayback(Playback{
		Track:    "Song",
		Artist:   "Artist",
		Album:    "Album",
		State:    "playing",
		Duration: 240000,
		Position: 30000,
	}, Options{})
	if err != nil {
		t.Fatalf("UpdatePlayback() error: %v", err)
	}

	if len(conn.activities) != 1 {
		t.Fatalf("activities sent = %d, want 1", len(conn.activities))
	}
	activity := conn.activities[0]
	if activity.Details != "Song" {
		t.Errorf("Details = %q, want the track title", activity.Details)
	}
	if activity.State != "by Artist • Album" {
		t.Errorf("State = %q, want the artist and album line", activity.State)
	}
	if activity.Timestamps == nil || activity.Timestamps.Start == nil || activity.Timestamps.End == nil {
		t.Errorf("timestamps = %+v, want a start and an end for a known duration", activity.Timestamps)
	}
}

func TestPresenceManagerUpdatePlaybackUnknownDuration(t *testing.T) {
	conn := &fakeConn{}
	pm := newFakeManager(conn)
	if err := pm.Connect(testClientID); err != nil {
		t.Fatalf("Connect() error: %v", err)
	}

	if err := pm.UpdatePlayback(Playback{Track: "Stream", State: "playing"}, Options{}); err != nil {
		t.Fatalf("UpdatePlayback() error: %v", err)
	}

	ts := conn.activities[0].Timestamps
	if ts == nil || ts.Start == nil {
		t.Fatalf("timestamps = %+v, want an elapsed timer", ts)
	}
	if ts.End != nil {
		t.Error("an unknown duration produced a progress-bar end timestamp")
	}
}

func TestPresenceManagerUpdatePlaybackRequiresConnection(t *testing.T) {
	pm := newFakeManager(&fakeConn{})

	if err := pm.UpdatePlayback(Playback{Track: "Song"}, Options{}); err == nil {
		t.Error("UpdatePlayback() = nil with no connection, want an error")
	}
}

func TestPresenceManagerClearSendsEmptyActivity(t *testing.T) {
	conn := &fakeConn{}
	pm := newFakeManager(conn)
	if err := pm.Connect(testClientID); err != nil {
		t.Fatalf("Connect() error: %v", err)
	}
	if err := pm.UpdatePlayback(Playback{Track: "Song", State: "playing"}, Options{}); err != nil {
		t.Fatalf("UpdatePlayback() error: %v", err)
	}

	if err := pm.ClearPresence(); err != nil {
		t.Fatalf("ClearPresence() error: %v", err)
	}

	if got := conn.activities[len(conn.activities)-1]; got.Details != "" || got.State != "" {
		t.Errorf("clear sent %+v, want an empty activity", got)
	}
	if !pm.IsConnected() {
		t.Error("ClearPresence() dropped the connection")
	}
	if pm.GetCurrentPresence() != nil {
		t.Error("ClearPresence() left presence data behind")
	}
}

func TestPresenceManagerReconnectOnClientIDChange(t *testing.T) {
	conn := &fakeConn{}
	pm := newFakeManager(conn)
	const otherClientID = "1463211692656689173"

	if err := pm.Connect(testClientID); err != nil {
		t.Fatalf("Connect() error: %v", err)
	}
	if err := pm.Connect(otherClientID); err != nil {
		t.Fatalf("Connect() with a new client ID: %v", err)
	}

	if conn.closed != 1 {
		t.Errorf("closes = %d, want the previous connection torn down once", conn.closed)
	}
	if pm.GetClientID() != otherClientID {
		t.Errorf("GetClientID() = %q, want the new client ID", pm.GetClientID())
	}
}

func TestPresenceManagerDisconnectIsIdempotent(t *testing.T) {
	conn := &fakeConn{}
	pm := newFakeManager(conn)

	if err := pm.Disconnect(); err != nil {
		t.Fatalf("Disconnect() on an unconnected manager: %v", err)
	}
	if conn.closed != 0 {
		t.Errorf("closes = %d, want 0 when never connected", conn.closed)
	}
}

func TestPresenceManagerClosesLostConnection(t *testing.T) {
	conn := &fakeConn{}
	pm := newFakeManager(conn)
	if err := pm.Connect(testClientID); err != nil {
		t.Fatalf("Connect() error: %v", err)
	}

	conn.activityErr = &ipc.ClosedError{}
	if err := pm.ClearPresence(); err == nil {
		t.Fatal("ClearPresence() on a dead socket returned nil")
	}
	if pm.IsConnected() {
		t.Error("IsConnected() = true after the connection was lost")
	}
	if conn.closed != 1 {
		t.Errorf("closed = %d, want 1 — the lost socket leaked", conn.closed)
	}

	conn.activityErr = nil
	if err := pm.Connect(testClientID); err != nil {
		t.Fatalf("reconnect error: %v", err)
	}
	if len(conn.logins) != 2 {
		t.Errorf("logins = %d, want 2 after reconnecting", len(conn.logins))
	}
}
