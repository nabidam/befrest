package presence

import (
	"encoding/json"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/nabidam/befrest/internal/proto"
)

func TestRegistryJoinValidatesAndDeduplicatesNames(t *testing.T) {
	registry := NewRegistry(nil)

	first, err := registry.Join("  Pixel 8  ", "mobile", false)
	if err != nil {
		t.Fatalf("Join() error = %v", err)
	}
	if !regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`).MatchString(first.ID) {
		t.Errorf("Join() ID = %q, want UUIDv4", first.ID)
	}
	if first.Name != "Pixel 8" || first.RawName != "Pixel 8" {
		t.Errorf("Join() device = %#v, want trimmed Pixel 8", first)
	}

	second, err := registry.Join("Pixel 8", "mobile", false)
	if err != nil {
		t.Fatalf("second Join() error = %v", err)
	}
	if second.Name != "Pixel 8 (2)" || second.RawName != "Pixel 8" {
		t.Errorf("second Join() device = %#v, want deduplicated name", second)
	}

	long, err := registry.Join(strings.Repeat("界", 40), "desktop", false)
	if err != nil {
		t.Fatalf("long Join() error = %v", err)
	}
	if utf8.RuneCountInString(long.Name) != 32 || utf8.RuneCountInString(long.RawName) != 32 {
		t.Errorf("long name was not clamped: %#v", long)
	}

	if _, err := registry.Join(" \t ", "desktop", false); err != ErrInvalidName {
		t.Errorf("Join(empty) error = %v, want %v", err, ErrInvalidName)
	}
}

func TestRegistryRenameRededuplicatesAndLeaveReleasesName(t *testing.T) {
	registry := NewRegistry(nil)
	first, _ := registry.Join("Laptop", "desktop", false)
	second, _ := registry.Join("Phone", "mobile", false)

	renamed, err := registry.Rename(second.ID, "Laptop")
	if err != nil {
		t.Fatalf("Rename() error = %v", err)
	}
	if renamed.Name != "Laptop (2)" || renamed.RawName != "Laptop" {
		t.Errorf("Rename() = %#v, want Laptop (2)", renamed)
	}
	if !registry.Leave(first.ID) {
		t.Fatal("Leave() = false, want true")
	}

	third, err := registry.Join("Laptop", "desktop", false)
	if err != nil {
		t.Fatalf("Join() after Leave() error = %v", err)
	}
	if third.Name != "Laptop" {
		t.Errorf("Join() name = %q, want released name Laptop", third.Name)
	}
	if _, err := registry.Rename("missing", "Anything"); err != ErrDeviceNotFound {
		t.Errorf("Rename(missing) error = %v, want %v", err, ErrDeviceNotFound)
	}
}

func TestRegistryFansOutSnapshotsForEveryChange(t *testing.T) {
	var snapshots [][]Device
	registry := NewRegistry(func(snapshot []Device) {
		snapshots = append(snapshots, snapshot)
	})

	first, _ := registry.Join("Laptop", "desktop", false)
	second, _ := registry.Join("Phone", "mobile", false)
	if _, err := registry.Rename(second.ID, "Tablet"); err != nil {
		t.Fatalf("Rename() error = %v", err)
	}
	registry.Leave(first.ID)

	if len(snapshots) != 4 {
		t.Fatalf("fanout count = %d, want 4", len(snapshots))
	}
	last := snapshots[len(snapshots)-1]
	if len(last) != 1 || last[0].ID != second.ID || last[0].Name != "Tablet" {
		t.Errorf("leave snapshot = %#v, want only renamed second device", last)
	}
	last[0].Name = "mutated"
	if registry.Snapshot()[0].Name != "Tablet" {
		t.Error("notifier snapshot aliases registry state")
	}
}

func TestPresenceDoesNotImportNetworkPackages(t *testing.T) {
	command := exec.Command("go", "list", "-json", ".")
	output, err := command.Output()
	if err != nil {
		t.Fatalf("go list presence imports: %v", err)
	}
	var packageInfo struct {
		Imports []string
	}
	if err := json.Unmarshal(output, &packageInfo); err != nil {
		t.Fatalf("decode go list output: %v", err)
	}
	for _, imported := range packageInfo.Imports {
		if imported == "net/http" || strings.Contains(imported, "websocket") {
			t.Errorf("presence imports forbidden network package %q", imported)
		}
	}
}

func TestPresenceFramesMarshalWithWireNames(t *testing.T) {
	frame, err := json.Marshal(proto.NeedName{Type: proto.MsgNeedName, Suggested: "Pixel 8"})
	if err != nil {
		t.Fatalf("marshal need-name: %v", err)
	}
	if string(frame) != `{"type":"need-name","suggested":"Pixel 8"}` {
		t.Errorf("need-name frame = %s", frame)
	}

	frame, err = json.Marshal(proto.Error{Type: proto.MsgError, Code: "target-gone", Message: "Target disconnected"})
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	if string(frame) != `{"type":"error","code":"target-gone","message":"Target disconnected"}` {
		t.Errorf("error frame = %s", frame)
	}
}
