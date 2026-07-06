// Package mirror clones or pulls a Git repository into a local directory,
// using go-git. This is the only package that touches go-git directly;
// callers supply plain connection data plus a decrypted token.
package mirror

import (
	"errors"
	"os"

	"github.com/go-git/go-git/v5"
	transport "github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/asano69/hashcards/internal/errs"
)

// Connection is the plain (non-secret) data needed to mirror one repository.
// The token is passed separately to Sync and never stored here, so a
// Connection value can be logged or included in error messages safely.
type Connection struct {
	Name      string
	RemoteURL string
	Username  string
	LocalPath string
}

// Sync brings LocalPath up to date with RemoteURL: it clones the repository
// if LocalPath doesn't exist yet, or pulls into an existing clone otherwise.
//
// token is the decrypted access token, used only for the duration of this
// call. Sync never retains a reference to it; the caller is responsible for
// zeroing it afterwards (see cryptoutil.Zero).
//
// Authentication is passed to go-git as a BasicAuth credential rather than
// embedded into the URL, so RemoteURL stays free of secrets and is safe to
// include in error messages.
func Sync(conn Connection, token []byte) error {
	auth := &transport.BasicAuth{Username: conn.Username, Password: string(token)}

	_, err := os.Stat(conn.LocalPath)
	switch {
	case os.IsNotExist(err):
		return clone(conn, auth)
	case err != nil:
		return errs.Newf("stat local path %s: %v", conn.LocalPath, err)
	default:
		return pull(conn, auth)
	}
}

func clone(conn Connection, auth *transport.BasicAuth) error {
	if _, err := git.PlainClone(conn.LocalPath, false, &git.CloneOptions{
		URL:  conn.RemoteURL,
		Auth: auth,
	}); err != nil {
		return errs.Newf("clone %q: %v", conn.Name, err)
	}
	return nil
}

func pull(conn Connection, auth *transport.BasicAuth) error {
	repo, err := git.PlainOpen(conn.LocalPath)
	if err != nil {
		return errs.Newf("open local repo for %q: %v", conn.Name, err)
	}
	worktree, err := repo.Worktree()
	if err != nil {
		return errs.Newf("get worktree for %q: %v", conn.Name, err)
	}
	err = worktree.Pull(&git.PullOptions{RemoteName: "origin", Auth: auth})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return errs.Newf("pull %q: %v", conn.Name, err)
	}
	return nil
}
