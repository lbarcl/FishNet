# FishNet

`fishnet-go` is a implementation FishNet networking protocol designed for efficient and secure TCP communication between clients and servers. It features built-in support for TLS encryption and automatic Gzip compression for large payloads.

## Features

* **TCP Communication**: Simple API for establishing TCP connections and sending data.
* **TLS Support**: Built-in support for encrypted communication with configurable certificate handling.
* **Automatic Compression**: Payloads exceeding a defined threshold are automatically compressed using Gzip to optimize bandwidth.
* **Efficient Buffering**: Utilizes `bytebufferpool` to reduce memory allocations and pressure on the garbage collector.
* **Asynchronous Event Handling**: Set custom callback functions for connection, data receipt, errors, and disconnection events.

## Installation

```bash
go get github.com/lbarcl/fishnet-go
```

## Getting Started

### Server Example

To create a server, define your `Settings` and initialize the server.

```go
package main

import (
    "net"
    "github.com/lbarcl/fishnet-go/server"
)

func main() {
    settings := server.Settings{
        Addr: net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080},
        UseTLS: false,
        ZipThreshold: 1024, // Compress payloads over 1KB
    }

    s, err := server.NewServer(settings)
    if err != nil {
        panic(err)
    }

    s.SetOnData(func(id string, payload []byte) {
        println("Received data from", id, ":", string(payload))
    })

    for {
        id, err := s.Accept()
        if err == nil {
            println("New connection established:", id)
        }
    }
}
```


### Client Example

The client can be initialized and connected to a server as follows:

```go
package main

import (
    "net"
    "github.com/lbarcl/fishnet-go/client"
)

func main() {
    settings := client.Settings{
        Addr: net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080},
        UseTLS: false,
        Timeout: 30,
    }

    c, err := client.NewClient(settings)
    if err != nil {
        panic(err)
    }

    c.SetOnConnect(func() {
        println("Connected to server!")
        c.Send([]byte("Hello, Server!"))
    })

    err = c.Connect()
    if err != nil {
        panic(err)
    }
}
```

### Configuration

Both the client and server use a `Settings` struct to tune connection behavior. While they share many fields, the server requires a certificate for TLS, and the client includes an option to trust unverified certificates.

#### Client Settings
```go
type Settings struct {
    Addr                 net.TCPAddr // The target server address
    UseTLS               bool        // Enables TLS encryption for the connection
    TrustUnverifiedCerts bool        // If true, skips verification for self-signed or invalid certificates
    Timeout              uint32      // Connection timeout and read/write deadline in seconds
    MaxFrameBytes        uint32      // Maximum allowed size for an incoming network frame
    MaxDecompressedBytes uint32      // Maximum allowed size for a payload after decompression
    ZipThreshold         uint32      // Payloads exceeding this byte size will be Gzip compressed
}
```

#### Server Settings
```go
type Settings struct {
    Addr                 net.TCPAddr     // The local address and port to bind the listener to
    UseTLS               bool            // Enables TLS encryption for the server
    Cert                 tls.Certificate // The certificate used for TLS (required if UseTLS is true)
    Timeout              uint32          // Connection timeout and read/write deadline in seconds
    MaxFrameBytes        uint32          // Maximum allowed size for an incoming network frame
    MaxDecompressedBytes uint32          // Maximum allowed size for a payload after decompression
    ZipThreshold         uint32          // Payloads exceeding this byte size will be Gzip compressed
}
```

### Key Parameter Details
* **TLS Configuration**: When `UseTLS` is enabled, the server must provide a valid `tls.Certificate`. The client can optionally set `TrustUnverifiedCerts` to `true` to connect to servers using self-signed keys.
* **Compression (`ZipThreshold`)**: To save bandwidth, FishNet automatically compresses any payload larger than this value.
* **Security Limits**: `MaxFrameBytes` and `MaxDecompressedBytes` act as safeguards against large-payload or "zip bomb" attacks by dropping frames that exceed these limits.

## License

This project is licensed under GNUv3.