// Package tinylog is a minimal structured logger.
package tinylog

import "fmt"

// Log writes a leveled message to standard output.
func Log(level, msg string) {
	fmt.Printf("[%s] %s\n", level, msg)
}
