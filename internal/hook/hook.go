// Package hook resolves and runs pre-installed post-sync scripts. Scripts
// are never uploaded or authored through the API; they are placed on disk
// by the operator ahead of time, so the only thing a connection stores is
// a name that gets resolved against this fixed, read-only directory.
package hook

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/asano69/hashcards/internal/errs"
)

// nameRe rejects anything but a bare identifier, so a name can never
// escape hooksDir via "../" or an absolute path.
var nameRe = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// List returns available hook names, or an empty list if hooksDir doesn't
// exist (e.g. the operator hasn't configured any hooks). This keeps
// installations without a hooks directory working exactly as before.
func List(hooksDir string) ([]string, error) {
	entries, err := os.ReadDir(hooksDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, errs.Newf("read hooks dir: %v", err)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() || !nameRe.MatchString(e.Name()) {
			continue
		}
		info, err := e.Info()
		if err != nil || !info.Mode().IsRegular() || info.Mode()&0111 == 0 {
			continue
		}
		names = append(names, e.Name())
	}
	return names, nil
}

// Resolve validates name and returns the absolute path of the hook script.
// It re-checks existence/executability at call time (not just at save
// time) so hooksDir contents can change without restarting the server.
// Returns ("", nil) when name is empty, meaning "no hook configured".
func Resolve(hooksDir, name string) (string, error) {
	if name == "" {
		return "", nil
	}
	if !nameRe.MatchString(name) {
		return "", errs.Newf("invalid hook name: %q", name)
	}
	path := filepath.Join(hooksDir, name)
	info, err := os.Stat(path)
	if err != nil {
		return "", errs.Newf("hook not found: %s", name)
	}
	if !info.Mode().IsRegular() || info.Mode()&0111 == 0 {
		return "", errs.Newf("hook is not executable: %s", name)
	}
	return path, nil
}

// Run executes the resolved script directly (no shell), passing the
// source and output directories as environment variables. sourceDir is
// the connection's git working tree; outputDir is where generated JSON
// decks should be written.
func Run(ctx context.Context, scriptPath, sourceDir, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return errs.Newf("create hook output dir: %v", err)
	}
	cmd := exec.CommandContext(ctx, scriptPath)
	cmd.Dir = sourceDir
	cmd.Env = append(os.Environ(),
		"HASHCARDS_SOURCE_DIR="+sourceDir,
		"HASHCARDS_OUTPUT_DIR="+outputDir,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return errs.Newf("hook %q failed: %v\n%s", filepath.Base(scriptPath), err, out)
	}
	return nil
}
