// Package mirror clones or pulls a Git repository into a local directory,
// using go-git. This is the only package that touches go-git directly;
// callers supply plain connection data plus a decrypted token.
package mirror

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	transport "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/sirupsen/logrus"

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

// isGitRepo reports whether path contains a ".git" entry, i.e. whether it is
// the root of an existing git working copy. This is deliberately stricter
// than "does the directory exist", because LocalPath may already exist as
// an empty directory (e.g. created by a volume mount) without ever having
// been cloned into.
func isGitRepo(path string) bool {
	info, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil && (info.IsDir() || info.Mode().IsRegular())
}

// Sync brings LocalPath up to date with RemoteURL: it clones the repository
// if LocalPath doesn't exist yet (or exists but isn't a git repo), or pulls
// into an existing clone otherwise.
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

	log := logrus.WithFields(logrus.Fields{
		"connection": conn.Name,
		"remote_url": conn.RemoteURL,
		"username":   conn.Username,
		"local_path": conn.LocalPath,
	})

	info, statErr := os.Stat(conn.LocalPath)
	switch {
	case os.IsNotExist(statErr):
		log.Info("mirror sync: local_path does not exist, cloning")
		return clone(conn, auth, log)

	case statErr != nil:
		log.WithError(statErr).Error("mirror sync: failed to stat local_path")
		return errs.Newf("stat local path %s: %v", conn.LocalPath, statErr)

	case !info.IsDir():
		log.Error("mirror sync: local_path exists but is not a directory")
		return errs.Newf("local path %s exists but is not a directory", conn.LocalPath)

	case !isGitRepo(conn.LocalPath):
		// local_path exists (e.g. an empty directory) but was never cloned
		// into. Cloning into an existing *empty* directory is fine for
		// go-git/git; a non-empty, non-repo directory will fail clone with
		// a clear error instead of the confusing "repository does not
		// exist" that PlainOpen used to produce here.
		log.Warn("mirror sync: local_path exists but has no .git, cloning into it")
		return clone(conn, auth, log)

	default:
		log.Info("mirror sync: local_path is an existing git repo, pulling")
		return pull(conn, auth, log)
	}
}

func clone(conn Connection, auth *transport.BasicAuth, log *logrus.Entry) error {
	_, err := git.PlainClone(conn.LocalPath, false, &git.CloneOptions{
		URL:  conn.RemoteURL,
		Auth: auth,
	})
	if err != nil {
		log.WithError(err).Error("mirror sync: clone failed")
		return errs.Newf("clone %q: %v", conn.Name, err)
	}
	log.Info("mirror sync: clone succeeded")
	return nil
}

func pull(conn Connection, auth *transport.BasicAuth, log *logrus.Entry) error {
	repo, err := git.PlainOpen(conn.LocalPath)
	if err != nil {
		log.WithError(err).Error("mirror sync: PlainOpen failed")
		return errs.Newf("open local repo for %q: %v", conn.Name, err)
	}

	// Guard against local_path being shared by two different connections:
	// if the repo's configured "origin" doesn't match this connection's
	// RemoteURL, pulling would silently send this connection's credentials
	// to whatever remote is actually configured — producing confusing
	// errors that look like an auth problem with the wrong remote.
	remote, err := repo.Remote("origin")
	if err != nil {
		log.WithError(err).Error("mirror sync: could not read origin remote")
		return errs.Newf("read origin remote for %q: %v", conn.Name, err)
	}
	actualURLs := remote.Config().URLs
	if len(actualURLs) == 0 || actualURLs[0] != conn.RemoteURL {
		log.WithField("origin_url", actualURLs).
			Error("mirror sync: local_path's origin does not match this connection's remote_url — local_path is likely shared by another connection")
		return errs.Newf(
			"local path %s is a git repo whose origin (%v) does not match this connection's remote_url (%s) — check that local_path is not shared with another connection",
			conn.LocalPath, actualURLs, conn.RemoteURL,
		)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		log.WithError(err).Error("mirror sync: get worktree failed")
		return errs.Newf("get worktree for %q: %v", conn.Name, err)
	}
	err = worktree.Pull(&git.PullOptions{RemoteName: "origin", Auth: auth})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		log.WithError(err).Error("mirror sync: pull failed")
		return errs.Newf("pull %q: %v", conn.Name, err)
	}
	if errors.Is(err, git.NoErrAlreadyUpToDate) {
		log.Info("mirror sync: pull succeeded (already up to date)")
	} else {
		log.Info("mirror sync: pull succeeded")
	}
	return nil
}
