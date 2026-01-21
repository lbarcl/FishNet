package tcp

import (
	"crypto/tls"
	"net"
)

const (
	FlagGzip FrameFlags = 1 << iota
)

func NewServer(settings Settings) *Server {
	s := &Server{
		settings: settings,
		cons:     make(map[string]*Connection),
	}

	var listener net.Listener
	var err error

	listener, err = net.ListenTCP("TCP", &settings.Addr)
	if err != nil {
		s.onErrorFunc("server", err)
	}

	if settings.UseTLS {
		config := &tls.Config{
			Certificates: []tls.Certificate{settings.Cert},
			MinVersion:   tls.VersionTLS12,
		}
		listener = tls.NewListener(listener, config)
	}

	s.listener = listener
	return s
}

func (s *Server) Update() {
	conn, err := s.listener.Accept()
	if err != nil {
		s.onErrorFunc("server", err)
		return
	}

	id, _ := newUID()
	s.cons[id] = &Connection{
		con: conn,
	}

	if s.onConnectFunc != nil {
		s.onConnectFunc(id)
	}

	go s.handleConnection(id)
}

func (s *Server) Send(id string, payload []byte) {
	conn := s.cons[id].con
	if conn == nil {
		return
	}

	var flags FrameFlags
	if len(payload) > int(s.settings.ZipThreshold) {
		flags |= FlagGzip

		payload, err := gzipFrame(payload, s.settings.MaxFrameBytes)
		if err != nil {
			s.onErrorFunc(id, err)
			return
		}

		if err := s.sendFrame(id, flags, payload); err != nil {
			s.onErrorFunc(id, err)
		}

		return
	}

	if err := s.sendFrame(id, flags, payload); err != nil {
		s.onErrorFunc(id, err)
	}
}

func (s *Server) OnData(f func(id string, payload []byte)) {
	s.onDataFunc = f
}

func (s *Server) OnError(f func(id string, err error)) {
	s.onErrorFunc = f
}

func (s *Server) OnConnect(f func(id string)) {
	s.onConnectFunc = f
}

func (s *Server) OnDisconnect(f func(id string)) {
	s.onDisconnectFunc = f
}
