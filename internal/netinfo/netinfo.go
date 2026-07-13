// Package netinfo finds the LAN address Befrest advertises to other devices.
package netinfo

import (
	"net"
	"sort"
	"strings"
)

// Interface is an injectable view of a network interface. It keeps ranking
// deterministic in tests and separates OS inspection from selection policy.
type Interface struct {
	ID        string
	Up        bool
	Loopback  bool
	Addresses []net.IP
}

// Candidate is an address that may be advertised in an invite URL.
type Candidate struct {
	ID      string
	Kind    string
	Address string
}

// Result holds all ranked candidates. Winner is set only when the top-ranked
// candidate is unambiguous; callers that defer ambiguity handling use Candidates[0].
type Result struct {
	Candidates []Candidate
	Winner     *Candidate
	Ambiguous  bool
}

// Discover enumerates the machine's interfaces and ranks their private IPv4
// addresses. The default-route address is discovered without sending traffic.
func Discover() (Result, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return Result{}, err
	}

	fixtures := make([]Interface, 0, len(interfaces))
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			return Result{}, err
		}
		addresses := make([]net.IP, 0, len(addrs))
		for _, addr := range addrs {
			if ip, _, err := net.ParseCIDR(addr.String()); err == nil {
				addresses = append(addresses, ip)
			}
		}
		fixtures = append(fixtures, Interface{
			ID:        iface.Name,
			Up:        iface.Flags&net.FlagUp != 0,
			Loopback:  iface.Flags&net.FlagLoopback != 0,
			Addresses: addresses,
		})
	}

	return Rank(fixtures, defaultRouteID()), nil
}

// Rank filters interfaces to up, non-loopback private IPv4 candidates and
// orders physical interfaces ahead of virtual ones. A default-route interface
// breaks a tie within the same class.
func Rank(interfaces []Interface, defaultRouteID string) Result {
	type ranked struct {
		candidate Candidate
		score     int
		defaulted bool
	}

	rankedCandidates := make([]ranked, 0)
	for _, iface := range interfaces {
		if !iface.Up || iface.Loopback {
			continue
		}
		addresses := append([]net.IP(nil), iface.Addresses...)
		sort.Slice(addresses, func(i, j int) bool {
			return addresses[i].String() < addresses[j].String()
		})
		for _, ip := range addresses {
			if !isPrivateIPv4(ip) {
				continue
			}
			rankedCandidates = append(rankedCandidates, ranked{
				candidate: Candidate{ID: iface.ID, Kind: interfaceKind(iface.ID), Address: ip.String()},
				score:     interfaceScore(iface.ID),
				defaulted: iface.ID == defaultRouteID,
			})
			break // An interface is one M3 choice, even if it has several addresses.
		}
	}

	sort.Slice(rankedCandidates, func(i, j int) bool {
		if rankedCandidates[i].score != rankedCandidates[j].score {
			return rankedCandidates[i].score > rankedCandidates[j].score
		}
		if rankedCandidates[i].defaulted != rankedCandidates[j].defaulted {
			return rankedCandidates[i].defaulted
		}
		if rankedCandidates[i].candidate.ID != rankedCandidates[j].candidate.ID {
			return rankedCandidates[i].candidate.ID < rankedCandidates[j].candidate.ID
		}
		return rankedCandidates[i].candidate.Address < rankedCandidates[j].candidate.Address
	})

	result := Result{Candidates: make([]Candidate, len(rankedCandidates))}
	for i, item := range rankedCandidates {
		result.Candidates[i] = item.candidate
	}
	if len(rankedCandidates) == 0 {
		return result
	}
	if len(rankedCandidates) == 1 || rankedCandidates[0].score != rankedCandidates[1].score || rankedCandidates[0].defaulted != rankedCandidates[1].defaulted {
		winner := rankedCandidates[0].candidate
		result.Winner = &winner
		return result
	}
	result.Ambiguous = true
	return result
}

func defaultRouteID() string {
	conn, err := net.Dial("udp4", "1.1.1.1:80")
	if err != nil {
		return ""
	}
	defer conn.Close()

	local, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok {
		return ""
	}
	interfaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ip, _, err := net.ParseCIDR(addr.String())
			if err == nil && ip.Equal(local.IP) {
				return iface.Name
			}
		}
	}
	return ""
}

func isPrivateIPv4(ip net.IP) bool {
	ip = ip.To4()
	return ip != nil && ip.IsPrivate()
}

func interfaceKind(id string) string {
	if isVirtual(id) {
		return "virtual"
	}
	return "physical"
}

func interfaceScore(id string) int {
	if isVirtual(id) {
		return 0
	}
	if strings.HasPrefix(id, "en") || strings.HasPrefix(id, "eth") || strings.HasPrefix(id, "wl") {
		return 2
	}
	return 1
}

func isVirtual(id string) bool {
	return strings.HasPrefix(id, "tun") || strings.HasPrefix(id, "tap") ||
		strings.HasPrefix(id, "docker") || strings.HasPrefix(id, "br-") ||
		strings.HasPrefix(id, "utun")
}

// CandidateByID resolves an M3 selection without re-reading OS interfaces.
func CandidateByID(candidates []Candidate, id string) (Candidate, bool) {
	for _, candidate := range candidates {
		if candidate.ID == id {
			return candidate, true
		}
	}
	return Candidate{}, false
}
