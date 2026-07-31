package static

import (
	"embed"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// Dist is populated at build time via //go:embed in embed_dist.go or empty placeholder.
//
//go:embed dist/*
var distEmbed embed.FS

func Handler() http.Handler {
	sub, err := fs.Sub(distEmbed, "dist")
	if err != nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = io.WriteString(w, "frontend not embedded; build with web/dist")
		})
	}
	fileServer := http.FileServer(http.FS(sub))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// SPA fallback: if path has no extension and file missing, serve index.html
		p := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if p == "" || p == "." {
			p = "index.html"
		}
		if _, err := fs.Stat(sub, p); err != nil {
			// no extension or missing asset -> index.html for client router
			if !strings.Contains(path.Base(p), ".") {
				r2 := r.Clone(r.Context())
				r2.URL.Path = "/"
				fileServer.ServeHTTP(w, r2)
				return
			}
			http.NotFound(w, r)
			return
		}
		fileServer.ServeHTTP(w, r)
	})
}
