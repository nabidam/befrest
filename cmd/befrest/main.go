package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/nabidam/befrest/internal/server"
	"github.com/nabidam/befrest/web"
)

const defaultPort = 5311

func main() {
	port, err := configuredPort()
	if err != nil {
		slog.Error("invalid port configuration", "err", err)
		os.Exit(2)
	}

	listener, boundPort, err := listenFrom(port)
	if err != nil {
		slog.Error("listen failed", "err", err)
		os.Exit(1)
	}
	defer listener.Close()

	handler, err := server.New(web.Files)
	if err != nil {
		slog.Error("create server", "err", err)
		os.Exit(1)
	}

	httpServer := server.HTTPServer(listener, handler)
	url := fmt.Sprintf("http://localhost:%d", boundPort)
	slog.Info("befrest listening", "url", url)

	go func() {
		if err := httpServer.Serve(listener); err != nil {
			slog.Error("server stopped", "err", err)
		}
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	<-signals
}

func configuredPort() (int, error) {
	defaultValue := strconv.Itoa(defaultPort)
	if envPort := os.Getenv("BEFREST_PORT"); envPort != "" {
		defaultValue = envPort
	}

	portValue := flag.String("port", defaultValue, "port to listen on")
	flag.Parse()

	port, err := strconv.Atoi(*portValue)
	if err != nil || port < 1 || port > 65535 {
		return 0, fmt.Errorf("port must be between 1 and 65535")
	}
	return port, nil
}

func listenFrom(port int) (net.Listener, int, error) {
	for candidate := port; candidate <= 65535; candidate++ {
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", candidate))
		if err == nil {
			return listener, candidate, nil
		}
	}
	return nil, 0, fmt.Errorf("no free port found at or above %d", port)
}

var _ fs.FS = web.Files
