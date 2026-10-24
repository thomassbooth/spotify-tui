package librespot

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
)

// EventConn is a WebSocket connection to go-librespot's /events endpoint.
type EventConn struct {
	conn   net.Conn
	reader *bufio.Reader
}

// DialEvents connects to go-librespot's WebSocket event stream.
// addr should be in the form "localhost:3678".
func DialEvents(addr string) (*EventConn, error) {
	// Strip scheme if present
	addr = strings.TrimPrefix(addr, "http://")
	addr = strings.TrimPrefix(addr, "https://")

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	key := make([]byte, 16)
	if _, err := rand.Read(key); err != nil {
		conn.Close()
		return nil, err
	}
	wsKey := base64.StdEncoding.EncodeToString(key)

	req := fmt.Sprintf(
		"GET /events HTTP/1.1\r\n"+
			"Host: %s\r\n"+
			"Upgrade: websocket\r\n"+
			"Connection: Upgrade\r\n"+
			"Sec-WebSocket-Key: %s\r\n"+
			"Sec-WebSocket-Version: 13\r\n"+
			"\r\n",
		addr, wsKey,
	)
	if _, err := conn.Write([]byte(req)); err != nil {
		conn.Close()
		return nil, err
	}

	reader := bufio.NewReader(conn)
	resp, err := http.ReadResponse(reader, nil)
	if err != nil {
		conn.Close()
		return nil, err
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusSwitchingProtocols {
		conn.Close()
		return nil, fmt.Errorf("librespot events: expected 101 Switching Protocols, got %d", resp.StatusCode)
	}

	return &EventConn{conn: conn, reader: reader}, nil
}

// ReadEvent blocks until the next WebSocket event arrives.
func (c *EventConn) ReadEvent() (*Event, error) {
	for {
		b0, err := c.reader.ReadByte()
		if err != nil {
			return nil, err
		}
		b1, err := c.reader.ReadByte()
		if err != nil {
			return nil, err
		}

		opcode := b0 & 0x0F
		masked := (b1 & 0x80) != 0
		payloadLen := int64(b1 & 0x7F)

		switch payloadLen {
		case 126:
			var ext uint16
			if err := binary.Read(c.reader, binary.BigEndian, &ext); err != nil {
				return nil, err
			}
			payloadLen = int64(ext)
		case 127:
			var ext uint64
			if err := binary.Read(c.reader, binary.BigEndian, &ext); err != nil {
				return nil, err
			}
			payloadLen = int64(ext)
		}

		var maskKey [4]byte
		if masked {
			if _, err := io.ReadFull(c.reader, maskKey[:]); err != nil {
				return nil, err
			}
		}

		payload := make([]byte, payloadLen)
		if _, err := io.ReadFull(c.reader, payload); err != nil {
			return nil, err
		}
		if masked {
			for i := range payload {
				payload[i] ^= maskKey[i%4]
			}
		}

		switch opcode {
		case 0x8: // close
			return nil, io.EOF
		case 0x9: // ping — send pong
			c.conn.Write([]byte{0x8A, 0x00})
			continue
		case 0xA: // pong — ignore
			continue
		}

		var event Event
		if err := json.Unmarshal(payload, &event); err != nil {
			return nil, err
		}
		return &event, nil
	}
}

func (c *EventConn) Close() error {
	return c.conn.Close()
}
