package widgets

import (
	"strconv"
	"testing"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
	"github.com/baibeicha/goatui/pkg/driver/input"
	"github.com/baibeicha/goatui/pkg/tea"
)

func TestVirtualListOneMillionItems(t *testing.T) {
	total := 1_000_000
	renderedItems := 0

	renderer := func(buf *buffer.Buffer, area buffer.Rect, index int, selected bool) {
		renderedItems++
		buf.SetString(area.X, area.Y, "Row "+strconv.Itoa(index), cell.DefaultColor(), cell.DefaultColor(), cell.AttrNone)
	}

	vl := NewVirtualList(total, renderer)
	buf := buffer.NewBuffer(80, 20)

	// Initial render: height 20 should only render 20 items
	vl.Draw(buf, buf.Area())
	if renderedItems != 20 {
		t.Errorf("Expected exactly 20 items rendered, got %d", renderedItems)
	}

	// Scroll to item 500,000
	renderedItems = 0
	vl.Select(500_000)
	vl.Draw(buf, buf.Area())
	if vl.Selected() != 500_000 {
		t.Errorf("Expected selected = 500,000, got %d", vl.Selected())
	}
	if renderedItems != 20 {
		t.Errorf("Expected 20 items rendered at scroll position, got %d", renderedItems)
	}
}

func TestVirtualTable(t *testing.T) {
	cols := []TableColumn{
		{Title: "ID", Width: 5, Align: AlignRight},
		{Title: "Name", Flex: 1, MinWidth: 10, Align: AlignLeft},
		{Title: "Status", Width: 10, Align: AlignCenter},
	}

	vt := NewTable(cols).
		SetRows([][]string{
			{"#1", "daemon-proc", "ACTIVE"},
			{"#2", "worker-proc", "IDLE"},
		}).
		SetPadding(1)

	buf := buffer.NewBuffer(50, 10)
	vt.Draw(buf, buf.Area())

	// Row 0 is header: with padding 1 and AlignRight on ID(width 5), ID should be right-aligned
	// Header separator is at Row 1
	cSep := buf.Cell(0, 1)
	if cSep == nil || cSep.Rune != '─' {
		t.Errorf("Expected header separator '─' at (0, 1), got %+v", cSep)
	}

	// Verify RowAt
	r0 := vt.RowAt(buf.Area(), 2)
	if r0 != 0 {
		t.Errorf("Expected RowAt(y=2) to be 0, got %d", r0)
	}
	r1 := vt.RowAt(buf.Area(), 3)
	if r1 != 1 {
		t.Errorf("Expected RowAt(y=3) to be 1, got %d", r1)
	}
	rOut := vt.RowAt(buf.Area(), 0) // Header row
	if rOut != -1 {
		t.Errorf("Expected RowAt(y=0) to be -1 (header), got %d", rOut)
	}
}

func TestTableFlexLayout(t *testing.T) {
	cols := []TableColumn{
		{Title: "ID", Width: 10},
		{Title: "Flex1", Flex: 1, MinWidth: 5},
		{Title: "Flex2", Flex: 2, MinWidth: 5},
	}
	vt := NewTable(cols).SetBorder(TableBorderClean)

	// Total width 100
	// Clean border has 2 column separators: net width = 100 - 2 = 98
	// Fixed: 10
	// Remaining: 98 - 10 = 88
	// Flex sum = 3: Flex1 gets (88*1)/3 = 29 (+1 remainder = 30), Flex2 gets (88*2)/3 = 58
	colX, colW := vt.ComputeColumnLayout(100)
	if len(colW) != 3 {
		t.Fatalf("Expected 3 columns, got %d", len(colW))
	}
	if colW[0] != 10 {
		t.Errorf("Expected Col 0 width 10, got %d", colW[0])
	}
	if colW[1]+colW[2] != 88 {
		t.Errorf("Expected sum of flex cols 88, got %d", colW[1]+colW[2])
	}
	if colX[0] != 0 {
		t.Errorf("Expected Col 0 X=0, got %d", colX[0])
	}
}

