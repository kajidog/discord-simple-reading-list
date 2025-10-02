package commands

import (
	"testing"

	"github.com/example/discord-bookmark-manager/internal/store"
)

func TestResolveColorKeepsExistingWhenNotProvided(t *testing.T) {
	existing := store.EmojiPreference{Color: 0xabcdef, HasColor: true}
	color, hasColor, err := resolveColor("", existing, true)
	if err != nil {
		t.Fatalf("resolveColor returned error: %v", err)
	}
	if !hasColor {
		t.Fatalf("expected hasColor to be true")
	}
	if color != 0xabcdef {
		t.Fatalf("expected color 0xabcdef, got %#x", color)
	}
}

func TestResolveColorParsesNewValue(t *testing.T) {
	existing := store.EmojiPreference{Color: 0x112233, HasColor: true}
	color, hasColor, err := resolveColor("#ffcc00", existing, true)
	if err != nil {
		t.Fatalf("resolveColor returned error: %v", err)
	}
	if !hasColor {
		t.Fatalf("expected hasColor to be true")
	}
	if color != 0xffcc00 {
		t.Fatalf("expected color 0xffcc00, got %#x", color)
	}
}

func TestResolveColorWithoutExisting(t *testing.T) {
	existing := store.EmojiPreference{}
	color, hasColor, err := resolveColor("", existing, false)
	if err != nil {
		t.Fatalf("resolveColor returned error: %v", err)
	}
	if hasColor {
		t.Fatalf("expected hasColor to be false")
	}
	if color != 0 {
		t.Fatalf("expected color 0, got %#x", color)
	}
}

func TestResolveColorInvalidInput(t *testing.T) {
	existing := store.EmojiPreference{Color: 0x123456, HasColor: true}
	if _, _, err := resolveColor("not-a-color", existing, true); err == nil {
		t.Fatalf("expected an error for invalid color input")
	}
}
