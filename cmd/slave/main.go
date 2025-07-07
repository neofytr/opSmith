package main

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/neofytr/opSmith/base"
)

const port string = "6969"

func main() {
	listener, err := net.Listen("tcp", "localhost:"+port)
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	fmt.Printf("Slave listening on port %s", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Failed to accept connection:", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	fmt.Println("Master connected:", conn.RemoteAddr().String())

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Printf("Failed to read from master %s -> %s\n", conn.RemoteAddr().String(), err.Error())
		return
	}

	// master will send just a single command
	var prim base.Primitive
	err = json.Unmarshal(buf[:n], &prim)
	if err != nil {
		fmt.Printf("Failed to unmarshal primitive from master %s -> %s\n", conn.RemoteAddr().String(), err.Error())
		return
	}

	response := prim.Run()
	data, err := json.Marshal(&response)
	if err != nil {
		fmt.Printf("Failed to marshal response from primitive %s -> %s\n", prim.Name, err.Error())
		return
	}

	_, err = conn.Write(data)
	if err != nil {
		fmt.Printf("Failed to write response to master %s -> %s\n", conn.RemoteAddr().String(), err.Error())
		return
	}

	fmt.Printf("Response sent to master %s: %s\n", conn.RemoteAddr().String(), response.Data)
}