func TestTableBorderRounded(t *testing.T) {
	cols := []TableColumn{
		{Title: "A", Width: 5},
		{Title: "B", Width: 5},
	}
	vt := NewTable(cols).SetBorder(TableBorderRounded)
	buf := buffer.NewBuffer(20, 6)
	vt.Draw(buf, buf.Area())

	// Check top-left corner
	if c := buf.Cell(0, 0); c == nil || c.Rune != '╭' {
		t.Errorf("Expected '╭' at (0, 0), got %+v", c)
	}
	// Check top cross
	// Col 0: width 5 -> X=1 to 5. Col sep at X=6
	if c := buf.Cell(6, 0); c == nil || c.Rune != '┬' {
		t.Errorf("Expected '┬' at (6, 0), got %+v", c)
	}
	// Check bottom-right corner
	if c := buf.Cell(buf.Area().Right()-1, buf.Area().Bottom()-1); c == nil || c.Rune != '╯' {
		t.Errorf("Expected '╯' at bottom right, got %+v", c)
	}
}

func TestBrailleCanvasSubpixels(t *testing.T) {
	bc := NewBrailleCanvas(10, 5)
	if bc.SubWidth() != 20 || bc.SubHeight() != 20 {
		t.Errorf("Expected 20x20 subpixels, got %dx%d", bc.SubWidth(), bc.SubHeight())
	}

	// Set pixel at (0, 0) -> dot 1 (0x01)
	bc.SetPixel(0, 0)
	buf := buffer.NewBuffer(10, 5)
	bc.Draw(buf, buf.Area())

	c := buf.Cell(0, 0)
	// Base braille 0x2800 + 0x01 = 0x2801 ('⠁')
	if c == nil || c.Rune != 0x2801 {
		t.Errorf("Expected Braille rune '⠁' (0x2801), got %q (0x%X)", c.Rune, c.Rune)
	}
}

func TestSparklineAndGauge(t *testing.T) {
	buf := buffer.NewBuffer(20, 3)

	sp := NewSparkline([]float64{10, 20, 50, 80, 100})
	sp.Draw(buf, buffer.NewRect(0, 0, 10, 1))

	g := NewGauge()
	g.SetPercent(0.5)
	g.Draw(buf, buffer.NewRect(0, 1, 10, 1))

	// First cell of gauge at 50% should be filled '█'
	c := buf.Cell(0, 1)
	if c == nil || c.Rune != '█' {
		t.Errorf("Expected gauge filled '█', got %+v", c)
	}
}

func TestTextInputEditing(t *testing.T) {
	ti := NewTextInput()
	ti.HandleKey(input.Key{Type: input.KeyRune, Rune: 'H'})
	ti.HandleKey(input.Key{Type: input.KeyRune, Rune: 'i'})
	if ti.Value() != "Hi" {
		t.Errorf("Expected 'Hi', got %q", ti.Value())
	}

	ti.HandleKey(input.Key{Type: input.KeyBackspace})
	if ti.Value() != "H" {
		t.Errorf("Expected 'H' after backspace, got %q", ti.Value())
	}

	// Word editing test
	ti.SetValue("hello world test")
	ti.HandleKey(input.Key{Type: input.KeyLeft, Mod: cell.AttrBold})
	// Cursor should jump to start of 'test'
	if ti.Cursor() != 12 {
		t.Errorf("Expected cursor at 12 after word jump, got %d", ti.Cursor())
	}
	ti.HandleKey(input.Key{Type: input.KeyBackspace, Mod: cell.AttrBold})
	// 'world ' should be deleted
	if ti.Value() != "hello test" {
		t.Errorf("Expected 'hello test' after word delete, got %q", ti.Value())
	}

	// Password mode test
	passInput := NewTextInput().SetPasswordMode(true)
	passInput.SetValue("secret123")
	if passInput.EchoMode() != EchoPassword {
		t.Errorf("Expected EchoPassword, got %v", passInput.EchoMode())
	}
	if passInput.Value() != "secret123" {
		t.Errorf("Expected plaintext value 'secret123', got %q", passInput.Value())
	}
	buf := buffer.NewBuffer(30, 1)
	passInput.Draw(buf, buffer.NewRect(0, 0, 20, 1))
	// Verify mask rune '•' was rendered
	foundMask := false
	for _, c := range buf.Cells() {
		if c.Rune == '•' {
			foundMask = true
			break
		}
	}
	if !foundMask {
		t.Errorf("Expected masked runes '•' to be drawn in password mode")
	}
}

