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

func runLLM(msg string) (response string, success bool) {
	ollamaReq := OllamaRequest{
		Model:  "llama2",
		Prompt: msg,
		Stream: false,
	}

	data, err := json.Marshal(ollamaReq)
	if err != nil {
		fmt.Println("Error marshalling ollama request: " + err.Error())
		return "", false
	}

	resp, err := http.Post("http://localhost:11434/api/generate", "application/json", bytes.NewReader(data))
	if err != nil {
		fmt.Println("Error posting to ollama: " + err.Error())
		return "", false
	}
	defer resp.Body.Close()

	var ollamaResp OllamaResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading ollama response: " + err.Error())
		return "", false
	}

	err = json.Unmarshal(body, &ollamaResp)
	if err != nil {
		fmt.Println("Error unmarshalling ollama response: " + err.Error())
		return "", false
	}

	return ollamaResp.Response, true
}

func masterRun(msg string) bool {

	response, success := runLLM(msg)
	if !success {
		fmt.Println("Error getting response from the LLM")
		return false
	}

	fmt.Println("Response from ollama: " + response)
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
