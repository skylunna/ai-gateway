// Package console serves the embedded Web Console static assets.
// Run `make build-web` to replace dist/ with the production frontend build.
package console

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed all:dist
var distFS embed.FS

// Handler returns an http.Handler that serves the embedded frontend.
// Any path without a file extension (e.g. /traces, /dashboard) returns
// index.html so the React SPA can handle client-side routing.
func Handler() http.Handler {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		panic("console: failed to sub dist FS: " + err.Error())
	}
	fileServer := http.FileServer(http.FS(sub))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Serve index.html for SPA routes (paths without a dot in the last segment).
		if !strings.Contains(r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:], ".") {
			r.URL.Path = "/"
		}
		fileServer.ServeHTTP(w, r)
	})
}
