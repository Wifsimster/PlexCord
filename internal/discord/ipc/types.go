// Package ipc implements a minimal Discord IPC (Rich Presence) client.
//
// It replaces the unmaintained github.com/hugolgst/rich-go dependency so that
// PlexCord can send fields rich-go does not support — notably the activity
// `type` (Playing/Listening/Watching) and `status_display_type` — and so that
// SET_ACTIVITY responses are parsed instead of ignored.
//
// The wire protocol is intentionally small: an opcode-framed stream carrying a
// JSON handshake (op 0) followed by SET_ACTIVITY command frames (op 1). See
// https://discord.com/developers/docs/topics/rpc for the framing details.
package ipc

import "time"

// ActivityType is the Discord activity type sent in the presence payload.
// Discord IPC has supported Listening/Watching over RPC since mid-2024; older
// clients silently render unknown types as Playing.
type ActivityType int

const (
	// ActivityPlaying renders as "Playing <app>" (the classic game activity).
	ActivityPlaying ActivityType = 0
	// ActivityListening renders as "Listening to <app>" with a Spotify-style card.
	ActivityListening ActivityType = 2
	// ActivityWatching renders as "Watching <app>".
	ActivityWatching ActivityType = 3
)

// StatusDisplayType controls which line Discord surfaces in the member list.
type StatusDisplayType int

const (
	// StatusDisplayName shows the application name (e.g. "PlexCord").
	StatusDisplayName StatusDisplayType = 0
	// StatusDisplayState shows the activity state line (e.g. "by Def Leppard").
	StatusDisplayState StatusDisplayType = 1
	// StatusDisplayDetails shows the activity details line (e.g. the track name).
	StatusDisplayDetails StatusDisplayType = 2
)

// Timestamps holds the start and optional end of an activity. When both are
// present Discord renders a live progress bar (Spotify-style); with only a
// start it renders an elapsed timer.
type Timestamps struct {
	Start *time.Time
	End   *time.Time
}

// Button is a clickable button rendered on the activity card.
type Button struct {
	Label string
	URL   string
}

// Activity is the high-level presence payload callers build and hand to
// Client.SetActivity. It mirrors the fields PlexCord actually uses; the wire
// encoding lives in payloadActivity.
type Activity struct {
	Type              ActivityType
	StatusDisplayType *StatusDisplayType
	Details           string
	State             string
	LargeImage        string
	LargeText         string
	SmallImage        string
	SmallText         string
	Timestamps        *Timestamps
	Buttons           []Button
}

// ----------------------------------------------------------------------------
// Wire types (JSON encoded into command frames)
// ----------------------------------------------------------------------------

type handshake struct {
	V        string `json:"v"`
	ClientID string `json:"client_id"`
}

type frame struct {
	Cmd   string `json:"cmd"`
	Args  args   `json:"args"`
	Nonce string `json:"nonce"`
}

type args struct {
	Pid      int              `json:"pid"`
	Activity *payloadActivity `json:"activity"`
}

type payloadActivity struct {
	// Type is always sent (0 Playing is meaningful), so no omitempty.
	Type              int                `json:"type"`
	StatusDisplayType *int               `json:"status_display_type,omitempty"`
	Details           string             `json:"details,omitempty"`
	State             string             `json:"state,omitempty"`
	Assets            *payloadAssets     `json:"assets,omitempty"`
	Timestamps        *payloadTimestamps `json:"timestamps,omitempty"`
	Buttons           []payloadButton    `json:"buttons,omitempty"`
}

type payloadAssets struct {
	LargeImage string `json:"large_image,omitempty"`
	LargeText  string `json:"large_text,omitempty"`
	SmallImage string `json:"small_image,omitempty"`
	SmallText  string `json:"small_text,omitempty"`
}

type payloadTimestamps struct {
	Start *uint64 `json:"start,omitempty"`
	End   *uint64 `json:"end,omitempty"`
}

type payloadButton struct {
	Label string `json:"label,omitempty"`
	URL   string `json:"url,omitempty"`
}

// responseFrame is the subset of a Discord op-1 response we care about.
type responseFrame struct {
	Cmd   string `json:"cmd"`
	Evt   string `json:"evt"`
	Nonce string `json:"nonce"`
	Data  struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"data"`
}

// closePayload is the body of an op-2 CLOSE frame.
type closePayload struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// toPayload converts a high-level Activity into its wire representation.
// Empty asset/timestamp groups are omitted so Discord clears them.
func (a Activity) toPayload() *payloadActivity {
	p := &payloadActivity{
		Type:    int(a.Type),
		Details: a.Details,
		State:   a.State,
	}

	if a.StatusDisplayType != nil {
		v := int(*a.StatusDisplayType)
		p.StatusDisplayType = &v
	}

	if a.LargeImage != "" || a.LargeText != "" || a.SmallImage != "" || a.SmallText != "" {
		p.Assets = &payloadAssets{
			LargeImage: a.LargeImage,
			LargeText:  a.LargeText,
			SmallImage: a.SmallImage,
			SmallText:  a.SmallText,
		}
	}

	if a.Timestamps != nil && a.Timestamps.Start != nil {
		start := uint64(a.Timestamps.Start.UnixMilli())
		ts := &payloadTimestamps{Start: &start}
		if a.Timestamps.End != nil {
			end := uint64(a.Timestamps.End.UnixMilli())
			ts.End = &end
		}
		p.Timestamps = ts
	}

	for _, b := range a.Buttons {
		p.Buttons = append(p.Buttons, payloadButton{Label: b.Label, URL: b.URL})
	}

	return p
}
