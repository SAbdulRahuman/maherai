// Package ui embeds the Next.js static export for serving the web UI.
package ui

import (
	"embed"
	"io/fs"
)

//go:embed all:static
var staticFS embed.FS

// Static returns the embedded filesystem rooted at the "static" directory.
// This contains the Next.js `output: "export"` build output.
func Static() (fs.FS, error) {
	return fs.Sub(staticFS, "static")
}
