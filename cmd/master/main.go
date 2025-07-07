package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"

	basepkg "github.com/neofytr/opSmith/base"
)

const loggingEnabled = false

// ai prompt for command generation
const aiPrompt = `You are a command-generating assistant that converts natural language to executable commands.

Available commands:
1. ReadFile <file_path> - reads entire file contents

Respond with:
{
  "commands": [
    {"name": "<CommandName1>", "args": ["<arg1>"]},
    {"name": "<CommandName2>", "args": ["<arg1>", "<arg2>"]}
  ]
}

Only return valid JSON, no extra text.

User request: `

// config holds master configuration
type config struct {
	clientIP   string
	clientPort string
	timeout    time.Duration
}

// parseArgs parses command line arguments
func parseArgs(args []string) (*config, string, error) {
	if len(args) < 2 {
		return nil, "", fmt.Errorf("insufficient arguments")
	}

	cfg := &config{timeout: 30 * time.Second}
	var command string

	// parse arguments
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--client":
			if i+2 >= len(args) {
				return nil, "", fmt.Errorf("--client requires IP and port")
			}
			cfg.clientIP = args[i+1]
			cfg.clientPort = args[i+2]
			i += 2
		case "--run":
			if i+1 >= len(args) {
				return nil, "", fmt.Errorf("--run requires a command")
			}
			command = args[i+1]
			i++
		case "--timeout":
			if i+1 >= len(args) {
				return nil, "", fmt.Errorf("--timeout requires a duration")
			}
			duration, err := time.ParseDuration(args[i+1])
			if err != nil {
				return nil, "", fmt.Errorf("invalid timeout format: %w", err)
			}
			cfg.timeout = duration
			i++
		case "--run-from-file":
			if i+1 >= len(args) {
				return nil, "", fmt.Errorf("--run-from-file requires a file path")
			}
			filePath := args[i+1]
			if filePath == "" {
				return nil, "", fmt.Errorf("file path cannot be empty")
			}

			file, err := os.Open(filePath)
			if err != nil {
				return nil, "", fmt.Errorf("could not open file %s: %w", filePath, err)
			}
			defer file.Close()

			data, err := io.ReadAll(file)
			if err != nil {
				return nil, "", fmt.Errorf("could not read file %s: %w", filePath, err)
			}

			command = strings.TrimSpace(string(data))
			if command == "" {
				return nil, "", fmt.Errorf("file is empty or contains only whitespace")
			}
		}
	}

	if cfg.clientIP == "" || cfg.clientPort == "" {
		return nil, "", fmt.Errorf("client IP and port must be specified")
	}

	if command == "" {
		return nil, "", fmt.Errorf("command must be specified with --run")
	}

	return cfg, command, nil
}

// generateCommands uses AI to convert natural language to commands
func generateCommands(ctx context.Context, message string) (string, error) {
	// create llm client
	config := basepkg.CreateConfig(basepkg.Llama2, "", "", 120*time.Second)
	client := basepkg.NewLLMClient(config)

	// get ai response
	response, err := client.GetResponse(ctx, aiPrompt+message)
	if err != nil {
		return "", fmt.Errorf("failed to get AI response: %w", err)
	}

	return strings.TrimSpace(response), nil
}

// sendToSlave sends commands to slave and returns response
func sendToSlave(cfg *config, commandJSON string) error {
	// connect to slave
	conn, err := net.DialTimeout("tcp", cfg.clientIP+":"+cfg.clientPort, cfg.timeout)
	if err != nil {
		return fmt.Errorf("failed to connect to slave at %s:%s: %w", cfg.clientIP, cfg.clientPort, err)
	}
	defer conn.Close()

	// send command
	_, err = conn.Write([]byte(commandJSON))
	if err != nil {
		return fmt.Errorf("failed to send command to slave: %w", err)
	}

	// read response
	buf := make([]byte, 8192) // larger buffer for multiple responses
	n, err := conn.Read(buf)
	if err != nil {
		return fmt.Errorf("failed to read response from slave: %w", err)
	}

	responseData := buf[:n]
	if loggingEnabled {
		log.Default().Printf("received response from slave: %s", string(responseData))
	}
	// parse as batch response first
	var batchResp basepkg.BatchResponse
	if err := json.Unmarshal(responseData, &batchResp); err == nil {
		// handle batch response
		if loggingEnabled {
			log.Default().Printf("Executed %d commands:\n", len(batchResp.Results))
		}

		for i, result := range batchResp.Results {
			fmt.Printf("Command %d ->\n", i+1)
			if result.Status == basepkg.StatusOK {
				if loggingEnabled {
					fmt.Printf("success\n%s\n", result.Data)
				}
				fmt.Printf("Response Data ->\n%s\n\n", result.Data)
			} else {
				fmt.Printf("failed: %s\n\n", result.Error)
			}
		}
		return nil
	}

	return fmt.Errorf("failed to unmarshal response from slave -> %s", err)
}

// runMaster executes the master logic
func runMaster(cfg *config, message string) error {
	// create context with cancellation
	ctx, cancel := context.WithTimeout(context.Background(), cfg.timeout)
	defer cancel()

	// handle interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	go func() {
		<-sigChan
		fmt.Println("\nreceived interrupt signal, shutting down...")
		cancel()
	}()

	// generate commands using AI
	if loggingEnabled {
		log.Default().Printf("generating commands for message: %s", message)
	}
	commandJSON, err := generateCommands(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to generate commands: %w", err)
	}

	if loggingEnabled {
		log.Default().Printf("generated commands: %s", commandJSON)
	}
	// send commands to slave
	return sendToSlave(cfg, commandJSON)
}

// showUsage displays usage information
func showUsage(programName string) {
	fmt.Printf("usage: %s --client <IP> <port> --run <command> [options]\n", programName)
	fmt.Println("options:")
	fmt.Println("  --client <IP> <port>  specify slave IP and port")
	fmt.Println("  --run <command>       natural language command to execute")
	fmt.Println("  --timeout <duration>  connection timeout (default: 30s)")
	fmt.Println("examples:")
	fmt.Printf("  %s --client 192.168.1.100 8080 --run \"read file config.txt\"\n", programName)
	fmt.Printf("  %s --client localhost 8080 --run \"read files a.txt and b.txt\" --timeout 60s\n", programName)
}

func main() {
	args := os.Args
	programName := args[0]

	// parse command line arguments
	cfg, command, err := parseArgs(args)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		showUsage(programName)
		os.Exit(1)
	}

	// run master
	if err := runMaster(cfg, command); err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}
