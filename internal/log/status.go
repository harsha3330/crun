package logger

import (
	"fmt"
	"os"
)

func Status(msg string, args ...any) {
	fmt.Fprintf(os.Stdout, msg+"\n", args...)
}

func Success(msg string, args ...any) {
	fmt.Fprintf(os.Stdout, "✔ "+msg+"\n", args...)
}

func Warn(msg string, args ...any) {
	fmt.Fprintf(os.Stdout, "⚠ "+msg+"\n", args...)
}

func Error(msg string, args ...any) {
	fmt.Fprintf(os.Stderr, "✖ "+msg+"\n", args...)
}
