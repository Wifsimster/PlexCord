package ipc

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
)

// opcode identifies the kind of an IPC frame.
type opcode uint32

const (
	opHandshake opcode = 0
	opFrame     opcode = 1
	opClose     opcode = 2
	opPing      opcode = 3
	opPong      opcode = 4
)

// maxFrameSize bounds the payload length accepted from the socket so a
// corrupt or malicious length header cannot trigger a huge allocation.
const maxFrameSize = 64 * 1024

// encodeFrame builds a Discord IPC frame: a little-endian opcode and payload
// length header followed by the raw payload bytes.
func encodeFrame(op opcode, payload []byte) []byte {
	buf := make([]byte, 8+len(payload))
	binary.LittleEndian.PutUint32(buf[0:4], uint32(op))
	binary.LittleEndian.PutUint32(buf[4:8], uint32(len(payload)))
	copy(buf[8:], payload)
	return buf
}

// readFrame reads a single framed message from r and returns its opcode and
// payload. It returns an error on a short read or an implausibly large frame.
func readFrame(r io.Reader) (opcode, []byte, error) {
	header := make([]byte, 8)
	if _, err := io.ReadFull(r, header); err != nil {
		return 0, nil, err
	}
	op := opcode(binary.LittleEndian.Uint32(header[0:4]))
	length := binary.LittleEndian.Uint32(header[4:8])
	if length > maxFrameSize {
		return 0, nil, fmt.Errorf("ipc: frame too large (%d bytes)", length)
	}
	payload := make([]byte, length)
	if _, err := io.ReadFull(r, payload); err != nil {
		return 0, nil, err
	}
	return op, payload, nil
}

// nonce returns a random UUID-v4-shaped string used to correlate a command
// frame with its response.
func nonce() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		// crypto/rand failing is catastrophic; a zero nonce is still a valid
		// (if non-unique) correlation id, which is acceptable for presence.
		return "00000000-0000-4000-0000-000000000000"
	}
	buf[6] = (buf[6] & 0x0f) | 0x40 // version 4
	buf[8] = (buf[8] & 0x3f) | 0x80 // variant 10
	return fmt.Sprintf("%x-%x-%x-%x-%x", buf[0:4], buf[4:6], buf[6:8], buf[8:10], buf[10:])
}
