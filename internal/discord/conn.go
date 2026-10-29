package discord

import "plexcord/internal/discord/ipc"

type Conn interface {
	Login(clientID string) error
	SetActivity(a ipc.Activity) error
	Close() error
}

type Dialer func() Conn

func defaultDialer() Conn { return ipc.New() }

var _ Conn = (*ipc.Client)(nil)
