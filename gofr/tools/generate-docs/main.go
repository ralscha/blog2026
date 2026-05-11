package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

const swagVersion = "v2.0.0-rc5"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cmd := exec.Command(
		"go",
		"run",
		"github.com/swaggo/swag/v2/cmd/swag@"+swagVersion,
		"init",
		"--generalInfo",
		"main.go",
		"--dir",
		".",
		"--output",
		"docs",
		"--outputTypes",
		"json,yaml",
		"--v3.1",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("generate OpenAPI docs: %w", err)
	}

	if err := copyFile(filepath.Join("docs", "swagger.json"), filepath.Join("static", "openapi.json")); err != nil {
		return fmt.Errorf("sync generated spec: %w", err)
	}

	return nil
}

func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	if err = os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	target, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer target.Close()

	if _, err = io.Copy(target, source); err != nil {
		return err
	}

	return target.Close()
}
