package main

import (
	"fmt"
	"os"
	"strings"
)

var (
	LogVerboseEnabled = false
	LogPrettyEnabled  = false
)

func LogError(format string, a ...interface{}) {
	// Remove "shrek: " from error messages before outputting them.
	for i, v := range a {
		if err, ok := v.(error); ok && err != nil {
			a[i] = strings.ReplaceAll(err.Error(), "shrek: ", "")
		}
	}

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

func Pretty(text, alt string) string {
	if LogPrettyEnabled {
		return text
	}
	return alt
}
