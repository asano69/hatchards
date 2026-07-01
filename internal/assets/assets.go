// internal/assets/assets.go
package assets

import (
	"embed"
	"io/fs"
)

//go:embed static
var rawStatic embed.FS

// FS is the embedded static tree (CSS, JS, KaTeX, favicon, HTML shells),
// served publicly under /static/.
var FS fs.FS

func init() {
	sub, err := fs.Sub(rawStatic, "static")
	if err != nil {
		// Only fails if the embed directive itself is wrong; that's a
		// build-time bug, not a runtime condition.
		panic(err)
	}
	FS = sub
}

// Sub returns the embedded static subtree rooted at dir (e.g. "katex").
func Sub(dir string) fs.FS {
	sub, err := fs.Sub(FS, dir)
	if err != nil {
		panic(err)
	}
	return sub
}
