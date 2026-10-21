package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/term"
)

// Movement         , ANSI Escape Code , Description
// Absolute Position, \033[<row>;<col>H, Moves cursor to a specific row and column (1-indexed).
// Move Up          , \033[<n>A        , Moves the cursor up n lines.
// Move Down        , \033[<n>B        , Moves the cursor down n lines.
// Move Forward     , \033[<n>C        , Moves the cursor right n columns.
// Move Backward    , \033[<n>D        , Moves the cursor left n columns.
// To Column 1      , \033[1G          , Moves the cursor to the beginning of the current line.
// Save / Restore   , \0337 / \0338    , Saves the current cursor position and restores it later.

func main() {
	fd := int(os.Stdout.Fd())
	if !term.IsTerminal(fd) {
		log.Fatal("Standard output is not a terminal")
	}

	// Hide the terminal cursor
	fmt.Print("\033[?25l")

	// Create a channel to receive OS signals
	sigChan := make(chan os.Signal, 1)

	// Register the channel to listen for SIGWINCH (Window Change)
	signal.Notify(sigChan, syscall.SIGWINCH, syscall.SIGINT)

	// Print the initial dimensions immediately
	printTerminalSize(fd)

	// Listen for resize events in a blocking loop
	for sig := range sigChan {
		if sig == syscall.SIGINT {
			signal.Reset()
			// Ensure the cursor is restored when the program exits
			fmt.Print("\033[?25h")
			break
		}
		printTerminalSize(fd)
	}
}

func printTerminalSize(fd int) {
	width, height, err := term.GetSize(fd)
	if err != nil {
		log.Printf("Error getting terminal size: %v\n", err)
		return
	}
	// \r returns the cursor to the beginning of the line for clean updates
	fmt.Printf("\rTerminal resized: width = %d, height = %d   ", width, height)
}