func TestVirtualTableNavigation(t *testing.T) {
	cols := []TableColumn{
		{Title: "ID", Width: 8},
		{Title: "Name", Flex: 1},
	}
	vt := NewTable(cols).SetTotalRows(100)
	if vt.Selected() != 0 {
		t.Fatalf("Expected initial selected 0, got %d", vt.Selected())
	}

	vt.PageDown(20)
	if vt.Selected() != 20 {
		t.Errorf("Expected selected 20 after PageDown, got %d", vt.Selected())
	}

	vt.PageUp(10)
	if vt.Selected() != 10 {
		t.Errorf("Expected selected 10 after PageUp, got %d", vt.Selected())
	}

	vt.ScrollToBottom()
	if vt.Selected() != 99 {
		t.Errorf("Expected selected 99 after ScrollToBottom, got %d", vt.Selected())
	}

	vt.ScrollToTop()
	if vt.Selected() != 0 {
		t.Errorf("Expected selected 0 after ScrollToTop, got %d", vt.Selected())
	}

	// Test HandleKey
	vt.HandleKey(input.Key{Type: input.KeyRune, Rune: 'j'}, 10)
	if vt.Selected() != 1 {
		t.Errorf("Expected selected 1 after 'j', got %d", vt.Selected())
	}

	vt.HandleKey(input.Key{Type: input.KeyRune, Rune: 'k'}, 10)
	if vt.Selected() != 0 {
		t.Errorf("Expected selected 0 after 'k', got %d", vt.Selected())
	}

	vt.HandleKey(input.Key{Type: input.KeyRune, Rune: 'G'}, 10)
	if vt.Selected() != 99 {
		t.Errorf("Expected selected 99 after 'G', got %d", vt.Selected())
	}
}

func TestVirtualListNavigation(t *testing.T) {
	vl := NewVirtualList(100, nil)
	if vl.Selected() != 0 {
		t.Fatalf("Expected initial selected 0, got %d", vl.Selected())
	}

	vl.PageDown(20)
	if vl.Selected() != 20 {
		t.Errorf("Expected selected 20 after PageDown, got %d", vl.Selected())
	}

	vl.PageUp(10)
	if vl.Selected() != 10 {
		t.Errorf("Expected selected 10 after PageUp, got %d", vl.Selected())
	}

	vl.ScrollToBottom()
	if vl.Selected() != 99 {
		t.Errorf("Expected selected 99 after ScrollToBottom, got %d", vl.Selected())
	}

	vl.ScrollToTop()
	if vl.Selected() != 0 {
		t.Errorf("Expected selected 0 after ScrollToTop, got %d", vl.Selected())
	}

	// Test HandleKey
	vl.HandleKey(input.Key{Type: input.KeyRune, Rune: 'j'}, 10)
	if vl.Selected() != 1 {
		t.Errorf("Expected selected 1 after 'j', got %d", vl.Selected())
	}

	vl.HandleKey(input.Key{Type: input.KeyRune, Rune: 'k'}, 10)
	if vl.Selected() != 0 {
		t.Errorf("Expected selected 0 after 'k', got %d", vl.Selected())
	}

	// Test HandleMouse click and wheel
	area := buffer.NewRect(0, 0, 50, 10)
	vl.HandleMouse(tea.MouseMsg{Mouse: input.Mouse{Action: input.MousePress, Button: input.MouseLeft, X: 5, Y: 4}}, area)
	if vl.Selected() != 4 {
		t.Errorf("Expected selected 4 after mouse click, got %d", vl.Selected())
	}

	vl.HandleMouse(tea.MouseMsg{Mouse: input.Mouse{Button: input.MouseWheelDown}}, area)
	if vl.Selected() != 5 {
		t.Errorf("Expected selected 5 after mouse wheel down, got %d", vl.Selected())
	}
}

