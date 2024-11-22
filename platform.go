package selfupdater

import (
	"fmt"
	"runtime"
)

func getPlatform() string {
	osMap := map[string]string{
		"darwin":  "Darwin",
		"linux":   "Linux",
		"windows": "Windows",
	}

	archMap := map[string]string{
		"amd64": "x86_64",
		"386":   "i386",
		"arm64": "arm64",
		"arm":   "arm",
	}

	return fmt.Sprintf("%s_%s", osMap[runtime.GOOS], archMap[runtime.GOARCH])
}
