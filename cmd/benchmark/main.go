// main パッケージは、TCPとQUICプロトコルのベンチマークを行うためのコマンドラインツールを提供します。
package main

import (
	"flag"
	"log"
	"time"

	"github.com/Shinonome517/tcp-quic-bench/internal/client"
	"github.com/Shinonome517/tcp-quic-bench/internal/data"
	"github.com/Shinonome517/tcp-quic-bench/internal/server"
)

// main はアプリケーションのエントリーポイントです。
func main() {
	// コマンドラインフラグを定義します。
	mode := flag.String("mode", "server", "server or client")
	proto := flag.String("proto", "quic", "tcp or quic")
	addr := flag.String("addr", "0.0.0.0:4242", "address and port")
	flag.Parse()

	// モードを確認し、対応するロジックを実行します。
	switch *mode {
	case "server":
		runServer(*proto, *addr)
	case "client":
		runClient(*proto, *addr)
	default:
		log.Fatalf("Unknown mode: %s. Please use 'server' or 'client'.", *mode)
	}
}

func runServer(proto, addr string) {
	log.Println("Generating 1GB of random data...")
	benchmarkData, err := data.Generate()
	if err != nil {
		log.Fatalf("Failed to generate data: %v", err)
	}
	log.Println("Data generation complete.")

	switch proto {
	case "tcp":
		log.Printf("Starting TCP server on %s...", addr)
		if err := server.RunTCPServer(addr, benchmarkData); err != nil {
			log.Fatalf("TCP server failed: %v", err)
		}
	case "quic":
		log.Printf("Starting QUIC server on %s...", addr)
		if err := server.RunQUICServer(addr, benchmarkData); err != nil {
			log.Fatalf("QUIC server failed: %v", err)
		}
	default:
		log.Fatalf("Unknown protocol: %s", proto)
	}
}

func runClient(proto, addr string) {
	var totalBytes int64
	var duration time.Duration
	var err error

	switch proto {
	case "tcp":
		totalBytes, duration, err = client.RunTCPClient(addr)
	case "quic":
		totalBytes, duration, err = client.RunQUICClient(addr)
	default:
		log.Fatalf("Unknown protocol: %s", proto)
	}

	if err != nil {
		log.Fatalf("Client run failed: %v", err)
	}

	client.PrintResults(totalBytes, duration)
}