func TestTextInputUnicodeAndPaste(t *testing.T) {
	ti := NewTextInput()

	// Cyrillic word movement and deletion
	ti.SetValue("Привет Мир Тест")
	ti.HandleKey(input.Key{Type: input.KeyLeft, Mod: input.ModCtrl})
	// Cursor should jump to start of "Тест" (index 11)
	if ti.Cursor() != 11 {
		t.Errorf("Expected cursor at 11 after Cyrillic word jump, got %d", ti.Cursor())
	}
	ti.HandleKey(input.Key{Type: input.KeyBackspace, Mod: input.ModCtrl})
	if ti.Value() != "Привет Тест" {
		t.Errorf("Expected 'Привет Тест' after Cyrillic word delete, got %q", ti.Value())
	}

	// Test InsertString (paste simulation with multiline)
	ti.SetValue("Hello ")
	ti.InsertString("World\r\n2026")
	if ti.Value() != "Hello World2026" {
		t.Errorf("Expected newlines stripped in InsertString, got %q", ti.Value())
	}

	// Test CJK wide-character cursor rendering
	ti.SetValue("你好世界")
	ti.SetCursor(0)
	buf := buffer.NewBuffer(20, 1)
	ti.Draw(buf, buffer.NewRect(0, 0, 20, 1))
	// Cell at 0 must have Rune: '你', Width: 2, and cell at 1 must have Width: 0 continuation
	c0 := buf.Cell(2, 0) // offset by prompt "> " (width 2)
	c1 := buf.Cell(3, 0)
	if c0 == nil || c0.Rune != '你' || c0.Width != 2 || !c0.Modifier.Has(cell.AttrReverse) {
		t.Errorf("Expected wide cursor cell at (2,0), got %+v", c0)
	}
	if c1 == nil || c1.Width != 0 {
		t.Errorf("Expected wide continuation cell at (3,0), got %+v", c1)
	}
}

func TestHotkeyModifierIsolation(t *testing.T) {
	// Verify that Ctrl+K does NOT trigger vim 'k' in VirtualTable or VirtualList
	vt := NewTable([]TableColumn{{Title: "A"}}).SetTotalRows(10)
	vt.Select(5)
	ctrlK := input.Key{Type: input.KeyRune, Rune: 'k', Mod: input.ModCtrl}
	if vt.HandleKey(ctrlK, 10) {
		t.Errorf("VirtualTable should NOT handle Ctrl+K as vim up")
	}
	if vt.Selected() != 5 {
		t.Errorf("VirtualTable selection should remain 5, got %d", vt.Selected())
	}

	vl := NewVirtualList(10, nil)
	vl.Select(5)
	if vl.HandleKey(ctrlK, 10) {
		t.Errorf("VirtualList should NOT handle Ctrl+K as vim up")
	}
	if vl.Selected() != 5 {
		t.Errorf("VirtualList selection should remain 5, got %d", vl.Selected())
	}

	// Verify Tabs does NOT handle Ctrl+hotkey
	tabs := NewTabs(TabItem{Title: "Home", Hotkey: 'k'})
	if tabs.HandleKey(ctrlK) {
		t.Errorf("Tabs should NOT trigger hotkey when Ctrl is active")
	}

	// Verify CheckboxGroup does NOT handle Ctrl+hotkey
	cg := NewCheckboxGroup(CheckboxItem{Label: "Item", Hotkey: 'k'})
	if cg.HandleKey(ctrlK) {
		t.Errorf("CheckboxGroup should NOT trigger hotkey when Ctrl is active")
	}
}

func BenchmarkVirtualListScroll(b *testing.B) {
	total := 1_000_000
	renderer := func(buf *buffer.Buffer, area buffer.Rect, index int, selected bool) {
		buf.SetString(area.X, area.Y, "Row Item", cell.DefaultColor(), cell.DefaultColor(), cell.AttrNone)
	}
	vl := NewVirtualList(total, renderer)
	buf := buffer.NewBuffer(80, 24)
	area := buf.Area()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vl.ScrollDown(1)
		vl.Draw(buf, area)
	}
}

