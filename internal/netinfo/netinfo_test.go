package netinfo

import (
	"net"
	"testing"
)

func TestRankFiltersAndRanksPrivateIPv4Interfaces(t *testing.T) {
	result := Rank([]Interface{
		{ID: "lo", Up: true, Loopback: true, Addresses: []net.IP{net.ParseIP("127.0.0.1")}},
		{ID: "down0", Up: false, Addresses: []net.IP{net.ParseIP("192.168.1.2")}},
		{ID: "public0", Up: true, Addresses: []net.IP{net.ParseIP("8.8.8.8")}},
		{ID: "docker0", Up: true, Addresses: []net.IP{net.ParseIP("172.17.0.1")}},
		{ID: "wlan0", Up: true, Addresses: []net.IP{net.ParseIP("192.168.1.30")}},
	}, "")

	if len(result.Candidates) != 2 {
		t.Fatalf("candidate count = %d, want 2", len(result.Candidates))
	}
	if got := result.Candidates[0]; got.ID != "wlan0" || got.Kind != "physical" || got.Address != "192.168.1.30" {
		t.Fatalf("first candidate = %#v, want wlan0 physical 192.168.1.30", got)
	}
	if result.Winner == nil || result.Winner.ID != "wlan0" {
		t.Fatalf("winner = %#v, want wlan0", result.Winner)
	}
}

func TestRankDefaultRouteBreaksTie(t *testing.T) {
	result := Rank([]Interface{
		{ID: "wlan0", Up: true, Addresses: []net.IP{net.ParseIP("192.168.1.30")}},
		{ID: "eth0", Up: true, Addresses: []net.IP{net.ParseIP("10.0.0.20")}},
	}, "eth0")

	if result.Winner == nil || result.Winner.ID != "eth0" {
		t.Fatalf("winner = %#v, want eth0", result.Winner)
	}
	if result.Ambiguous {
		t.Fatal("result unexpectedly ambiguous")
	}
}

func TestRankReportsAmbiguousTopCandidates(t *testing.T) {
	result := Rank([]Interface{
		{ID: "eth0", Up: true, Addresses: []net.IP{net.ParseIP("10.0.0.20")}},
		{ID: "wlan0", Up: true, Addresses: []net.IP{net.ParseIP("192.168.1.30")}},
	}, "")

	if !result.Ambiguous {
		t.Fatal("result should be ambiguous")
	}
	if result.Winner != nil {
		t.Fatalf("winner = %#v, want nil", result.Winner)
	}
	if result.Candidates[0].ID != "eth0" {
		t.Fatalf("top-ranked candidate = %q, want deterministic eth0", result.Candidates[0].ID)
	}
}
