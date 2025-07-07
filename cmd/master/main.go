package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"time"

	basepkg "github.com/neofytr/opSmith/base"
)

const genMsg string = `You are a command-generating assistant that talks to a client machine.

Your job is to understand the user's natural language input and convert it into **valid command instructions** that the client (slave) can understand and execute.

The client supports the following commands:

1. ReadFile <file_path> - Opens a file and reads its entire contents, then closes it.

Respond with a single **JSON object only**, following this structure:

{
  "name": "<CommandName>",
  "args": ["<arg1>", "<arg2>", ...]
}

Do not include any extra text, markdown, or explanation — just return the pure JSON object.

Begin below:
`

func usage(programName string) {
	fmt.Println("Correct Usage:")
	fmt.Println(programName + " --client <IP>:<Port> [<command> <args>]")
	fmt.Println("Available Commands and their arguments:")
	fmt.Println("1. --run <msg>")
}

func masterRun(msg, clientIP, clientPort string) error {
	config := basepkg.CreateConfig(basepkg.Llama2, "", "", 120*time.Second)
	client := basepkg.NewLLMClient(config)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt) // ctrl + C capture

	go func() {
		<-sigChan
		fmt.Println("Received interrupt signal, shutting down...")
		cancel() // cancel the context to stop the client
	}()

	// this is a json response consisting of the command(s)
	response, err := client.GetResponse(ctx, genMsg+"\n"+msg)
	if err != nil {
		return fmt.Errorf("couldn't get AI response: %s", err)
	}

	conn, err := net.Dial("tcp", clientIP+":"+clientPort)
	if err != nil {
		return fmt.Errorf("couldn't connect to client at %s:%s: %s", clientIP, clientPort, err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte(response))
	if err != nil {
		return fmt.Errorf("couldn't send message to client: %s", err)
	}

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return fmt.Errorf("couldn't read response from client: %s", err)
	}

	if string(buf[:n]) == "OK" {
		fmt.Println("Command executed successfully on the client.")
	} else {
		return fmt.Errorf("command execution failure on the client: %s", string(buf[:n]))
	}

	return nil
}

func main() {
	args := os.Args
	programName := args[0]
	if len(args) < 2 {
		usage(programName)
		return
	}

	clientIP := ""
	clientPort := ""
	for index, argument := range args {
		if argument == "--client" {
			if index+2 >= len(args) {
				fmt.Println("Missing argument for --client")
				usage(programName)
				return
			}

			clientIP = args[index+1]
			clientPort = args[index+2]

			if clientIP == "" || clientPort == "" {
				fmt.Println("Invalid client IP or Port")
				usage(programName)
				return
			}
		}
	}

	if clientIP == "" || clientPort == "" {
		fmt.Println("Client IP and Port must be specified with --client")
		usage(programName)
		return
	}

	for index, argument := range args {
		switch argument {
		case "--run":
			{
				if index+1 >= len(args) {
					fmt.Println("Missing argument for --run")
					usage(programName)
					return
				}
				if err := masterRun(args[index+1], clientIP, clientPort); err != nil {
					fmt.Printf("Couldn't run the command -> %s\n", args[index+1])
					fmt.Printf("Error -> %v\n", err)
				}
			}
		}
	}

}
