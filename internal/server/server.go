// server パッケージは、TCPおよびQUICサーバーを作成するための機能を提供します。
package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"
	// pprofのためのブランクインポート。これはpprofをインクルードする標準的な方法です。
	_ "net/http/pprof"

	tlsutil "github.com/Shinonome517/tcp-quic-bench/internal/tls"
	"github.com/quic-go/quic-go"
	"golang.org/x/sys/unix"
)

// pprofServer は、pprofデータを提供するためにlocalhost:6060でHTTPサーバーを開始します。
// この関数はブロッキングするため、別のゴルーチンで実行する必要があります。
func pprofServer() {
	log.Println("Starting pprof server on :6060")
	if err := http.ListenAndServe("localhost:6060", nil); err != nil {
		log.Fatalf("pprof server failed: %v", err)
	}
}

// RunTCPServer は、指定されたアドレスでTCPサーバーを開始します。接続してきたクライアントに
// 提供されたデータを送信します。
func RunTCPServer(addr string, data []byte) error {
	// pprofサーバーを別のゴルーチンで開始し、ブロッキングしないようにします。
	go pprofServer()

	// TLS設定を取得（自己署名証明書）
	tlsConfig, err := tlsutil.Setup()
	if err != nil {
		return fmt.Errorf("failed to setup TLS: %w", err)
	}

	// TCPリスナーを生成
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	defer l.Close()
	log.Printf("TCP server listening on %s", addr)

	// 無限ループで新しい接続を受け入れます。
	for {
		rawConn, err := l.Accept()
		if err != nil {
			log.Printf("failed to accept connection: %v", err)
			continue
		}

		// TCP接続のファイルディスクリプタを取得し、MSSを設定
		// TCPPayload1240 + IPHeader20 + TCPHeader20 = 1280
		tcpConn, ok := rawConn.(*net.TCPConn)
		if !ok {
			log.Printf("failed to get TCP connection")
			rawConn.Close()
			continue
		}
		syscallConn, err := tcpConn.SyscallConn()
		if err != nil {
			log.Printf("failed to get syscall connection: %v", err)
			rawConn.Close()
			continue
		}
		err = syscallConn.Control(func(fd uintptr) {
			err := unix.SetsockoptInt(int(fd), unix.IPPROTO_TCP, unix.TCP_MAXSEG, 1240) // MSSを1240に設定
			if err != nil {
				log.Printf("failed to set TCP_MAXSEG: %v", err)
			}
		})
		if err != nil {
			log.Printf("failed to control raw connection: %v", err)
			rawConn.Close()
			continue
		}

		// TLSハンドシェイク
		conn := tls.Server(rawConn, tlsConfig)

		log.Printf("Accepted TCP connection from %s", conn.RemoteAddr())

		// 各接続を新しいゴルーチンで処理します。
		go func(c net.Conn) {
			// 関数が返るときに接続をクローズします。
			defer c.Close()
			// データをクライアントに書き込みます。
			if _, err := c.Write(data); err != nil {
				log.Printf("failed to write data to client: %v", err)
			}
		}(conn)
	}
}

// RunQUICServer は、指定されたアドレスでQUICサーバーを開始します。接続してきたクライアントに
// 提供されたデータを送信します。
func RunQUICServer(addr string, data []byte) error {
	// pprofサーバーを別のゴルーチンで開始し、ブロッキングしないようにします。
	go pprofServer()

	// QUICのためのTLS設定をセットアップします。
	tlsConfig, err := tlsutil.Setup()
	if err != nil {
		return fmt.Errorf("failed to setup TLS: %w", err)
	}

	// QUICの設定
	// DisablePathMTUDiscoveryをtrueに設定し、Path MTU Discovery（RFC 8899）を無効化
	// InitialPacketSizeがquic.Configにはセットされ，それはprotocol.InitialPacketSize = 1280に対応する．
	// UDPPayload1252 + IPHeader20 + UDPHeader8 = 1280
	quicConfig := &quic.Config{
		DisablePathMTUDiscovery: true,
		MaxIdleTimeout:          time.Minute,
	}

	// QUIC接続をリッスンします。
	l, err := quic.ListenAddr(addr, tlsConfig, quicConfig)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	// アプリケーション終了時にリスナーをクローズします。
	defer l.Close()
	log.Printf("QUIC server listening on %s", addr)

	// 無限ループで新しい接続を受け入れます。
	for {
		conn, err := l.Accept(context.Background())
		if err != nil {
			log.Printf("failed to accept connection: %v", err)
			continue
		}
		log.Printf("Accepted QUIC connection from %s", conn.RemoteAddr())

		// 各接続を新しいゴルーチンで処理します。
		go func(c *quic.Conn) {
			// 新しいストリームを開きます。
			stream, err := c.OpenStreamSync(context.Background())
			if err != nil {
				log.Printf("failed to open stream: %v", err)
				return
			}
			// 関数が返るときにストリームをクローズします。
			defer stream.Close()

			// データをクライアントに書き込みます。
			if _, err := stream.Write(data); err != nil {
				log.Printf("failed to write data to client: %v", err)
			}
		}(conn)
	}
}
