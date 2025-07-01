# TCP vs QUIC Performance Benchmark Tool

## Overview

This tool helps you compare how fast TCP and QUIC protocols transfer data. It's a simple benchmark application.

The server starts with a chosen protocol (TCP or QUIC) and creates 1GB of random data. When a client connects, the server sends this data to the client.

The client connects to the server using the specified protocol and receives the data. After receiving all data, it calculates and shows the time taken, total bytes received, and throughput (in Gbps).

## How to Run

You need to run the server and client in separate terminal windows.

### 1. Build the Executable

First, build the program:

```sh
go build -o tcp-quic-bench cmd/benchmark/main.go
```

### 2. Start the Server

Next, start the server. You can choose `tcp` or `quic` using the `-proto` flag. It's a good idea to run the server in the background (add `&` at the end).

#### For QUIC Server

```sh
./tcp-quic-bench -mode server -proto quic &
```

#### For TCP Server

```sh
./tcp-quic-bench -mode server -proto tcp &
```

The server will listen on `127.0.0.1:4242` by default.

### 3. Run the Client

Once the server is running, run the client to measure performance.

#### For QUIC Client

```sh
./tcp-quic-bench -mode client -proto quic -addr 127.0.0.1:4242
```

#### For TCP Client

```sh
./tcp-quic-bench -mode client -proto tcp -addr 127.0.0.1:4242
```

After the client finishes, you'll see benchmark results in your terminal, like this:

```
--- Benchmark Results ---
Total bytes received per run: 1073741824 bytes
Number of measurement runs: 10
-------------------------
Handshake time (Mean):      0.0000 s
Handshake time (StdDev):    0.0000 s
-------------------------
Data transfer time (Mean):  0.0000 s
Data transfer time (StdDev): 0.0000 s
-------------------------
Total time (Mean):          0.0000 s
Total time (StdDev):        0.0000 s
-------------------------
Throughput (Mean):          0.0000 Gbps
-------------------------
```
