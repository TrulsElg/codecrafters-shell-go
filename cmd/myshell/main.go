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

func parseTokens(input string) ([]string, error) {
	var tokens []string
	var current strings.Builder
	inSingleQuote := false
	inDoubleQuote := false
	escapeNext := false

	for i := 0; i < len(input); i++ {
		ch := input[i]

		if !inSingleQuote && !inDoubleQuote && escapeNext {
			current.WriteByte(ch)
			escapeNext = false
			continue
		}

		if !inSingleQuote && !inDoubleQuote && ch == '\\' {
			escapeNext = true
			continue
		}

		switch ch {
		case '\'':
			if !inDoubleQuote {
				inSingleQuote = !inSingleQuote
			} else {
				current.WriteByte(ch)
			}
		case '"':
			if !inSingleQuote {
				inDoubleQuote = !inDoubleQuote
			} else {
				current.WriteByte(ch)
			}
		case ' ', '\t':
			if inSingleQuote || inDoubleQuote {
				current.WriteByte(ch)
			} else if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		default:
			current.WriteByte(ch)
		}
	}

	if escapeNext {
		return nil, fmt.Errorf("unexpected end of input after backslash")
	}
	if inSingleQuote || inDoubleQuote {
		return nil, fmt.Errorf("unclosed quote")
	}
	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}
	return tokens, nil
}

func main() {
	for {
		fmt.Fprint(os.Stdout, "$ ")

		// Wait for user input
		input, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			os.Exit(1)
		}

		// Parse tokens + handle quoting and escaping
		tokens, err := parseTokens(strings.TrimSpace(input))
		if err != nil {
			fmt.Fprintln(os.Stderr, "Parse error:", err)
			continue
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
