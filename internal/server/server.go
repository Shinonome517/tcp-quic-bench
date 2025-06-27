// Package server provides functionality for creating TCP and QUIC servers.
package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	// Blank import for pprof. This is the standard way to include pprof.
	_ "net/http/pprof"

	"github.com/quic-go/quic-go"
	"github.com/Shinonome517/tcp-quic-bench/internal/tls"
)

// pprofServer starts an HTTP server on localhost:6060 to serve pprof data.
// This function blocks, so it should be run in a separate goroutine.
func pprofServer() {
	log.Println("Starting pprof server on :6060")
	if err := http.ListenAndServe("localhost:6060", nil); err != nil {
		log.Fatalf("pprof server failed: %v", err)
	}
}

// TCPServer starts a TCP server on the given address. It sends the provided data
// to any client that connects.
func TCPServer(addr string, data []byte) error {
	// Start the pprof server in a separate goroutine so it doesn't block.
	go pprofServer()

	// Listen for incoming TCP connections.
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	// Close the listener when the application closes.
	defer l.Close()
	log.Printf("TCP server listening on %s", addr)

	// Loop forever, accepting new connections.
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("failed to accept connection: %v", err)
			continue
		}
		log.Printf("Accepted TCP connection from %s", conn.RemoteAddr())

		// Handle each connection in a new goroutine.
		go func(c net.Conn) {
			// Close the connection when the function returns.
			defer c.Close()
			// Write the data to the client.
			if _, err := c.Write(data); err != nil {
				log.Printf("failed to write data to client: %v", err)
			}
		}(conn)
	}
}

// QUICServer starts a QUIC server on the given address. It sends the provided data
// to any client that connects.
func QUICServer(addr string, data []byte) error {
	// Start the pprof server in a separate goroutine so it doesn't block.
	go pprofServer()

	// Set up the TLS configuration for QUIC.
	tlsConfig, err := tls.Setup()
	if err != nil {
		return fmt.Errorf("failed to setup TLS: %w", err)
	}

	// Listen for incoming QUIC connections.
	l, err := quic.ListenAddr(addr, tlsConfig, nil)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	// Close the listener when the application closes.
	defer l.Close()
	log.Printf("QUIC server listening on %s", addr)

	// Loop forever, accepting new connections.
	for {
		conn, err := l.Accept(context.Background())
		if err != nil {
			log.Printf("failed to accept connection: %v", err)
			continue
		}
		log.Printf("Accepted QUIC connection from %s", conn.RemoteAddr())

		// Handle each connection in a new goroutine.
		go func(c *quic.Conn) {
			// Open a new stream.
			stream, err := c.OpenStreamSync(context.Background())
			if err != nil {
				log.Printf("failed to open stream: %v", err)
				return
			}
			// Close the stream when the function returns.
			defer stream.Close()

			// Write the data to the client.
			if _, err := stream.Write(data); err != nil {
				log.Printf("failed to write data to client: %v", err)
			}
		}(conn)
	}
}