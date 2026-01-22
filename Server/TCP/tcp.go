package tcp

import (
	"crypto/tls"
	"fmt"
	"net"
)

const (
	FlagGzip FrameFlags = 1 << iota
)

func NewServer(settings Settings) (*Server, error) {
	s := &Server{
		settings: settings,
		cons:     make(map[string]*Connection),
	}

	var listener *net.TCPListener
	var err error

	listener, err = net.ListenTCP("TCP", &settings.Addr)
	if err != nil {
		return nil, err
	}

	if settings.UseTLS {
		config := &tls.Config{
			Certificates: []tls.Certificate{settings.Cert},
			MinVersion:   tls.VersionTLS12,
		}
		listener = tls.NewListener(listener, config).(*net.TCPListener)
	}

	s.listener = listener
	return s, nil
}

func (s *Server) Accept() (string, error) {
	conn, err := s.listener.AcceptTCP()
	if err != nil {
		return "", err
	}

	id, err := s.SetConnection(conn)
	if err != nil {
		return "", err
	}

	s.onConnectFunc(id)
	go s.handleConnection(id)
	return id, nil
}

func (s *Server) Send(id string, payload []byte) error {
	var flags FrameFlags
	if len(payload) > int(s.settings.ZipThreshold) {
		flags |= FlagGzip

		payload, err := gzipFrame(payload, s.settings.MaxFrameBytes)
		if err != nil {
			return err
		}

		if err := s.sendFrame(id, flags, payload); err != nil {
			return err
		}

		return nil
	}

	if err := s.sendFrame(id, flags, payload); err != nil {
		return err
	}

	return nil
}

func (s *Server) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, conn := range s.cons {
		conn.con.Close()
		delete(s.cons, id)
	}

	return s.listener.Close()
}

func (s *Server) GetConnection(id string) (*net.TCPConn, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	conn, ok := s.cons[id]
	if !ok {
		return nil, fmt.Errorf("Connection not found for ID: %s", id)
	}

	return &conn.con, nil
}

func (s *Server) SetConnection(conn *net.TCPConn) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id, _ := newUID()
	for _, ok := s.cons[id]; ok; {
		id, _ = newUID()
	}

	s.cons[id] = &Connection{
		con: conn,
	}
	return id, nil
}

func (s *Server) RemoveConnection(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	conn, ok := s.cons[id]
	if !ok {
		return fmt.Errorf("Connection not found for ID: %s", id)
	}

	conn.con.Close()
	delete(s.cons, id)
	return nil
}

func (s *Server) SetOnData(f func(id string, payload []byte)) {
	s.onDataFunc = f
}

func (s *Server) SetOnError(f func(id string, err error)) {
	s.onErrorFunc = f
}

func (s *Server) SetOnConnect(f func(id string)) {
	s.onConnectFunc = f
}

func (s *Server) SetOnDisconnect(f func(id string)) {
	s.onDisconnectFunc = f
}
