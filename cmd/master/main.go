package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	basepkg "github.com/neofytr/opSmith/base"
)

func usage(programName string) {
	fmt.Println("Correct Usage:")
	fmt.Println(programName + " <command> <args>")
	fmt.Println("Available Commands and their arguments:")
	fmt.Println("1. --run <msg>")
	fmt.Println("2. --chat")
}

func masterRun(msg string) bool {
	config := basepkg.CreateConfig(basepkg.Llama3, "", "", 120*time.Second)
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

	response, err := client.GetResponse(ctx, msg)
	if err != nil {
		fmt.Println("Error getting response:", err)
		return false
	}

	fmt.Printf("Response: %s\n", response)
	return true
}

func main() {
	args := os.Args
	programName := args[0]
	if len(args) < 2 {
		usage(programName)
		return
	}

	if args[1] == "--run" {
		if len(args) < 3 {
			fmt.Println("Missing argument for --run")
			usage(programName)
			return
		}
		ret := masterRun(args[2])
		if !ret {
			fmt.Println("Error running the command '" + args[1] + "'")
		}
	} else {
		fmt.Println("Invalid command " + args[1])
	}
}
