// Package main provides a command-line tool for benchmarking TCP and QUIC protocols.
package main

import (
	"flag" // flag package implements command-line flag parsing.
	"log"  // log package implements a simple logging package.

	"github.com/Shinonome517/tcp-quic-bench/internal/data"
	"github.com/Shinonome517/tcp-quic-bench/internal/server"
)

// main is the entry point of the application.
func main() {
	// Define command-line flags.
	mode := flag.String("mode", "server", "server or client")       // mode flag to specify whether to run as a server or client.
	proto := flag.String("proto", "quic", "tcp or quic")            // proto flag to specify the protocol (tcp or quic).
	addr := flag.String("addr", "0.0.0.0:4242", "address and port") // addr flag to specify the address and port to listen on or connect to.
	flag.Parse()                                                    // Parse the command-line flags.

	// Check the mode and execute the corresponding logic.
	if *mode == "server" {
		log.Println("Generating 1GB of random data...")
		// Generate 1GB of random data for benchmarking.
		benchmarkData, err := data.Generate()
		if err != nil {
			log.Fatalf("Failed to generate data: %v", err) // Log and exit if data generation fails.
		}
		log.Println("Data generation complete.")

		// Start the server based on the specified protocol.
		switch *proto {
		case "tcp":
			log.Printf("Starting TCP server on %s...", *addr)
			if err := server.TCPServer(*addr, benchmarkData); err != nil {
				log.Fatalf("TCP server failed: %v", err) // Log and exit if TCP server fails.
			}
		case "quic":
			log.Printf("Starting QUIC server on %s...", *addr)
			if err := server.QUICServer(*addr, benchmarkData); err != nil {
				log.Fatalf("QUIC server failed: %v", err) // Log and exit if QUIC server fails.
			}
		default:
			log.Fatalf("Unknown protocol: %s", *proto) // Log and exit for unknown protocols.
		}
	} else {
		// Client mode is not yet implemented.
		log.Println("Client mode is not implemented yet.")
	}
}
