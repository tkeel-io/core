package print //nolint

import (
	"fmt"
	"io"
	"runtime"
)

const (
	windowsOS = "windows"
)

// SuccessStatusEvent reports on a success event.
func SuccessStatusEvent(w io.Writer, fmtstr string, a ...interface{}) {
	if runtime.GOOS == windowsOS {
		fmt.Fprintf(w, "%s\n", fmt.Sprintf(fmtstr, a...))
	} else {
		fmt.Fprintf(w, "✅  %s\n", fmt.Sprintf(fmtstr, a...))
	}
}

// FailureStatusEvent reports on a failure event.
func FailureStatusEvent(w io.Writer, fmtstr string, a ...interface{}) {
	if runtime.GOOS == windowsOS {
		fmt.Fprintf(w, "%s\n", fmt.Sprintf(fmtstr, a...))
	} else {
		fmt.Fprintf(w, "❌  %s\n", fmt.Sprintf(fmtstr, a...))
	}
}

// WarningStatusEvent reports on a failure event.
func WarningStatusEvent(w io.Writer, fmtstr string, a ...interface{}) {
	if runtime.GOOS == windowsOS {
		fmt.Fprintf(w, "%s\n", fmt.Sprintf(fmtstr, a...))
	} else {
		fmt.Fprintf(w, "⚠  %s\n", fmt.Sprintf(fmtstr, a...))
	}
}

// PendingStatusEvent reports on a pending event.
func PendingStatusEvent(w io.Writer, fmtstr string, a ...interface{}) {
	if runtime.GOOS == windowsOS {
		fmt.Fprintf(w, "%s\n", fmt.Sprintf(fmtstr, a...))
	} else {
		fmt.Fprintf(w, "⌛  %s\n", fmt.Sprintf(fmtstr, a...))
	}
}

// InfoStatusEvent reports status information on an event.
func InfoStatusEvent(w io.Writer, fmtstr string, a ...interface{}) {
	if runtime.GOOS == windowsOS {
		fmt.Fprintf(w, "%s\n", fmt.Sprintf(fmtstr, a...))
	} else {
		fmt.Fprintf(w, "ℹ️  %s\n", fmt.Sprintf(fmtstr, a...))
	}
}
