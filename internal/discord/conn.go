package discord

import "plexcord/internal/discord/ipc"

// Conn is the Discord IPC connection as the presence manager needs it: log in,
// push an activity, close. The concrete *ipc.Client satisfies it.
//
// PresenceManager depends on this interface rather than *ipc.Client so the
// connection lifecycle can be exercised without a running Discord (DIP), and so
// a different transport — a future websocket or a gateway relay — can be
// dropped in without touching presence logic. It stays at three methods: the
// exact surface presence management uses (ISP).
type Conn interface {
	// Login opens the connection and performs the handshake for clientID.
	Login(clientID string) error
	// SetActivity pushes an activity (the zero Activity clears the presence).
	SetActivity(a ipc.Activity) error
	// Close tears the connection down.
	Close() error
}

// Dialer constructs a not-yet-connected Conn. NewPresenceManager defaults to
// the real Discord IPC socket; WithDialer substitutes a fake.
type Dialer func() Conn

// defaultDialer returns a Conn backed by the local Discord IPC socket.
func defaultDialer() Conn { return ipc.New() }

// Compile-time assertion that the production IPC client satisfies the contract.
var _ Conn = (*ipc.Client)(nil)
