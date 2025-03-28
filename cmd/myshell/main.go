package main

import (
	"fmt"
	"golang.org/x/term"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
)

// Key constants (raw terminal mode input)
const (
	KEY_TAB       = 9
	KEY_ENTER     = 13
	KEY_BACKSPACE = 127
	KEY_CTRL_C    = 3
	KEY_ESC       = 27
)

var builtinCommands = []string{
	"exit",
	"echo",
	"pwd",
	"cd",
	"type",
}

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

func handleLine(lineInput string, oldState *term.State) {
	// Parse tokens + handle quoting and escaping
	tokens, err := parseTokens(strings.TrimSpace(lineInput))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Parse error:", err)
		return
	}

	if len(tokens) == 1 && tokens[0] == "" {
		return
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
				return
			}
			var err error
			outputFile, err = os.Create(tokens[i+1])
			if err != nil {
				fmt.Fprintln(os.Stderr, "cannot open file for writing:", err)
				return
			}
			outputWriter = outputFile
			tokens = append(tokens[:i], tokens[i+2:]...) // remove redirect tokens test
			i -= 1                                       // step back to recheck this index
		case ">>", "1>>":
			if i+1 >= len(tokens) {
				fmt.Fprintln(os.Stderr, "syntax error: expected filename after", tokens[i])
				return
			}
			var err error
			outputFile, err = os.OpenFile(tokens[i+1], os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Fprintln(os.Stderr, "cannot open file for appending:", err)
				return
			}
			outputWriter = outputFile
			tokens = append(tokens[:i], tokens[i+2:]...) // remove redirect tokens
			i -= 1                                       // step back to recheck this index
		case "2>":
			if i+1 >= len(tokens) {
				fmt.Fprintln(os.Stderr, "syntax error: expected filename after", tokens[i])
				return
			}
			var err error
			errorFile, err = os.Create(tokens[i+1])
			if err != nil {
				fmt.Fprintln(os.Stderr, "cannot open file for writing:", err)
				return
			}
			errorWriter = errorFile
			tokens = append(tokens[:i], tokens[i+2:]...) // remove redirect tokens
			i -= 1                                       // step back
		case "2>>":
			if i+1 >= len(tokens) {
				fmt.Fprintln(os.Stderr, "syntax error: expected filename after", tokens[i])
				return
			}
			var err error
			errorFile, err = os.OpenFile(tokens[i+1], os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Fprintln(os.Stderr, "cannot open error file for appending:", err)
				return
			}
			errorWriter = errorFile
			tokens = append(tokens[:i], tokens[i+2:]...) // remove redirect tokens
			i -= 1                                       // step back
		}
	}

	switch command {
	case "exit":
		if len(tokens) == 1 {
			term.Restore(int(os.Stdin.Fd()), oldState)
			os.Exit(0)
		}
		if exitCode, err := strconv.ParseInt(tokens[1], 10, 64); err == nil && tokens[0] == "exit" {
			term.Restore(int(os.Stdin.Fd()), oldState)
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
			return
		}

		err := os.Chdir(dir)
		if err != nil {
			fmt.Println("cd: " + tokens[1] + ": No such file or directory")
		}
	case "type":
		if len(tokens) == 1 {
			fmt.Fprintln(os.Stdout, "type: missing argument")
		} else {
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
			term.Restore(int(os.Stdin.Fd()), oldState) // Exit raw mode
			cmd := exec.Command(command, tokens[1:]...)
			cmd.Stdout = outputWriter
			cmd.Stderr = errorWriter
			cmd.Stdin = os.Stdin

			err := cmd.Run()

			oldState, _ = term.MakeRaw(int(os.Stdin.Fd())) // Re-enter raw mode

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

func handleAutocomplete(input []rune, cursorPos int) ([]rune, int) {
	prefix := string(input[:cursorPos])
	matches := []string{}

	for _, cmd := range builtinCommands {
		if strings.HasPrefix(cmd, prefix) {
			matches = append(matches, cmd)
		}
	}

	switch len(matches) {
	case 0:
		// No match: do nothing
		return input, cursorPos

	case 1:
		// Single match: insert remaining characters
		match := matches[0]
		remaining := match[cursorPos:]

		// Insert the remaining characters of the match
		for _, r := range remaining {
			input = append(input[:cursorPos], append([]rune{r}, input[cursorPos:]...)...)
			cursorPos++
		}

		// Insert a trailing space
		input = append(input[:cursorPos], append([]rune{' '}, input[cursorPos:]...)...)
		cursorPos++

		// Redraw the rest of the input
		restAfter := string(input[cursorPos:])
		fmt.Print(remaining + " " + restAfter)

		// Move cursor back to the correct position
		for i := 0; i < len(restAfter); i++ {
			fmt.Print("\x1b[D")
		}

	default:
		// Multiple matches: print suggestions on a new line
		fmt.Print("\n") // clean new line for suggestions
		for _, match := range matches {
			fmt.Print(match + " ")
		}
		fmt.Print("\n") // another newline to separate from prompt

		// Redraw the prompt and user input
		fmt.Printf("\r$ %s", string(input))

		// Move cursor to correct position
		for i := 0; i < len(input)-cursorPos; i++ {
			fmt.Print("\x1b[D")
		}
	}

	return input, cursorPos
}

func main() {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to enter raw mode:", err)
		os.Exit(1)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	fmt.Print("$ ")
	var input []rune
	var cursorPos int

	for {
		var buf [1]byte
		os.Stdin.Read(buf[:])
		ch := buf[0]

		switch ch {
		case KEY_CTRL_C:
			fmt.Println("\n\rExiting.\r")
			term.Restore(int(os.Stdin.Fd()), oldState)
			os.Exit(0)

		case KEY_ENTER:
			fmt.Print("\n\r")
			line := string(input)

			if line == "" {
				// Empty line — just reset and continue
				input = nil
				cursorPos = 0
				fmt.Printf("\r$ ")
				continue
			}

			handleLine(line, oldState)
			input = nil
			cursorPos = 0
			fmt.Printf("\r$ ")

		case KEY_BACKSPACE:
			if cursorPos > 0 {
				// Remove the character before the cursor
				input = append(input[:cursorPos-1], input[cursorPos:]...)
				cursorPos--

				// Redraw the rest of the input from the cursor position
				rest := string(input[cursorPos:])
				fmt.Print("\x1b[D")   // Move back to the deleted char
				fmt.Print(rest + " ") // Print rest and overwrite trailing char
				// Move cursor back to its proper position
				for i := 0; i < len(rest)+1; i++ {
					fmt.Print("\x1b[D")
				}
			}

		case KEY_TAB:
			trimmed := strings.TrimLeft(string(input[:cursorPos]), " \t")

			if len(trimmed) == 0 {
				// Just insert spaces
				tabWidth := 4
				spaces := []rune("    ")

				input = append(input[:cursorPos], append(spaces, input[cursorPos:]...)...)
				cursorPos += tabWidth

				rest := string(input[cursorPos:])
				fmt.Print("    " + rest)

				for i := 0; i < len(rest); i++ {
					fmt.Print("\x1b[D")
				}
			} else {
				// Only pass trimmed prefix, not the entire input
				input, cursorPos = handleAutocomplete(input, cursorPos)
			}

		case KEY_ESC:
			// Read 2 more bytes
			var seq [2]byte
			os.Stdin.Read(seq[:])

			if seq[0] == '[' {
				switch seq[1] {
				case 'A': // Up arrow
					// Intercept, no-op
				case 'B': // Down arrow
					// Intercept, no-op
				case 'C': // Right arrow
					if cursorPos < len(input) {
						cursorPos++
						fmt.Print("\x1b[C")
					}
				case 'D': // Left arrow
					if cursorPos > 0 {
						cursorPos--
						fmt.Print("\x1b[D")
					}
				}
			}

		default:
			chRune := rune(ch)

			// Insert character into input buffer at cursorPos
			if cursorPos < 0 {
				cursorPos = 0
			}
			if cursorPos > len(input) {
				cursorPos = len(input)
			}

			input = append(input[:cursorPos], append([]rune{chRune}, input[cursorPos:]...)...)
			cursorPos++

			// Redraw from the current cursor position
			rest := string(input[cursorPos:])
			fmt.Printf("%c%s", chRune, rest)

			// Move cursor back to logical position
			for i := 0; i < len(rest); i++ {
				fmt.Print("\x1b[D")
			}
		}

	}

}
