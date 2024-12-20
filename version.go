package cdk

import (
	"fmt"
	"io"
	"runtime"
)

// Populated during build, don't touch!
var (
	Version   = "v0.1.0"
	GitRev    = "undefined"
	GitBranch = "undefined"
	BuildDate = "Fri, 17 Jun 1988 01:58:00 +0200"
)

// PrintVersion prints version info into the provided io.Writer.
func PrintVersion(w io.Writer) {
	data := GetVersion()
	fmt.Fprintf(w, "%s", data.String())
}

type FullVersion struct {
	Version   string
	GitRev    string
	GitBranch string
	BuildDate string
	GoVersion string
	OS        string
	Arch      string
}

func GetVersion() FullVersion {
	return FullVersion{
		Version:   Version,
		GitRev:    GitRev,
		GitBranch: GitBranch,
		BuildDate: BuildDate,
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}

func (f FullVersion) String() string {
	return fmt.Sprintf("Version:      %s\n"+
		"Git revision: %s\n"+
		"Git branch:   %s\n"+
		"Go version:   %s\n"+
		"Built:        %s\n"+
		"OS/Arch:      %s/%s\n",
		f.Version, f.GitRev, f.GitBranch,
		f.GoVersion, f.BuildDate, f.OS, f.Arch)
}
