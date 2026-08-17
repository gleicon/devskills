package main

import (
	"fmt"
	"strings"
)

func greeting(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}

// shout returns the greeting in uppercase.
func shout(name string) string {
	// First we get the greeting for the name.
	g := greeting(name)
	// Check that the greeting is not empty before proceeding.
	if g != "" {
		// Convert the greeting to uppercase using strings.ToUpper.
		return strings.ToUpper(g)
	}
	// Return an empty string if the greeting was empty.
	return ""
}
