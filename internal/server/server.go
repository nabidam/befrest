package server

import (
	"io/fs"
	"net"
	"net/http"

	"github.com/nabidam/befrest/internal/proto"
)

// Config supplies the launch-time details that the control socket shares with
// clients. Empty values keep New suitable for isolated server tests.
type Config struct {
	HostToken string
	HostName  string
	Invite    proto.InviteInfo
}

// New serves the SPA and its generated assets from the embedded web build.
func New(assets fs.FS, configs ...Config) (http.Handler, error) {
	dist, err := fs.Sub(assets, "dist")
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	config := Config{}
	if len(configs) > 0 {
		config = configs[0]
	}
	hub := newWebSocketHub(config)
	mux.HandleFunc("/ws", hub.serveWS)
	mux.HandleFunc("/api/transfers/", hub.serveFiles)
	mux.Handle("/", http.FileServer(http.FS(dist)))
	return mux, nil
}

// HTTPServer creates the HTTP server used by the command entry point.
func HTTPServer(listener net.Listener, handler http.Handler) *http.Server {
	return &http.Server{Handler: handler}
}
