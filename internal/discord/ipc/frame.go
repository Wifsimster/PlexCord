package ipc

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
)

type opcode uint32

const (
	opHandshake opcode = 0
	opFrame     opcode = 1
	opClose     opcode = 2
	opPing      opcode = 3
	opPong      opcode = 4
)

const maxFrameSize = 64 * 1024

func encodeFrame(op opcode, payload []byte) []byte {
	n := len(payload)
	if n > maxFrameSize {
		n = maxFrameSize
		payload = payload[:n]
	}
	buf := make([]byte, 8+n)
	binary.LittleEndian.PutUint32(buf[0:4], uint32(op))
	binary.LittleEndian.PutUint32(buf[4:8], uint32(n))
	copy(buf[8:], payload)
	return buf
}

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

func nonce() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "00000000-0000-4000-0000-000000000000"
	}
	buf[6] = (buf[6] & 0x0f) | 0x40
	buf[8] = (buf[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", buf[0:4], buf[4:6], buf[6:8], buf[8:10], buf[10:])
}
