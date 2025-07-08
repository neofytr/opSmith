package base

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"runtime"
)

// status constants for command execution
const (
	StatusOK    = 0
	StatusError = -1
)

// command represents a single executable command
type Command struct {
	Name string   `json:"name"`
	Args []string `json:"args"`
}

// response represents the result of command execution
type Response struct {
	Data   string `json:"data"`
	Error  string `json:"error"`
	Status int    `json:"status"`
}

// batch represents multiple commands to execute
type Batch struct {
	Commands []Command `json:"commands"`
}

// batchResponse represents results from multiple commands
type BatchResponse struct {
	Results []Response `json:"results"`
	Status  int        `json:"status"`
}

// primitiveFunc defines the signature for primitive functions
type primitiveFunc func(args []string) (string, error)

// primitiveRegistry holds all available primitives
var primitiveRegistry = map[string]primitiveFunc{
	"ReadFile":    readFile,
	"CreateFile":  createFile,
	"DeleteFile":  deleteFile,
	"WriteFile":   writeFile,
	"AppendFile":  appendFile,
	"CommandExec": commandExec,
	// new primitives here
}

func commandExec(args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("Command requires exactly one argument (command to execute)")
	}

	command := args[0]
	if command == "" {
		return "", fmt.Errorf("command cannot be empty")
	}

	cmd := exec.Command("/bin/bash", "-c", command)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("could not execute command %s: %w", command, err)
	}

	if len(out) == 0 {
		return "", fmt.Errorf("command %s returned no output", command)
	}

	return string(out), nil
}

func appendFile(args []string) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("AppendFile requires exactly two arguments, filepath an contents")
	}

	filepath := args[0]
	if filepath == "" {
		return "", fmt.Errorf("file path cannot be empty")
	}

	var err error
	if runtime.GOOS == "linux" {
		filepath, err = expandPath(filepath)
		if err != nil {
			return "", fmt.Errorf("could not expand file path %s: %w", filepath, err)
		}
	}

	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_WRONLY, 0644) // wont create file it it doesn't exist (still takes the third argument)
	if err != nil {
		return "", fmt.Errorf("could not open file %s: %w", filepath, err)
	}
	defer file.Close()

	_, err = file.WriteString(args[1])
	if err != nil {
		return "", fmt.Errorf("could not append to file %s: %w", filepath, err)
	}

	return "Appended Successfully", nil
}

func deleteFile(args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("DeleteFile requires exactly one argument (file path)")
	}

	filepath := args[0]
	if filepath == "" {
		return "", fmt.Errorf("file path cannot be empty")
	}

	var err error
	if runtime.GOOS == "linux" {
		filepath, err = expandPath(filepath)
		if err != nil {
			return "", fmt.Errorf("could not expand file path %s: %w", filepath, err)
		}
	}

	err = os.Remove(filepath)
	if err != nil {
		return "", fmt.Errorf("could not delete file %s: %w", filepath, err)
	}

	return fmt.Sprintf("File %s deleted successfully", filepath), nil
}

func expandPath(path string) (string, error) {
	if path[:2] == "~/" {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		filepath := usr.HomeDir + "/" + path[2:] // usr.HomeDir is /home/username
		return filepath, nil
	}
	return path, nil
}

func createFile(args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("CreateFile requires exactly one argument(file path)")
	}

	filepath := args[0]
	var err error
	if runtime.GOOS == "linux" {
		filepath, err = expandPath(filepath)
		if err != nil {
			return "", fmt.Errorf("could not expand file path %s: %w", filepath, err)
		}
	}

	file, err := os.Create(filepath)
	if err != nil {
		return "", fmt.Errorf("could not create file %s: %w", filepath, err)
	}

	file.Close()
	return fmt.Sprintf("File %s created successfully", filepath), nil
}

func writeFile(args []string) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("writeFile requires exactly two arguments (file path and content)")
	}

	filepath := args[0]
	var err error
	if runtime.GOOS == "linux" {
		filepath, err = expandPath(filepath)
		if err != nil {
			return "", fmt.Errorf("could not expand file path %s: %w", filepath, err)
		}
	}
	content := args[1]
	if filepath == "" {
		return "", fmt.Errorf("file path cannot be empty")
	}

	file, err := os.OpenFile(filepath, os.O_WRONLY, 0644) // wont create file it it doesn't exist (still takes the third argument)
	if err != nil {
		return "", fmt.Errorf("could not open file %s: %w", filepath, err)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return "", fmt.Errorf("could not write to file %s: %w", filepath, err)
	}

	return "File written successfully", nil
}

// readFile primitive - reads entire file contents
func readFile(args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("readFile requires exactly one argument (file path)")
	}

	filepath := args[0]
	if filepath == "" {
		return "", fmt.Errorf("file path cannot be empty")
	}

	var err error
	if runtime.GOOS == "linux" {
		filepath, err = expandPath(filepath)
		if err != nil {
			return "", fmt.Errorf("could not expand file path %s: %w", filepath, err)
		}
	}

	file, err := os.Open(filepath)
	if err != nil {
		return "", fmt.Errorf("could not open file %s: %w", filepath, err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("could not read file %s: %w", filepath, err)
	}

	return string(data), nil
}

// registerPrimitive adds a new primitive to the registry
func RegisterPrimitive(name string, fn primitiveFunc) {
	primitiveRegistry[name] = fn
}

// run executes a single command
func (c *Command) Run() Response {
	if c.Name == "" {
		return Response{"", "command name cannot be empty", StatusError}
	}

	// look up primitive function in registry
	fn, exists := primitiveRegistry[c.Name]
	if !exists {
		return Response{"", fmt.Sprintf("primitive %s is not implemented", c.Name), StatusError}
	}

	// execute the primitive function
	data, err := fn(c.Args)
	if err != nil {
		return Response{"", fmt.Sprintf("error running primitive %s: %v", c.Name, err), StatusError}
	}

	return Response{data, "", StatusOK}
}

// runBatch executes multiple commands in sequence
func (b *Batch) RunBatch() BatchResponse {
	results := make([]Response, len(b.Commands))
	overallStatus := StatusOK

	// execute each command in the batch
	for i, cmd := range b.Commands {
		results[i] = cmd.Run()

		// if any command fails, mark overall status as error
		if results[i].Status != StatusOK {
			overallStatus = StatusError
		}
	}

	return BatchResponse{
		Results: results,
		Status:  overallStatus,
	}
}
