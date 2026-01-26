package TCPClient

import (
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"net"
)

const (
	FlagGzip FrameFlags = 1 << iota
	FlagTLS  FrameFlags = 1 << iota
)

func NewClient(settings Settings) (*Client, error) {
	conn, err := net.Dial("tcp", settings.Addr.String())
	if err != nil {
		return nil, err
	}

	var frames FrameFlags
	header := make([]byte, 5)
	binary.BigEndian.PutUint32(header[:4], 0)

	if settings.UseTLS {
		frames |= FlagTLS
		header[4] = byte(frames)

		if _, err := conn.Write(header); err != nil {
			return nil, fmt.Errorf("Error sending TLS signal: %v", err)
		}

		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
			MinVersion:         tls.VersionTLS12,
		}

		tlsConn := tls.Client(conn, tlsConfig)
		if err := tlsConn.Handshake(); err != nil {
			return nil, fmt.Errorf("TLS Handshake failed: %v", err)
		}

		fmt.Println("TLS Connection established successfully!")

		return &Client{
			settings: settings,
			conn:     tlsConn,
		}, nil
	}

	header[4] = byte(frames)

	if _, err := conn.Write(header); err != nil {
		return nil, fmt.Errorf("Error sending TLS signal: %v", err)
	}

	return &Client{
		settings: settings,
		conn:     conn,
	}, nil
}
