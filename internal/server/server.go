package server

import (
	"io/fs"
	"net"
	"net/http"
)

// New serves the SPA and its generated assets from the embedded web build.
func New(assets fs.FS) (http.Handler, error) {
	dist, err := fs.Sub(assets, "dist")
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(dist)))
	return mux, nil
}

// HTTPServer creates the HTTP server used by the command entry point.
func HTTPServer(listener net.Listener, handler http.Handler) *http.Server {
	return &http.Server{Handler: handler}
}
