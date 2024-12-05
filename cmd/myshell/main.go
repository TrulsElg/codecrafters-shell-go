package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Fprint

func main() {
	for true {
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
			if exitCode, err := strconv.ParseInt(words[1], 10, 64); err == nil && words[0] == "exit" {
				os.Exit(int(exitCode))
			} else {
				os.Exit(0)
			}
		case "echo":
			fmt.Println(strings.Join(words[1:], " "))
		default:
			fmt.Fprintln(os.Stdout, command+": command not found")
		}
	}

}
