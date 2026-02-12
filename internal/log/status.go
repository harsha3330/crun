package logger

import (
	"fmt"
	"os"
	"strings"
)

type Console struct{}

func (c Console) Status(msg string, args ...any)  { c.print(os.Stdout, "•", msg, args...) }
func (c Console) Step(msg string, args ...any)    { c.print(os.Stdout, "→", msg, args...) }
func (c Console) Success(msg string, args ...any) { c.print(os.Stdout, "✔", msg, args...) }
func (c Console) Warn(msg string, args ...any)    { c.print(os.Stdout, "⚠", msg, args...) }
func (c Console) Error(msg string, args ...any)   { c.print(os.Stderr, "✖", msg, args...) }

func (Console) print(out *os.File, prefix, msg string, args ...any) {
	if len(args) == 0 {
		fmt.Fprintln(out, prefix, msg)
		return
	}

	// Build key=value pairs
	var b strings.Builder
	b.WriteString(msg)

	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			b.WriteString(fmt.Sprintf(" %v=%v", args[i], args[i+1]))
		} else {
			// odd number of args
			b.WriteString(fmt.Sprintf(" %v", args[i]))
		}
	}

	fmt.Fprintln(out, prefix, b.String())
}
