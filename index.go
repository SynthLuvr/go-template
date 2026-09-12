// Package gotemplate is a trivial starter module — replace it with your
// code.
package gotemplate

import "fmt"

// Greet returns a greeting.
func Greet() string {
	return "hello"
}

// ToLabel labels a string or an int.
func ToLabel[T string | int](value T) string {
	return fmt.Sprintf("label: %v", value)
}
