package main

import (
	"fmt"
	"os"
)

var (
	LogVerboseEnabled = false
	LogPrettyEnabled  = false
)

func LogError(format string, a ...interface{}) {
	_, _ = fmt.Fprintln(os.Stderr, fmt.Sprintf(format, a...))
}

func LogInfo(format string, a ...interface{}) {
	_, _ = fmt.Fprintln(os.Stdout, fmt.Sprintf(format, a...))
}

func LogVerbose(format string, a ...interface{}) {
	if LogVerboseEnabled {
		_, _ = fmt.Fprintln(os.Stdout, fmt.Sprintf(format, a...))
	}
}

func Pretty(text string) string {
	if LogPrettyEnabled {
		return text
	}
	return ""
}
