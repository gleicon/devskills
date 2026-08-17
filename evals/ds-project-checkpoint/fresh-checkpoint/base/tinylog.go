// Package tinylog is a minimal structured logger.
package tinylog

import "fmt"

// Log writes a message to standard output.
func Log(msg string) {
	fmt.Println(msg)
}
