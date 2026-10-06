package driver

import (
	"os"
	"testing"

	"github.com/baibeicha/goatui/pkg/driver/input"
)

func TestDriverConfigOptions(t *testing.T) {
	cfg := DefaultDriverConfig()
	if cfg.KittyFlags != input.KittyModeDisambiguateEscapeCodes {
		t.Errorf("expected KittyFlags %d, got %d", input.KittyModeDisambiguateEscapeCodes, cfg.KittyFlags)
	}
	if cfg.DisableKitty {
		t.Errorf("expected DisableKitty to be false by default")
	}

	opt1 := WithKittyFlags(input.KittyModeDisambiguateEscapeCodes | input.KittyModeReportEventTypes)
	opt1(&cfg)
	if cfg.KittyFlags != 3 || cfg.DisableKitty {
		t.Errorf("WithKittyFlags failed: %+v", cfg)
	}

	opt2 := WithoutKittyKeyboard()
	opt2(&cfg)
	if !cfg.DisableKitty {
		t.Errorf("WithoutKittyKeyboard failed: %+v", cfg)
	}
}

func TestNewDriverOptions(t *testing.T) {
	// Create driver with custom Kitty flags
	d, err := NewDriver(WithKittyFlags(1))
	if err != nil {
		t.Fatalf("NewDriver failed: %v", err)
	}
	if d == nil {
		t.Fatal("expected non-nil driver")
	}

	// Create driver with Kitty disabled
	d2, err := NewDriver(WithoutKittyKeyboard())
	if err != nil {
		t.Fatalf("NewDriver without kitty failed: %v", err)
	}
	if d2 == nil {
		t.Fatal("expected non-nil driver")
	}
}

func TestTeardownHook(t *testing.T) {
	called := false
	RegisterTeardown(func() {
		called = true
	})

	TearDown()
	if !called {
		t.Fatal("expected teardown hook to be executed")
	}
}

func TestConinOpening(t *testing.T) {
	f, err := os.OpenFile("CONIN$", os.O_RDWR, 0)
	if err != nil {
		t.Logf("CONIN$ open: %v", err)
	} else {
		defer f.Close()
		t.Logf("CONIN$ opened successfully: fd=%v", f.Fd())
	}

	fout, err := os.OpenFile("CONOUT$", os.O_RDWR, 0)
	if err != nil {
		t.Logf("CONOUT$ open: %v", err)
	} else {
		defer fout.Close()
		t.Logf("CONOUT$ opened successfully: fd=%v", fout.Fd())
	}
}
