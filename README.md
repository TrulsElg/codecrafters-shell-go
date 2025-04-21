Minimal Go Shell
================

A simple, interactive shell built in Go, created as part of the Codecrafters challenge. It supports a basic set of built-in commands, output redirection, quoting, and command autocomplete.

Supported Features
------------------

*   **Built-in commands:**
    
    *   ```exit```
        
    *   ```echo```
        
    *   ```pwd```
        
    *   ```cd```
        
    *   ```type```
        
*   **External command execution**
    
*   **Output redirection:**
    
    *   ```>``` and ```>>``` (stdout)
        
    *   ```2>``` and ```2>>``` (stderr)
        
*   **Interactive terminal mode:**
    * Basic autocomplete
      * Missing completions, bell sound made when no match is found
        
    *   Command history navigation (up/down arrows, limited to 5 latest commands)
        
    *   Line editing (backspace)
        

Planned Features
----------------
*   **Richer autocomplete**
  
    * Executable completion
  
    * Multiple completions
  
    *  Partial completions

*  **Richer history**
  
    * Deeper history
  
    * Reverse-i-search (CTRL+R)

*  **Piping**
  
    * Allow chaining commands using the pipe (|) operator
      * Example: ```ls -l | grep ".go" | sort```
  
    * Handle input and output streams properly between piped commands

*  **Environment variable**

    *   Access existing environment variables (e.g., echo $HOME)
    
    *   Allow setting new environment variables (e.g., MYVAR=value)
    
    *   Variable expansion within commands and arguments (e.g., cd $MYVAR/projects)

Build and Run
-------------

### Prerequisites

*   Go installed (version 1.21 or newer recommended)
    

### Building

`   go build -o mysh   `

### Running

To launch the shell interactively:

`   ./mysh   `

### Supported platforms
Tested and verified support on MacOS and Linux(Ubuntu).

