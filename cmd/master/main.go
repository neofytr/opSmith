package main

import (
	"fmt"
	"os"

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
	config = 
	client := basepkg.NewLLMClient()
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
