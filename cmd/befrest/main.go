package main

import (
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/grandcat/zeroconf"
	"github.com/nabidam/befrest/internal/netinfo"
	"github.com/nabidam/befrest/internal/proto"
	"github.com/nabidam/befrest/internal/server"
	"github.com/nabidam/befrest/web"
	"github.com/pkg/browser"
)

const defaultPort = 5311

func main() {
	logPath, closeLog := configureLogger()
	defer closeLog()
	noOpen := flag.Bool("no-open", false, "do not open Befrest in a browser")
	noMDNS := flag.Bool("no-mdns", false, "do not announce Befrest on the local network")
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

	hostName, err := os.Hostname()
	if err != nil {
		slog.Error("read hostname", "err", err)
		os.Exit(1)
	}
	hostToken, err := mintHostToken()
	if err != nil {
		slog.Error("mint host token", "err", err)
		os.Exit(1)
	}
	address := advertisedAddress()
	invite := serverInvite(address, boundPort)
	handler, err := server.New(web.Files, server.Config{HostToken: hostToken, HostName: hostName, Invite: invite})
	if err != nil {
		slog.Error("create server", "err", err)
		os.Exit(1)
	}

	httpServer := server.HTTPServer(listener, handler)
	hostURL := fmt.Sprintf("%s/?hostToken=%s", invite.URLs.IP, hostToken)
	slog.Info("befrest listening", "url", invite.URLs.IP)

	go func() {
		if err := httpServer.Serve(listener); err != nil {
			slog.Error("server stopped", "err", err)
		}
	}()

	var mdns *zeroconf.Server
	if !*noMDNS {
		mdns, err = zeroconf.RegisterProxy("befrest", "_http._tcp", "local.", boundPort, "befrest.local.", []string{address}, nil, nil)
		if err != nil {
			slog.Error("start mDNS", "err", err)
		} else {
			defer mdns.Shutdown()
		}
	}

	quit := make(chan struct{})
	if trayAvailable() {
		startTray(func() {
			if err := browser.OpenURL(hostURL); err != nil {
				slog.Error("open browser", "err", err)
			}
		}, func() {
			if err := browser.OpenFile(logPath); err != nil {
				slog.Error("open log", "err", err)
			}
		}, func() { close(quit) })
	} else {
		slog.Warn("system tray unavailable; running without tray")
	}
	if !*noOpen {
		if err := browser.OpenURL(hostURL); err != nil {
			slog.Error("open browser", "err", err)
		}
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	select {
	case <-signals:
	case <-quit:
	}
}

func configureLogger() (string, func()) {
	path := filepath.Join(os.TempDir(), "befrest.log")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return path, func() {}
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(io.MultiWriter(os.Stderr, file), nil)))
	return path, func() { _ = file.Close() }
}

func mintHostToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func advertisedAddress() string {
	result, err := netinfo.Discover()
	if err != nil {
		slog.Error("discover LAN address", "err", err)
		return "127.0.0.1"
	}
	if len(result.Candidates) > 0 {
		return result.Candidates[0].Address
	}
	slog.Warn("no private LAN address found; invites are local-only")
	return "127.0.0.1"
}

func serverInvite(address string, port int) proto.InviteInfo {
	return proto.InviteInfo{
		Type: proto.MsgInviteInfo,
		URLs: proto.InviteURLs{
			MDNS: fmt.Sprintf("http://befrest.local:%d", port),
			IP:   fmt.Sprintf("http://%s:%d", address, port),
		},
		Port: port,
	}
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
