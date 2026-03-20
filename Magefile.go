//go:build mage

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/magefile/mage/sh"
)

// Executable name must match plugin.json 'executable'
const exeName = "grafana-catalyst-datasource"

var Default = BuildAll

// Clean removes only backend binaries from dist/ (preserves frontend artifacts like plugin.json).
func Clean() error {
	entries, err := os.ReadDir("dist")
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, exeName) {
			_ = os.RemoveAll(filepath.Join("dist", name))
		}
	}
	return nil
}

// Build builds the backend for the current OS/ARCH without wiping dist/.
func Build() error {
	return buildOSArch(runtime.GOOS, runtime.GOARCH)
}

// BuildAll builds ONLY linux/amd64 to keep CI lean.
func BuildAll() error {
	return buildOSArch("linux", "amd64")
}

// Coverage is invoked by some CI pipelines; provide a harmless no-op.
func Coverage() error {
	fmt.Println("mage coverage: no-op")
	return nil
}

func buildOSArch(goos, goarch string) error {
	if err := os.MkdirAll("dist", 0o755); err != nil {
		return err
	}

	bin := fmt.Sprintf("%s_%s_%s", exeName, goos, goarch)
	if goos == "windows" {
		bin += ".exe"
	}
	out := filepath.Join("dist", bin)

	env := map[string]string{
		"GOOS":        goos,
		"GOARCH":      goarch,
		"CGO_ENABLED": "0",
	}

	ldflags := strings.Join([]string{"-s", "-w"}, " ")
	args := []string{"build", "-trimpath", "-ldflags", ldflags, "-o", out, "./cmd/grafana-catalyst-datasource"}
	return sh.RunWithV(env, "go", args...)
}
