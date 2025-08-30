//go:build mage

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Executable name must match plugin.json 'backend.executable'
const exeName = "grafana-catalyst-datasource"

var Default = BuildAll

// Clean removes the dist directory.
func Clean() error {
	return os.RemoveAll("dist")
}

// Build builds the backend for the current OS/ARCH.
func Build() error {
	mg.Deps(Clean)
	return buildOSArch(runtime.GOOS, runtime.GOARCH)
}

// BuildAll builds the backend for a set of OS/ARCH targets used by Grafana plugins.
func BuildAll() error {
	mg.Deps(Clean)
	targets := []struct{ OS, Arch string }{
		{"linux", "amd64"},
	}
	for _, t := range targets {
		if err := buildOSArch(t.OS, t.Arch); err != nil {
			return err
		}
	}
	return nil
}

func buildOSArch(goos, goarch string) error {
	outDir := "dist"
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	bin := fmt.Sprintf("%s_%s_%s", exeName, goos, goarch)
	if goos == "windows" {
		bin += ".exe"
	}
	out := filepath.Join(outDir, bin)

	env := map[string]string{
		"GOOS":        goos,
		"GOARCH":      goarch,
		"CGO_ENABLED": "0",
	}

	ldflags := strings.Join([]string{
		"-s", "-w",
	}, " ")

	args := []string{"build", "-trimpath", "-ldflags", ldflags, "-o", out, "./cmd/grafana-catalyst-datasource"}
	return sh.RunWithV(env, "go", args...)
}

// Coverage is invoked by CI packaging; we don't need Go test coverage here.
// Provide a no-op so the packaging action doesn't fail.
func Coverage() error {
	fmt.Println("mage coverage: no-op")
	return nil
}
