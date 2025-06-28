package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/quic-go/quic-go"
)

// RunTCPClient はTCPサーバーに接続し、パフォーマンスを測定します。
// 指定されたアドレスにTCP接続を試み、サーバーからのデータストリームを受信します。
// 受信したデータは破棄され、転送にかかった時間と総バイト数を返します。
func RunTCPClient(addr string) (int64, time.Duration, error) {
	log.Println("Connecting via TCP...")
	// 計測開始
	startTime := time.Now()

	// TCPサーバーにダイヤル
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to connect to TCP server: %w", err)
	}
	defer conn.Close()

	log.Println("TCP connection established. Receiving data...")

	// サーバーからのデータをすべて受信し、io.Discardで破棄する
	bytesCopied, err := io.Copy(io.Discard, conn)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to receive data over TCP: %w", err)
	}

	// 計測終了
	duration := time.Since(startTime)
	log.Println("TCP data transfer complete.")
	return bytesCopied, duration, nil
}

// RunQUICClient はQUICサーバーに接続し、パフォーマンスを測定します。
// 自己署名証明書を許容するTLS設定でQUIC接続を試み、サーバーからのストリームを受信します。
// 受信したデータは破棄され、転送にかかった時間と総バイト数を返します。
func RunQUICClient(addr string) (int64, time.Duration, error) {
	log.Println("Connecting via QUIC...")
	// QUIC接続のためのTLS設定
	tlsConf := &tls.Config{
		InsecureSkipVerify: true, // サーバーは自己署名証明書のため検証をスキップ
		NextProtos:         []string{"quic-speed-test"},
	}

	// 計測開始
	startTime := time.Now()

	// QUICサーバーにダイヤル
	conn, err := quic.DialAddr(context.Background(), addr, tlsConf, nil)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to connect to QUIC server: %w", err)
	}
	defer conn.CloseWithError(0, "")

	log.Println("QUIC connection established. Opening stream...")

	// サーバーからのストリームを受け入れる
	stream, err := conn.AcceptStream(context.Background())
	if err != nil {
		return 0, 0, fmt.Errorf("failed to open QUIC stream: %w", err)
	}
	defer stream.Close()

	log.Println("QUIC stream opened. Receiving data...")

	// ストリームからのデータをすべて受信し、io.Discardで破棄する
	bytesCopied, err := io.Copy(io.Discard, stream)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to receive data over QUIC: %w", err)
	}

	// 計測終了
	duration := time.Since(startTime)
	log.Println("QUIC data transfer complete.")
	return bytesCopied, duration, nil
}

// PrintResults はベンチマーク結果を計算して表示します。
// 総バイト数と時間からスループット（Gbps）を算出し、整形して標準出力に表示します。
func PrintResults(totalBytes int64, duration time.Duration) {
	durationSeconds := duration.Seconds()
	if durationSeconds == 0 {
		log.Println("Duration was zero, cannot calculate throughput.")
		return
	}
	// スループットをGbps (Giga-bits per second) で計算
	throughputGbps := (float64(totalBytes) * 8) / (durationSeconds * 1e9)

	fmt.Println("\n--- Benchmark Results ---")
	fmt.Printf("Total bytes received: %d bytes\n", totalBytes)
	fmt.Printf("Total time taken:     %.2fs\n", durationSeconds)
	fmt.Printf("Throughput:           %.4f Gbps\n", throughputGbps)
	fmt.Println("-------------------------")
}
