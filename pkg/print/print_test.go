package print //nolint

import (
	"os"
	"testing"
)

func TestPrint(t *testing.T) {
	InfoStatusEvent(os.Stdout, "test InfoStatusEvent")
	SuccessStatusEvent(os.Stdout, "test SuccessStatusEvent")
	FailureStatusEvent(os.Stdout, "test FailureStatusEvent")
	WarningStatusEvent(os.Stdout, "test WarningStatusEvent")
	PendingStatusEvent(os.Stdout, "test PendingStatusEvent")
}
