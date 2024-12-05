package main

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Fprint

func main() {
	for {
		fmt.Fprint(os.Stdout, "$ ")

		// Wait for user input
		input, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			os.Exit(1)
		}

		input = strings.TrimSuffix(input, "\n")
		words := strings.Fields(input)
		command := words[0]

		switch command {
		case "exit":
			if len(words) == 1 {
				os.Exit(0)
			}
			if exitCode, err := strconv.ParseInt(words[1], 10, 64); err == nil && words[0] == "exit" {
				os.Exit(int(exitCode))
			}
		case "echo":
			fmt.Println(strings.Join(words[1:], " "))
		case "type":
			if len(words) == 1 {
				fmt.Fprintln(os.Stdout, "type: missing argument")
			} else {
				builtinCommands := []string{
					"exit",
					"echo",
					"type",
				}
				command := words[1]
				if slices.Contains(builtinCommands, command) {
					fmt.Fprintln(os.Stdout, command+" is a shell builtin")
				} else {
					fmt.Fprintln(os.Stdout, command+": command not found")
				}
			}
		default:
			fmt.Fprintln(os.Stdout, command+": command not found")
		}
	}

}
