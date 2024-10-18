package testutils

import (
	"fmt"
	"strings"
)

// CheckError checks the given error taking into account if it was expected and
// potentially the message it should carry.
func CheckError(err error, expected bool, msg string) error {
	if !expected && err != nil {
		return fmt.Errorf("unexpected error %w", err)
	}
	if expected {
		if err == nil {
			return fmt.Errorf("expected error didn't happen")
		}
		if msg == "" {
			return fmt.Errorf("expected error message not defined")
		}
		if !strings.HasPrefix(err.Error(), msg) {
			return fmt.Errorf("wrong error, expected %q, got %q", msg, err.Error())
		}
	}
	return nil
}
