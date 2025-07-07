package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"

	basepkg "github.com/neofytr/opSmith/base"
)

// handleConnection processes incoming client connections
func handleConnection(conn net.Conn) {
	defer conn.Close()

	// read incoming data
	buf := make([]byte, 4096) // increased buffer size for multiple commands
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Printf("error reading from connection: %v\n", err)
		return
	}

	data := buf[:n]
	fmt.Printf("received: %s\n", string(data))

	// try to parse as batch first (multiple commands)
	var batch basepkg.Batch
	if err := json.Unmarshal(data, &batch); err == nil && len(batch.Commands) > 0 {
		// execute batch of commands
		batchResponse := batch.RunBatch()

		responseData, err := json.Marshal(batchResponse)
		if err != nil {
			fmt.Printf("error marshaling batch response: %v\n", err)
			return
		}

		conn.Write(responseData)
		return
	}
}

// startServer starts the slave server
func startServer(port string) error {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to start server on port %s: %w", port, err)
	}
	defer listener.Close()

	fmt.Printf("slave server listening on port %s\n", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("error accepting connection: %v\n", err)
			continue
		}

		// handle each connection in a separate goroutine
		go handleConnection(conn)
	}
}

func main() {
	args := os.Args
	if len(args) < 3 || args[1] != "--port" {
		fmt.Println("usage: slave --port <port_number>")
		return
	}

	port := args[2]
	if port == "" {
		fmt.Println("port number cannot be empty")
		return
	}

	if err := startServer(port); err != nil {
		fmt.Printf("server error: %v\n", err)
		os.Exit(1)
	}
}
