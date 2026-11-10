// Package web embeds the browser pages and runtime assets.
package web

import "embed"

// Files contains only runtime pages and assets, relative to this package.
//
//go:embed pages/*.html static/js/*.js static/css/*.css
var Files embed.FS

// MustRead reads a build-time embedded resource. An invalid path is a programmer error.
func MustRead(path string) []byte {
	contents, err := Files.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return contents
}
