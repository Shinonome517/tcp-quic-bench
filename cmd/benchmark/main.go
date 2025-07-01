// main パッケージは、TCPとQUICプロトコルのベンチマークを行うためのコマンドラインツールを提供する。
package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/Shinonome517/tcp-quic-bench/internal/client"
	"github.com/Shinonome517/tcp-quic-bench/internal/data"
	"github.com/Shinonome517/tcp-quic-bench/internal/server"
)

const (
	warmupRuns      = 2  // ウォームアップ実行回数
	measurementRuns = 10 // 計測実行回数
)

// main はアプリケーションのエントリーポイントである。
func main() {
	// コマンドラインフラグを定義する。
	mode := flag.String("mode", "server", "server or client")
	proto := flag.String("proto", "quic", "tcp or quic")
	addr := flag.String("addr", "127.0.0.1:4242", "address and port")
	flag.Parse()

	// モードを確認し、対応するロジックを実行する。
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
	handshakeDurations := make([]time.Duration, 0, measurementRuns)
	dataTransferDurations := make([]time.Duration, 0, measurementRuns)

	log.Printf("Starting %s client with %d warmup runs and %d measurement runs...", proto, warmupRuns, measurementRuns)

	for i := 0; i < warmupRuns+measurementRuns; i++ {
		var hsDur, dtDur time.Duration
		var err error
		var bytes int64

		switch proto {
		case "tcp":
			bytes, hsDur, dtDur, err = client.RunTCPClient(addr)
		case "quic":
			bytes, hsDur, dtDur, err = client.RunQUICClient(addr)
		default:
			log.Fatalf("Unknown protocol: %s", proto)
		}

		if err != nil {
			log.Fatalf("Client run failed: %v", err)
		}

		if i == 0 { // 最初の実行で totalBytes を取得
			totalBytes = bytes
		}

		if i >= warmupRuns {
			handshakeDurations = append(handshakeDurations, hsDur)
			dataTransferDurations = append(dataTransferDurations, dtDur)
		}
		time.Sleep(100 * time.Millisecond) // 各実行間の短い待機
	}

	PrintResults(totalBytes, handshakeDurations, dataTransferDurations)
}

// calculateStatistics は time.Duration のスライスから平均値と標準偏差を計算する。
func calculateStatistics(durations []time.Duration) (mean, stdDev time.Duration) {
	if len(durations) == 0 {
		return 0, 0
	}

	var sum float64
	for _, d := range durations {
		sum += d.Seconds()
	}
	meanSeconds := sum / float64(len(durations))

	var sumSqDiff float64
	for _, d := range durations {
		diff := d.Seconds() - meanSeconds
		sumSqDiff += diff * diff
	}

	variance := sumSqDiff / float64(len(durations))
	stdDevSeconds := math.Sqrt(variance)

	mean = time.Duration(meanSeconds * float64(time.Second))
	stdDev = time.Duration(stdDevSeconds * float64(time.Second))
	return
}

// PrintResults はベンチマーク結果を計算して表示する。
// 総バイト数と時間からスループット（Gbps）を算出し、整形して標準出力に表示する。
func PrintResults(totalBytes int64, handshakeDurations, dataTransferDurations []time.Duration) {
	handshakeMean, handshakeStdDev := calculateStatistics(handshakeDurations)
	dataTransferMean, dataTransferStdDev := calculateStatistics(dataTransferDurations)

	// 各実行の合計時間を計算し、その統計情報を取得
	var totalDurations []time.Duration
	for i := 0; i < len(handshakeDurations); i++ {
		totalDurations = append(totalDurations, handshakeDurations[i]+dataTransferDurations[i])
	}
	totalMean, totalStdDev := calculateStatistics(totalDurations)

	// 平均スループットの計算
	// 各実行のスループットを計算し、その平均を取る
	var throughputs []float64
	for _, totalDur := range totalDurations {
		totalDurSeconds := totalDur.Seconds()
		if totalDurSeconds > 0 {
			throughputs = append(throughputs, (float64(totalBytes)*8)/(totalDurSeconds*1e9))
		}
	}

	var sumThroughput float64
	for _, t := range throughputs {
		sumThroughput += t
	}
	meanThroughput := sumThroughput / float64(len(throughputs))

	fmt.Println("\n--- Benchmark Results ---")
	fmt.Printf("Total bytes received per run: %d bytes\n", totalBytes)
	fmt.Printf("Number of measurement runs: %d\n", len(handshakeDurations))
	fmt.Println("-------------------------")

	fmt.Printf("Handshake time (Mean):      %.4f s\n", handshakeMean.Seconds())
	fmt.Printf("Handshake time (StdDev):    %.4f s\n", handshakeStdDev.Seconds())
	fmt.Println("-------------------------")

	fmt.Printf("Data transfer time (Mean):  %.4f s\n", dataTransferMean.Seconds())
	fmt.Printf("Data transfer time (StdDev): %.4f s\n", dataTransferStdDev.Seconds())
	fmt.Println("-------------------------")

	fmt.Printf("Total time (Mean):          %.4f s\n", totalMean.Seconds())
	fmt.Printf("Total time (StdDev):        %.4f s\n", totalStdDev.Seconds())
	fmt.Println("-------------------------")

	fmt.Printf("Throughput (Mean):          %.4f Gbps\n", meanThroughput)
	fmt.Println("-------------------------")
}
