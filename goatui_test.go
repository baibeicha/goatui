package goatui_test

import (
	"testing"
	"time"

	"github.com/baibeicha/goatui"
)

func TestGoatUIExports(t *testing.T) {
	// Widgets
	tabs := goatui.NewTabs(
		goatui.TabItem{ID: "1", Title: "Tab 1", Hotkey: '1'},
		goatui.TabItem{ID: "2", Title: "Tab 2", Hotkey: '2'},
	)
	if tabs == nil {
		t.Fatal("expected tabs != nil")
	}

	tree := goatui.NewTreeView(
		&goatui.TreeNode{ID: "root", Label: "Root"},
	)
	if tree == nil {
		t.Fatal("expected tree != nil")
	}

	// Modal
	modal := goatui.NewInputModal("Title", "Prompt:", "Default", nil, nil)
	if modal == nil {
		t.Fatal("expected modal != nil")
	}

	// Replace Cmd
	replaceCmd := goatui.NavigateReplace("/test")
	msg := replaceCmd()
	repMsg, ok := msg.(goatui.NavigateReplaceMsg)
	if !ok || repMsg.URL != "/test" {
		t.Fatalf("unexpected replace msg: %v", msg)
	}

	// Animation
	c1 := goatui.NewAnimationController(100*time.Millisecond, goatui.Linear)
	c2 := goatui.NewAnimationController(100*time.Millisecond, goatui.EaseInOutQuad)
	seq := goatui.NewSequenceController(c1, c2)
	if seq == nil {
		t.Fatal("expected seq != nil")
	}
	par := goatui.NewParallelController(c1, c2)
	if par == nil {
		t.Fatal("expected par != nil")
	}

	ps := goatui.NewParticleSystem(goatui.NewRect(0, 0, 80, 24), 9.8)
	if ps == nil {
		t.Fatal("expected ps != nil")
	}
	ps.EmitConfetti(10, 40, 12)
	ps.EmitSparks(10, 40, 12)
	ps.Update(0.016)

	// Sixel
	enc := goatui.NewSixelEncoder()
	if enc == nil {
		t.Fatal("expected sixel encoder != nil")
	}
}
