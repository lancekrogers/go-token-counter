package ui

import (
	"fmt"
	"os"
)

// UI handles output display.
type UI struct {
	noColor bool
	verbose bool
}

// New creates a new UI instance.
func New(noColor, verbose bool) *UI {
	return &UI{
		noColor: noColor,
		verbose: verbose,
	}
}

// Info prints an info message.
func (u *UI) Info(format string, args ...any) {
	fmt.Fprintf(os.Stdout, format+"\n", args...)
}

// Success prints a success message.
func (u *UI) Success(format string, args ...any) {
	fmt.Fprintf(os.Stdout, format+"\n", args...)
}

// Warning prints a warning message to stderr.
func (u *UI) Warning(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "Warning: "+format+"\n", args...)
}

// Error prints an error message to stderr.
func (u *UI) Error(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
}

// Verbose returns whether verbose mode is enabled.
func (u *UI) Verbose() bool {
	return u.verbose
}
