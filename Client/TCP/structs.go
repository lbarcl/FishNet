package TCPClient

import (
	"context"
	"net"
)

type FrameFlags uint8

type Settings struct {
	Addr   net.TCPAddr
	UseTLS bool

	MaxFrameBytes        uint32
	MaxDecompressedBytes uint32
	ZipThreshold         uint32
}

type Client struct {
	settings Settings
	conn     net.Conn
	ctx      context.Context
}
