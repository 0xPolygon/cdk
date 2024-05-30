package version

import (
	"fmt"
	"io"
	"runtime"
)

var (
	Version = "v0.1.0"
)

// PrintVersion prints version info into the provided io.Writer.
func PrintVersion(w io.Writer) {
	fmt.Fprintf(w, "Version:      %s\n", Version)
	fmt.Fprintf(w, "Go version:   %s\n", runtime.Version())
	fmt.Fprintf(w, "OS/Arch:      %s/%s\n", runtime.GOOS, runtime.GOARCH)
}
