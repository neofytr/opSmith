package base

import (
	"fmt"
	"io"
	"os"
)

/*

Primitives

1. ReadFile - Opens a file and reads it's entire contents
Contains only a single argument - the file path

*/

type Primitive struct {
	Name string   `json:"name"`
	Args []string `json:"args"`
}

func runReadFile(filepath string) (string, error) {
	if filepath == "" {
		return "", fmt.Errorf("file path cannot be empty")
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

func (p *Primitive) Run() (string, error) {
	if p.Name == "" {
		return "", fmt.Errorf("primitive name cannot be empty")
	}

	switch p.Name {
	case "ReadFile":
		{
			if len(p.Args) != 1 {
				return "", fmt.Errorf("primitive %s requires exactly one argument", p.Name)
			}
			return runReadFile(p.Args[0])
		}
	default:
		{
			return "", fmt.Errorf("primitive %s is not implemented", p.Name)
		}
	}
}
