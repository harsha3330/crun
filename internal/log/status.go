package logger

import (
	"fmt"
	"os"
)

type Console struct{}

func (Console) Status(msg string, args ...any) {
	fmt.Fprintf(os.Stdout, "• "+msg+"\n", args...)
}

func (Console) Step(msg string, args ...any) {
	fmt.Fprintf(os.Stdout, "→ "+msg+"\n", args...)
}

func (Console) Success(msg string, args ...any) {
	fmt.Fprintf(os.Stdout, "✔ "+msg+"\n", args...)
}

func (Console) Warn(msg string, args ...any) {
	fmt.Fprintf(os.Stdout, "⚠ "+msg+"\n", args...)
}

func (Console) Error(msg string, args ...any) {
	fmt.Fprintf(os.Stderr, "✖ "+msg+"\n", args...)
}
