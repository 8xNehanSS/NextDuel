package middleware

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// SPA serves a built single-page app from root. Existing files are served
// directly; any other path gets index.html so client-side routes deep-link.
// Vite's hashed /assets/* are cached forever; index.html is never cached so
// deploys take effect immediately.
func SPA(root string) http.Handler {
	files := http.FileServer(http.Dir(root))
	index := filepath.Join(root, "index.html")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clean := path.Clean("/" + r.URL.Path)
		info, err := os.Stat(filepath.Join(root, filepath.FromSlash(clean)))
		if err == nil && !info.IsDir() {
			if strings.HasPrefix(clean, "/assets/") {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			}
			files.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Cache-Control", "no-cache")
		http.ServeFile(w, r, index)
	})
}
