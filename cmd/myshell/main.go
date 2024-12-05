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
		command, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			os.Exit(1)
		}
		command = strings.TrimSuffix(command, "\n")
		words := strings.Fields(command)
		if len(words) > 1 {
			if exit_code, err := strconv.ParseInt(words[1], 10, 64); err == nil && words[0] == "exit" {
				os.Exit(int(exit_code))
			}
		}
		
		fmt.Fprintln(os.Stdout, command+": command not found")
	}

}
