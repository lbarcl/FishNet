package main

import (
	tcp "fishnet/TCP"
	"fmt"
	"net"
)

func main() {
	serverSettings := tcp.Settings{
		Addr:                 net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080},
		UseTLS:               false,
		Timeout:              1000,
		MaxDecompressedBytes: 1024 * 1024,
		MaxFrameBytes:        1024 * 512,
	}

	tcpServer := tcp.NewServer(serverSettings)

	tcpServer.OnConnect(func(id string) {
		fmt.Println("New connection:", id)
	})

	for {
		tcpServer.Update()
	}
}
