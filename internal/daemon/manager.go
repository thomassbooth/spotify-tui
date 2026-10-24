package daemon

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Manager owns the go-librespot subprocess lifecycle.
type Manager struct {
	cmd *exec.Cmd
}

// candidatePaths returns locations to search for the go-librespot binary.
func candidatePaths() []string {
	home, _ := os.UserHomeDir()
	return []string{
		"go-librespot",
		filepath.Join(home, ".local", "bin", "go-librespot"),
		filepath.Join(home, "go", "bin", "go-librespot"),
		"./go-librespot",
	}
}

// FindBinary returns the path to the go-librespot binary, or "" if not found.
func FindBinary() string {
	for _, p := range candidatePaths() {
		if path, err := exec.LookPath(p); err == nil {
			return path
		}
	}
	return ""
}

// Start launches go-librespot as a subprocess and returns a Manager.
func Start(ctx context.Context, binaryPath string) (*Manager, error) {
	cmd := exec.CommandContext(ctx, binaryPath)
	cmd.Stdout = os.Stderr // forward daemon logs to our stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start go-librespot: %w", err)
	}
	return &Manager{cmd: cmd}, nil
}

// WaitReady polls ping until go-librespot's HTTP API responds or ctx expires.
func WaitReady(ctx context.Context, ping func(context.Context) error) error {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("go-librespot did not become ready: %w", ctx.Err())
		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
			err := ping(pingCtx)
			cancel()
			if err == nil {
				return nil
			}
		}
	}
}

// Stop kills the subprocess.
func (m *Manager) Stop() {
	if m.cmd != nil && m.cmd.Process != nil {
		m.cmd.Process.Kill()
	}
}
