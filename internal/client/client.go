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
	"golang.org/x/sys/unix"
)

// RunTCPClient はTCPサーバーに接続し、パフォーマンスを測定します。
// 指定されたアドレスにTCP接続を試み、サーバーからのデータストリームを受信します。
// 受信したデータは破棄され、転送にかかった時間と総バイト数を返します。
func RunTCPClient(addr string) (int64, time.Duration, time.Duration, error) {
	// TLS設定を作成（自己署名証明書を許容）
	tlsConf := &tls.Config{
		InsecureSkipVerify: true, // サーバーは自己署名証明書のため検証をスキップ
		NextProtos:         []string{"tcp-quic-bench"},
	}

	log.Println("Connecting via TCP...")

	// TCP接続を確立
	handshakeStartTime := time.Now()
	dialer := &net.Dialer{}
	rawConn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to dial TCP: %w", err)
	}

	// TCP接続のファイルディスクリプタを取得し、MSSを設定
	tcpConn, ok := rawConn.(*net.TCPConn)
	if !ok {
		return 0, 0, 0, fmt.Errorf("failed to get TCP connection")
	}
	syscallConn, err := tcpConn.SyscallConn()
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get syscall connection: %w", err)
	}
	err = syscallConn.Control(func(fd uintptr) {
		err := unix.SetsockoptInt(int(fd), unix.IPPROTO_TCP, unix.TCP_MAXSEG, 1240) // MSSを1240に設定
		if err != nil {
			log.Printf("failed to set TCP_MAXSEG: %v", err)
		}
	})
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to control raw connection: %w", err)
	}

	// TLSハンドシェイク
	conn := tls.Client(rawConn, tlsConf)
	if err := conn.Handshake(); err != nil {
		return 0, 0, 0, fmt.Errorf("failed to perform TLS handshake: %w", err)
	}
	handshakeDuration := time.Since(handshakeStartTime)
	defer conn.Close()

	log.Println("TCP connection established. Receiving data...")

	// サーバーからのデータをすべて受信し、io.Discardで破棄する
	dataTransferStartTime := time.Now()
	bytesCopied, err := io.Copy(io.Discard, conn)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to receive data over TCP: %w", err)
	}
	dataTransferDuration := time.Since(dataTransferStartTime)

	log.Println("TCP data transfer complete.")
	return bytesCopied, handshakeDuration, dataTransferDuration, nil
}

// RunQUICClient はQUICサーバーに接続し、パフォーマンスを測定します。
// 自己署名証明書を許容するTLS設定でQUIC接続を試み、サーバーからのストリームを受信します。
// 受信したデータは破棄され、転送にかかった時間と総バイト数を返します。
func RunQUICClient(addr string) (int64, time.Duration, time.Duration, error) {
	// QUIC接続のためのTLS設定
	tlsConf := &tls.Config{
		InsecureSkipVerify: true, // サーバーは自己署名証明書のため検証をスキップ
		NextProtos:         []string{"tcp-quic-bench"},
	}

	log.Println("Connecting via QUIC...")

	// QUICの設定
	quicConfig := &quic.Config{
		DisablePathMTUDiscovery: true,
		MaxIdleTimeout:          time.Minute,
	}

	handshakeStartTime := time.Now()
	// QUICサーバーにダイヤル
	conn, err := quic.DialAddr(context.Background(), addr, tlsConf, quicConfig)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to connect to QUIC server: %w", err)
	}
	handshakeDuration := time.Since(handshakeStartTime)
	defer conn.CloseWithError(0, "")

	log.Println("QUIC connection established. Opening stream...")

	// サーバーからのストリームを受け入れる
	stream, err := conn.AcceptStream(context.Background())
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to open QUIC stream: %w", err)
	}
	defer stream.Close()

	log.Println("QUIC stream opened. Receiving data...")

	// ストリームからのデータをすべて受信し、io.Discardで破棄する
	dataTransferStartTime := time.Now()
	bytesCopied, err := io.Copy(io.Discard, stream)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to receive data over QUIC: %w", err)
	}
	dataTransferDuration := time.Since(dataTransferStartTime)

	log.Println("QUIC data transfer complete.")
	return bytesCopied, handshakeDuration, dataTransferDuration, nil
}
