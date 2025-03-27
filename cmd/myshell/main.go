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

		// Handle single quotes — completely literal
		if inSingleQuote {
			if ch == '\'' {
				inSingleQuote = false
			} else {
				current.WriteByte(ch)
			}
			continue
		}

		// Handle double quotes with limited escape support
		if inDoubleQuote {
			if escapeNext {
				switch ch {
				case '\\', '"', '$', '\n':
					current.WriteByte(ch)
				default:
					current.WriteByte('\\') // preserve the backslash
					current.WriteByte(ch)
				}
				escapeNext = false
				continue
			}

			if ch == '\\' {
				escapeNext = true
				continue
			}
			if ch == '"' {
				inDoubleQuote = false
			} else {
				current.WriteByte(ch)
			}
			continue
		}

		// Outside quotes
		if escapeNext {
			current.WriteByte(ch)
			escapeNext = false
			continue
		}

		switch ch {
		case '\\':
			escapeNext = true
		case '\'':
			inSingleQuote = true
		case '"':
			inDoubleQuote = true
		case ' ', '\t':
			if current.Len() > 0 {
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

		outputWriter := os.Stdout
		errorWriter := os.Stderr
		var outputFile, errorFile *os.File

		for i := 0; i < len(tokens); i++ {
			switch tokens[i] {
			case ">", "1>":
				if i+1 >= len(tokens) {
					fmt.Fprintln(os.Stderr, "syntax error: expected filename after", tokens[i])
					continue
				}
				var err error
				outputFile, err = os.Create(tokens[i+1])
				if err != nil {
					fmt.Fprintln(os.Stderr, "cannot open file for writing:", err)
					continue
				}
				outputWriter = outputFile
				tokens = append(tokens[:i], tokens[i+2:]...) // remove redirect tokens
				i -= 1                                       // step back to recheck this index
			case ">>", "1>>":
				if i+1 >= len(tokens) {
					fmt.Fprintln(os.Stderr, "syntax error: expected filename after", tokens[i])
					continue
				}
				var err error
				outputFile, err = os.OpenFile(tokens[i+1], os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					fmt.Fprintln(os.Stderr, "cannot open file for appending:", err)
					continue
				}
				outputWriter = outputFile
				tokens = append(tokens[:i], tokens[i+2:]...) // remove redirect tokens
				i -= 1                                       // step back to recheck this index
			case "2>":
				if i+1 >= len(tokens) {
					fmt.Fprintln(os.Stderr, "syntax error: expected filename after", tokens[i])
					continue
				}
				var err error
				errorFile, err = os.Create(tokens[i+1])
				if err != nil {
					fmt.Fprintln(os.Stderr, "cannot open file for writing:", err)
					continue
				}
				errorWriter = errorFile
				tokens = append(tokens[:i], tokens[i+2:]...) // remove redirect tokens
				i -= 1                                       // step back
			}
		}

		switch command {
		case "exit":
			if len(tokens) == 1 {
				os.Exit(0)
			}
			if exitCode, err := strconv.ParseInt(tokens[1], 10, 64); err == nil && tokens[0] == "exit" {
				os.Exit(int(exitCode))
			}
		case "echo":
			fmt.Fprintln(outputWriter, strings.Join(tokens[1:], " "))
			if outputFile != nil {
				if cerr := outputFile.Close(); cerr != nil {
					fmt.Fprintln(os.Stderr, "error closing output file:", cerr)
				}
			}
			if errorFile != nil {
				if cerr := errorFile.Close(); cerr != nil {
					fmt.Fprintln(os.Stderr, "error closing error file:", cerr)
				}
			}
		case "pwd":
			dir, err := os.Getwd()
			if err != nil {
			}
			fmt.Fprintln(outputWriter, dir)
			if outputFile != nil {
				if cerr := outputFile.Close(); cerr != nil {
					fmt.Fprintln(os.Stderr, "error closing output file:", cerr)
				}
			}
			if errorFile != nil {
				if cerr := errorFile.Close(); cerr != nil {
					fmt.Fprintln(os.Stderr, "error closing error file:", cerr)
				}
			}
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
				fmt.Fprintln(errorWriter, "cd: too many arguments")
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
					fmt.Fprintln(outputWriter, command+" is a shell builtin")
				} else if commandPath, err := exec.LookPath(command); err == nil {
					fmt.Fprintln(outputWriter, command+" is "+commandPath)
				} else {
					fmt.Fprintln(outputWriter, command+": not found")
				}

				if outputFile != nil {
					if cerr := outputFile.Close(); cerr != nil {
						fmt.Fprintln(os.Stderr, "error closing output file:", cerr)
					}
				}
				if errorFile != nil {
					if cerr := errorFile.Close(); cerr != nil {
						fmt.Fprintln(os.Stderr, "error closing error file:", cerr)
					}
				}
			}
		default:
			_, err := exec.LookPath(command)
			if err != nil {
				fmt.Fprintln(os.Stderr, command+": command not found")
			} else {
				cmd := exec.Command(command, tokens[1:]...)
				cmd.Stdout = outputWriter
				cmd.Stderr = errorWriter

				err := cmd.Run()

				if outputFile != nil {
					if cerr := outputFile.Close(); cerr != nil {
						fmt.Fprintln(os.Stderr, "error closing output file:", cerr)
					}
				}
				if errorFile != nil {
					if cerr := errorFile.Close(); cerr != nil {
						fmt.Fprintln(os.Stderr, "error closing error file:", cerr)
					}
				}

				if err != nil {
					//fmt.Fprintln(os.Stderr, err)
				}
			}

		}
	}

}
