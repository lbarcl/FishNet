package tcp

import (
	"context"
	"crypto/tls"
	"net"
	"sync"
)

type FrameFlags uint8

type Connection struct {
	con net.Conn
	wmu sync.Mutex
}

type Settings struct {
	Addr                 net.TCPAddr
	UseTLS               bool
	Cert                 tls.Certificate
	Timeout              uint32
	MaxFrameBytes        uint32
	MaxDecompressedBytes uint32
	ZipThreshold         uint32
}

type Server struct {
	settings Settings
	listener net.Listener
	cons     map[string]*Connection
	mu       sync.Mutex
	cancel   context.CancelFunc

	onDataFunc       func(id string, payload []byte)
	onErrorFunc      func(id string, err error)
	onConnectFunc    func(id string)
	onDisconnectFunc func(id string)
}
