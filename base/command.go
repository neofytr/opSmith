package base

import (
	"fmt"
	"io"
	"os"
)

const StatusOK int = 0
const StatusError int = -1

/*

Primitives

1. ReadFile - Opens a file and reads it's entire contents
Contains only a single argument - the file path

*/

type Primitive struct {
	Name string   `json:"name"`
	Args []string `json:"args"`
}

type Response struct {
	Data   string `json:"data"`
	Err    error  `json:"err,omitempty"`
	Status int    `json:"status,omitempty"`
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

func (p *Primitive) Run() Response {
	if p.Name == "" {
		return Response{"", fmt.Errorf("primitive name cannot be empty"), StatusError}
	}

	switch p.Name {
	case "ReadFile":
		{
			if len(p.Args) != 1 {
				return Response{"", fmt.Errorf("primitive %s requires exactly one argument", p.Name), StatusError}
			}

			data, err := runReadFile(p.Args[0])
			if err != nil {
				return Response{"", fmt.Errorf("error running primitive %s -> %w", p.Name, err), StatusError}
			}

			return Response{data, nil, StatusOK}
		}
	default:
		{
			return Response{"", fmt.Errorf("primitive %s is not implemented", p.Name), StatusError}
		}
	}
}
