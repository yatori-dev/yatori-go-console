// Package frontend embeds the Next.js static export so the console can be
// shipped as a single binary.
package frontend

import (
	"embed"
	"io/fs"
)

// The all: prefix is required: Next.js writes its assets into _next/,
// and a plain //go:embed out would skip files and directories starting
// with "_" or ".".
//
//go:embed all:out
var embedded embed.FS

// FS returns the embedded frontend rooted at out/.
func FS() fs.FS {
	sub, err := fs.Sub(embedded, "out")
	if err != nil {
		panic("embedded frontend is missing the out directory: " + err.Error())
	}
	return sub
}

// Available reports whether a usable frontend was embedded at build time.
func Available() bool {
	if _, err := fs.Stat(FS(), "index.html"); err != nil {
		return false
	}
	return true
}
