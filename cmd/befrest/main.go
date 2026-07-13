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
	"sync"
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
	interfaceID := flag.String("interface", "", "network interface to advertise")
	hostNameFlag := flag.String("name", "", "host device display name")
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
	if *hostNameFlag != "" {
		hostName = *hostNameFlag
	}
	hostToken, err := mintHostToken()
	if err != nil {
		slog.Error("mint host token", "err", err)
		os.Exit(1)
	}
	discovery, address, err := advertisedNetwork(*interfaceID)
	if err != nil {
		slog.Error("select LAN address", "err", err)
		os.Exit(2)
	}
	reachabilityHint := ""
	if err := netinfo.Probe(address, boundPort); err != nil {
		reachabilityHint = fmt.Sprintf("If scanning doesn't work, check your firewall allows befrest on port %d.", boundPort)
		slog.Warn("advertised address may be unreachable", "address", address, "port", boundPort, "err", err)
	}
	invite := serverInvite(address, boundPort)
	invite.ReachabilityHint = reachabilityHint

	var mdnsMu sync.Mutex
	var mdns *zeroconf.Server
	announce := func(nextAddress string) error {
		if *noMDNS {
			return nil
		}
		mdnsMu.Lock()
		defer mdnsMu.Unlock()
		if mdns != nil {
			mdns.Shutdown()
			mdns = nil
		}
		registered, err := zeroconf.RegisterProxy("befrest", "_http._tcp", "local.", boundPort, "befrest.local.", []string{nextAddress}, nil, nil)
		if err != nil {
			return err
		}
		mdns = registered
		return nil
	}
	choices := wireChoices(discovery, *interfaceID == "")
	handler, err := server.New(web.Files, server.Config{
		HostToken: hostToken, HostName: hostName, Invite: invite, InterfaceChoices: choices,
		PickInterface: func(id string) (proto.InviteInfo, error) {
			candidate, ok := netinfo.CandidateByID(discovery.Candidates, id)
			if !ok {
				return proto.InviteInfo{}, fmt.Errorf("unknown interface %q", id)
			}
			if err := announce(candidate.Address); err != nil {
				slog.Error("re-announce mDNS", "err", err)
			}
			next := serverInvite(candidate.Address, boundPort)
			next.ReachabilityHint = reachabilityHint
			return next, nil
		},
	})
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

	if !*noMDNS {
		if err := announce(address); err != nil {
			slog.Error("start mDNS", "err", err)
		}
		defer func() {
			if mdns != nil {
				mdns.Shutdown()
			}
		}()
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

func advertisedNetwork(interfaceID string) (netinfo.Result, string, error) {
	result, err := netinfo.Discover()
	if err != nil {
		return netinfo.Result{}, "", err
	}
	if interfaceID != "" {
		candidate, ok := netinfo.CandidateByID(result.Candidates, interfaceID)
		if !ok {
			return result, "", fmt.Errorf("interface %q has no private IPv4 address", interfaceID)
		}
		return result, candidate.Address, nil
	}
	if len(result.Candidates) > 0 {
		return result, result.Candidates[0].Address, nil
	}
	slog.Warn("no private LAN address found; invites are local-only")
	return result, "127.0.0.1", nil
}

func wireChoices(result netinfo.Result, ambiguous bool) []proto.InterfaceChoice {
	if !ambiguous || !result.Ambiguous {
		return nil
	}
	choices := make([]proto.InterfaceChoice, len(result.Candidates))
	for i, candidate := range result.Candidates {
		choices[i] = proto.InterfaceChoice{ID: candidate.ID, Kind: candidate.Kind, Address: candidate.Address}
	}
	return choices
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
