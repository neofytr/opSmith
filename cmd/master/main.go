package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"` // true if we want to stream the response
}

type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

func usage(programName string) {
	fmt.Println("Correct Usage:")
	fmt.Println(programName + " <command> <args>")
	fmt.Println("Available Commands and their arguments:")
	fmt.Println("1. --run <msg>")
}

func masterRun(msg string) bool {
	ollamaReq := OllamaRequest{
		Model:  "llama2",
		Prompt: msg,
		Stream: false,
	}

	data, _ := json.Marshal(ollamaReq)
	resp, err := http.Post("http://localhost:11434/api/generate", "application/json", bytes.NewReader(data))
	if err != nil {
		panic("Error posting to ollama: " + err.Error())
	}
	defer resp.Body.Close()

	var ollamaResp OllamaResponse
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &ollamaResp)

	fmt.Println("Response from ollama: " + ollamaResp.Response)
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
