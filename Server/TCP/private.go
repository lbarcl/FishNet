package TCPServer

import (
	"encoding/binary"
	"fmt"
	"io"
	"time"
)

func (s *Server) handleFrame(id string, flags FrameFlags, payload []byte) {
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
	conn, err := s.GetConnection(id)
	if err != nil {
		s.onErrorFunc(id, err)
		return
	}

	// Centralized cleanup
	defer func() {
		s.RemoveConnection(id)
		if s.onDisconnectFunc != nil {
			s.onDisconnectFunc(id)
		}
	}()

	// Watcher with exit signal to prevent leaks
	watcherDone := make(chan struct{})
	defer close(watcherDone)
	go func() {
		select {
		case <-s.ctx.Done():
			s.RemoveConnection(id)
		case <-watcherDone:
		}
	}()

	setDeadline := func() {
		if s.settings.Timeout != 0 && conn != nil {
			_ = conn.SetReadDeadline(time.Now().Add(time.Duration(s.settings.Timeout) * time.Second))
		}
	}

	headerBuf := make([]byte, 5) // Allocated once outside the loop
	for {
		setDeadline()

		if _, err := io.ReadFull(conn, headerBuf); err != nil {
			if s.ctx.Err() == nil {
				s.onErrorFunc(id, err)
			}
			return
		}

		payloadSize := binary.BigEndian.Uint32(headerBuf[:4])
		if payloadSize > s.settings.MaxFrameBytes {
			s.onErrorFunc(id, fmt.Errorf("payload size %d exceeds max", payloadSize))
			return
		}

		flags := FrameFlags(headerBuf[4])
		payload := make([]byte, payloadSize) // Consider sync.Pool for large payloads
		if _, err := io.ReadFull(conn, payload); err != nil {
			if s.ctx.Err() == nil {
				s.onErrorFunc(id, err)
			}
			return
		}

		s.handleFrame(id, flags, payload)
	}
}

func (s *Server) sendFrame(id string, flags FrameFlags, payload []byte) error {
	conn, err := s.GetConnection(id)
	if err != nil {
		return err
	}

	if len(payload) > int(s.settings.MaxFrameBytes) {
		return fmt.Errorf("payload size exceeds maximum allowed: %d", len(payload))
	}

	frame := make([]byte, 5+len(payload))
	binary.BigEndian.PutUint32(frame[:4], uint32(len(payload)))
	frame[4] = byte(flags)

	copy(frame[5:], payload)
	s.sendLockTheID(id)
	defer s.sendUnlockTheID(id)

	if _, err := conn.Write(frame); err != nil {
		return err
	}

	return nil
}

func (s *Server) sendLockTheID(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if conn, ok := s.cons[id]; ok {
		conn.wmu.Lock()
	}
}

func (s *Server) sendUnlockTheID(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if conn, ok := s.cons[id]; ok {
		conn.wmu.Unlock()
	}
}

func (s *Server) getConnectionWrapper(id string) (*Connection, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	conn, ok := s.cons[id]
	if !ok {
		return nil, fmt.Errorf("Connection not found for ID: %s", id)
	}

	return conn, nil
}

func (s *Server) setEstablished(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if conn, ok := s.cons[id]; ok {
		conn.established = true
	}
}
