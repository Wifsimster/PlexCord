package ipc

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"testing"
	"time"
)

func TestEncodeReadFrame_RoundTrip(t *testing.T) {
	payload := []byte(`{"cmd":"SET_ACTIVITY"}`)
	encoded := encodeFrame(opFrame, payload)

	// Header: opcode + length, both little-endian.
	if got := binary.LittleEndian.Uint32(encoded[0:4]); got != uint32(opFrame) {
		t.Errorf("opcode header = %d, want %d", got, opFrame)
	}
	if got := binary.LittleEndian.Uint32(encoded[4:8]); got != uint32(len(payload)) {
		t.Errorf("length header = %d, want %d", got, len(payload))
	}

	op, out, err := readFrame(bytes.NewReader(encoded))
	if err != nil {
		t.Fatalf("readFrame: %v", err)
	}
	if op != opFrame {
		t.Errorf("op = %d, want %d", op, opFrame)
	}
	if !bytes.Equal(out, payload) {
		t.Errorf("payload = %q, want %q", out, payload)
	}
}

func TestReadFrame_ShortHeader(t *testing.T) {
	if _, _, err := readFrame(bytes.NewReader([]byte{1, 2, 3})); err == nil {
		t.Error("expected error on truncated header")
	}
}

func TestReadFrame_RejectsHugeLength(t *testing.T) {
	header := make([]byte, 8)
	binary.LittleEndian.PutUint32(header[0:4], uint32(opFrame))
	binary.LittleEndian.PutUint32(header[4:8], maxFrameSize+1)
	if _, _, err := readFrame(bytes.NewReader(header)); err == nil {
		t.Error("expected error for oversized frame length")
	}
}

func TestNonce_Unique(t *testing.T) {
	a, b := nonce(), nonce()
	if a == "" || b == "" {
		t.Fatal("nonce returned empty string")
	}
	if a == b {
		t.Error("expected distinct nonces")
	}
}

func TestActivity_ToPayload_Timestamps(t *testing.T) {
	start := time.UnixMilli(1_000)
	end := time.UnixMilli(241_000)

	t.Run("start and end", func(t *testing.T) {
		p := Activity{Timestamps: &Timestamps{Start: &start, End: &end}}.toPayload()
		if p.Timestamps == nil || p.Timestamps.Start == nil || p.Timestamps.End == nil {
			t.Fatal("expected start and end timestamps")
		}
		if *p.Timestamps.Start != 1_000 || *p.Timestamps.End != 241_000 {
			t.Errorf("timestamps = %d/%d, want 1000/241000", *p.Timestamps.Start, *p.Timestamps.End)
		}
	})

	t.Run("start only", func(t *testing.T) {
		p := Activity{Timestamps: &Timestamps{Start: &start}}.toPayload()
		if p.Timestamps == nil || p.Timestamps.Start == nil {
			t.Fatal("expected start timestamp")
		}
		if p.Timestamps.End != nil {
			t.Error("expected no end timestamp")
		}
	})

	t.Run("none", func(t *testing.T) {
		if p := (Activity{}).toPayload(); p.Timestamps != nil {
			t.Error("expected no timestamps for empty activity")
		}
	})
}

func TestActivity_ToPayload_TypeAndDisplayEncoding(t *testing.T) {
	sd := StatusDisplayDetails
	a := Activity{Type: ActivityListening, StatusDisplayType: &sd, Details: "Song"}
	p := a.toPayload()

	raw, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded["type"] != float64(2) {
		t.Errorf("type = %v, want 2", decoded["type"])
	}
	if decoded["status_display_type"] != float64(2) {
		t.Errorf("status_display_type = %v, want 2", decoded["status_display_type"])
	}
}

func TestActivity_ToPayload_OmitsStatusDisplayWhenNil(t *testing.T) {
	raw, _ := json.Marshal((Activity{Type: ActivityPlaying}).toPayload())
	if bytes.Contains(raw, []byte("status_display_type")) {
		t.Errorf("status_display_type should be omitted when nil, got %s", raw)
	}
	// Type 0 is meaningful (Playing) and must always be present.
	if !bytes.Contains(raw, []byte(`"type":0`)) {
		t.Errorf("type should always be sent, got %s", raw)
	}
}