func BenchmarkVirtualTableScroll(b *testing.B) {
	cols := []TableColumn{
		{Title: "ID", Width: 8, Align: AlignRight},
		{Title: "Name", Flex: 1, MinWidth: 15, Align: AlignLeft},
		{Title: "Memory", Width: 12, Align: AlignRight},
		{Title: "Status", Width: 10, Align: AlignCenter},
	}
	cachedRow := []string{"#000042", "system-daemon-worker", "512 MB", "RUNNING"}
	vt := NewTable(cols).
		SetTotalRows(100_000).
		SetRowProvider(func(r int) []string {
			return cachedRow
		}).
		SetZebra(true).
		SetBorder(TableBorderClean)

	buf := buffer.NewBuffer(100, 30)
	area := buf.Area()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vt.ScrollDown(1)
		vt.Draw(buf, area)
	}
}

func TestTabsWidget(t *testing.T) {
	tabs := NewTabs(
		TabItem{ID: "t1", Title: "First", Hotkey: '1'},
		TabItem{ID: "t2", Title: "Second", Hotkey: '2', Badge: "New"},
		TabItem{ID: "t3", Title: "Third", Hotkey: '3'},
	)

	if tabs.Active() != 0 || tabs.ActiveItem().ID != "t1" {
		t.Errorf("Expected active tab t1, got %v", tabs.ActiveItem())
	}

	// Select next via KeyRight
	tabs.HandleKey(input.Key{Type: input.KeyRight})
	if tabs.Active() != 1 || tabs.ActiveItem().ID != "t2" {
		t.Errorf("Expected active tab t2, got %v", tabs.ActiveItem())
	}

	// Hotkey '3'
	tabs.HandleKey(input.Key{Type: input.KeyRune, Rune: '3'})
	if tabs.Active() != 2 || tabs.ActiveItem().ID != "t3" {
		t.Errorf("Expected active tab t3, got %v", tabs.ActiveItem())
	}

	// Draw into buffer
	buf := buffer.NewBuffer(40, 2)
	tabs.Draw(buf, buffer.NewRect(0, 0, 40, 1))

	// Mouse click on tab 1
	clicked := tabs.HandleMouse(tea.MouseMsg{
		Mouse: input.Mouse{
			X:      2,
			Y:      0,
			Button: input.MouseLeft,
			Action: input.MousePress,
		},
	})
	if !clicked || tabs.Active() != 0 {
		t.Errorf("Expected mouse click to select tab 0, clicked=%v, active=%d", clicked, tabs.Active())
	}
}

func TestTreeViewWidget(t *testing.T) {
	root := &TreeNode{
		ID:       "root",
		Label:    "Root",
		Expanded: true,
		Children: []*TreeNode{
			{ID: "c1", Label: "Child 1"},
			{
				ID:       "c2",
				Label:    "Child 2",
				Expanded: false,
				Children: []*TreeNode{
					{ID: "gc1", Label: "Grandchild 1"},
				},
			},
		},
	}

	tv := NewTreeView(root)
	if tv.Selected() == nil || tv.Selected().ID != "root" {
		t.Fatalf("Expected selected root, got %v", tv.Selected())
	}

	// Down -> Child 1
	tv.HandleKey(input.Key{Type: input.KeyDown})
	if tv.Selected().ID != "c1" {
		t.Errorf("Expected selected c1, got %s", tv.Selected().ID)
	}

	// Down -> Child 2
	tv.HandleKey(input.Key{Type: input.KeyDown})
	if tv.Selected().ID != "c2" {
		t.Errorf("Expected selected c2, got %s", tv.Selected().ID)
	}

	// Expand Child 2
	tv.HandleKey(input.Key{Type: input.KeyRight})
	if !tv.Selected().Expanded {
		t.Errorf("Expected c2 to be expanded")
	}

	// Down -> Grandchild 1
	tv.HandleKey(input.Key{Type: input.KeyDown})
	if tv.Selected().ID != "gc1" {
		t.Errorf("Expected selected gc1, got %s", tv.Selected().ID)
	}

	// Draw
	buf := buffer.NewBuffer(30, 10)
	tv.Draw(buf, buffer.NewRect(0, 0, 30, 10))
}
