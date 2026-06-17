package codegen

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
)

// EnsureUpToDate checks whether schema.graphqls has changed since the last
// codegen run (via a SHA-256 sentinel file). If it has—or if the sentinel is
// absent—it calls runner() to regenerate, then updates the sentinel.
// Returns a non-nil error if runner() fails; the caller should treat this as
// fatal (server cannot start with stale generated code).
func EnsureUpToDate(schemaPath, sentinelPath string, runner func() error) error {
	contents, err := os.ReadFile(schemaPath)
	if err != nil {
		return err
	}

	hash := sha256.Sum256(contents)
	currentHex := hex.EncodeToString(hash[:])

	stored := ""
	if data, err := os.ReadFile(sentinelPath); err == nil {
		stored = string(data)
	}

	if currentHex == stored {
		return nil
	}

	if err := runner(); err != nil {
		return err
	}

	return os.WriteFile(sentinelPath, []byte(currentHex), 0o644)
}

// DefaultRunner returns a runner that invokes gqlgen generate.
func DefaultRunner() func() error {
	return func() error {
		cmd := exec.Command("go", "run", "github.com/99designs/gqlgen", "generate")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
}
