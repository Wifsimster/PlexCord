package ipc

import "time"

type ActivityType int

const (
	ActivityPlaying   ActivityType = 0
	ActivityListening ActivityType = 2
	ActivityWatching  ActivityType = 3
)

type StatusDisplayType int

const (
	StatusDisplayName    StatusDisplayType = 0
	StatusDisplayState   StatusDisplayType = 1
	StatusDisplayDetails StatusDisplayType = 2
)

type Timestamps struct {
	Start *time.Time
	End   *time.Time
}

type Button struct {
	Label string
	URL   string
}

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

type responseFrame struct {
	Cmd   string `json:"cmd"`
	Evt   string `json:"evt"`
	Nonce string `json:"nonce"`
	Data  struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"data"`
}

type closePayload struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (a Activity) isZero() bool {
	return a.Type == 0 && a.StatusDisplayType == nil && a.Details == "" && a.State == "" &&
		a.LargeImage == "" && a.LargeText == "" && a.SmallImage == "" && a.SmallText == "" &&
		a.Timestamps == nil && len(a.Buttons) == 0
}

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
		p.Buttons = append(p.Buttons, payloadButton(b))
	}

	return p
}
