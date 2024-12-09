package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
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

		s := strings.Trim(input, "\r\n")
		var tokens []string
		for {
			start := strings.Index(s, "'")
			if start == -1 {
				tokens = append(tokens, strings.Fields(s)...)
				break
			}
			tokens = append(tokens, strings.Fields(s[:start])...)
			s = s[start+1:]
			end := strings.Index(s, "'")
			token := s[:end]
			tokens = append(tokens, token)
			s = s[end+1:]
		}

		if len(tokens) == 1 && tokens[0] == "" {
			continue
		}
		command := tokens[0]

		switch command {
		case "exit":
			if len(tokens) == 1 {
				os.Exit(0)
			}
			if exitCode, err := strconv.ParseInt(tokens[1], 10, 64); err == nil && tokens[0] == "exit" {
				os.Exit(int(exitCode))
			}
		case "echo":
			fmt.Println(strings.Join(tokens[1:], " "))
		case "pwd":
			dir, err := os.Getwd()
			if err != nil {
			}
			fmt.Println(dir)
		case "cd":
			// This approach works for relative and absolute paths
			dir := ""
			switch len(tokens[1:]) {
			case 0:
				dir = os.Getenv("HOME")
			case 1:
				dir = tokens[1]
				if dir[0] == '~' && len(dir) > 1 {
					dir = os.Getenv("HOME") + dir[1:]
				} else if dir[0] == '~' {
					dir = os.Getenv("HOME")
				}
			default:
				fmt.Println("cd: too many arguments")
				continue
			}

			err := os.Chdir(dir)
			if err != nil {
				fmt.Println("cd: " + tokens[1] + ": No such file or directory")
			}
		case "type":
			if len(tokens) == 1 {
				fmt.Fprintln(os.Stdout, "type: missing argument")
			} else {
				builtinCommands := []string{
					"exit",
					"echo",
					"pwd",
					"cd",
					"type",
				}
				command := tokens[1]
				if slices.Contains(builtinCommands, command) {
					fmt.Fprintln(os.Stdout, command+" is a shell builtin")
				} else if commandPath, err := exec.LookPath(command); err == nil {
					fmt.Fprintln(os.Stdout, command+" is "+commandPath)
				} else {
					fmt.Fprintln(os.Stdout, command+": not found")
				}
			}
		default:
			_, err := exec.LookPath(command)
			if err != nil {
				fmt.Fprintln(os.Stdout, command+": command not found")
			} else {
				cmd := exec.Command(command, tokens[1:]...)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr

				err := cmd.Run()
				if err != nil {
					//fmt.Fprintln(os.Stderr, err)
				}
			}

		}
	}

}
