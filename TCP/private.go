package tcp

import (
	"encoding/binary"
	"fmt"
	"io"
	"time"
)

func (s *Server) handleFrame(id string, flags FrameFlags, payload []byte) {
	s.mu.Lock()
	conn := s.cons[id]
	s.mu.Unlock()

	if conn == nil {
		s.onErrorFunc(id, fmt.Errorf("Connection not found for ID: %s", id))
		return
	}

	if hasFlag(flags, FlagGzip) {
		data, err := gunzipFrame(payload, s.settings.MaxDecompressedBytes)
		if err != nil {
			s.onErrorFunc(id, fmt.Errorf("Error gunzipping frame: %v", err))
			return
		}

		payload = data
	}

	if s.onDataFunc != nil {
		s.onDataFunc(id, payload)
	}
}

// [4 Bytes payload size][1 Byte flags][N Bytes payload]

func (s *Server) handleConnection(id string) {
	s.mu.Lock()
	conn := s.cons[id].con
	s.mu.Unlock()

	if conn == nil {
		s.onErrorFunc(id, fmt.Errorf("Connection not found for ID: %s", id))
		return
	}

	defer func() {
		if conn != nil {
			if err := conn.Close(); err != nil {
				s.onErrorFunc(id, fmt.Errorf("Error closing connection: %v", err))
			}
		}

		s.mu.Lock()
		delete(s.cons, id)
		s.mu.Unlock()

		if s.onDisconnectFunc != nil {
			s.onDisconnectFunc(id)
		}
	}()

	setDeadline := func() {
		if s.settings.Timeout != 0 && conn != nil {
			_ = conn.SetReadDeadline(time.Now().Add(time.Duration(s.settings.Timeout) * time.Second))
		}
	}
	setDeadline()

	for {
		buf := make([]byte, 5)
		_, err := io.ReadFull(conn, buf)
		if err != nil {
			s.onErrorFunc(id, err)
			return
		}
		setDeadline()

		payloadSize := binary.BigEndian.Uint32(buf[:4])
		if payloadSize > s.settings.MaxFrameBytes {
			err := fmt.Errorf("Payload size exceeds maximum allowed: %d", payloadSize)
			s.onErrorFunc(id, err)
			return
		}

		flags := FrameFlags(buf[4])

		payload := make([]byte, payloadSize)
		_, err = io.ReadFull(conn, payload)
		if err != nil {
			s.onErrorFunc(id, err)
			return
		}

		s.handleFrame(id, flags, payload)
	}
}

func (s *Server) sendFrame(id string, flags FrameFlags, payload []byte) error {
	conn := s.cons[id].con
	if conn == nil {
		return fmt.Errorf("connection not found for ID: %s", id)
	}

	if len(payload) > int(s.settings.MaxFrameBytes) {
		return fmt.Errorf("payload size exceeds maximum allowed: %d", len(payload))
	}

	frame := make([]byte, 5+len(payload))
	binary.BigEndian.PutUint32(frame[:4], uint32(len(payload)))
	frame[4] = byte(flags)

	copy(frame[5:], payload)
	s.cons[id].wmu.Lock()
	defer s.cons[id].wmu.Unlock()

	if _, err := conn.Write(frame); err != nil {
		return err
	}

	return nil
}
