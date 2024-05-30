package erigondriver

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

// ErigonDriver is a driver for the Erigon client
type ErigonDriver struct {
	// Path to the erigon binary
	Path string
}

// NewErigonDriver creates a new ErigonDriver
func NewErigonDriver(path string) *ErigonDriver {
	return &ErigonDriver{
		Path: path,
	}
}

// Start starts the Erigon client
func (e *ErigonDriver) Start(ctx context.Context) error {
	cmd := exec.Command(e.Path)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start erigon: %w", err)
	}

	return nil
}
