package widgets

import (
	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
	"github.com/baibeicha/goatui/pkg/driver/input"
	"github.com/baibeicha/goatui/pkg/style"
	"github.com/baibeicha/goatui/pkg/tea"
)

// Alignment defines the horizontal text alignment within a table cell or header.
type Alignment = buffer.Alignment

const (
	AlignLeft   = buffer.AlignLeft
	AlignCenter = buffer.AlignCenter
	AlignRight  = buffer.AlignRight
)

// TableColumn specifies a table column definition with rich customization options.
type TableColumn struct {
	Title       string    // Column header title
	Width       int       // Fixed column width (excluding separators). 0 if Flex > 0.
	Flex        int       // Flex weight for proportional expansion to fill available table width.
	MinWidth    int       // Minimum width when Flex > 0 (defaults to 4 if 0).
	Align       Alignment // Alignment for cell data (AlignLeft, AlignCenter, AlignRight).
	HeaderAlign Alignment // Alignment for column header. Defaults to Align if not set.
}

// TableBorder defines the box-drawing characters used for rendering table frames and separators.
type TableBorder struct {
	HasOuter     bool
	HasColSep    bool
	HasHeaderSep bool

	Top         rune
	Bottom      rune
	Left        rune
	Right       rune
	TopLeft     rune
	TopRight    rune
	BottomLeft  rune
	BottomRight rune
	TopCross    rune
	BottomCross rune

	ColSep      rune
	HeaderSep   rune
	HeaderLeft  rune
	HeaderRight rune
	HeaderCross rune
}

var (
	// TableBorderRounded uses curved corners with full column dividers and header separators: ╭─┬─╮ │ │ ├─┼─┤ ╰─┴─╯
	TableBorderRounded = TableBorder{
		HasOuter:     true,
		HasColSep:    true,
		HasHeaderSep: true,
		Top:          '─',
		Bottom:       '─',
		Left:         '│',
		Right:        '│',
		TopLeft:      '╭',
		TopRight:     '╮',
		BottomLeft:   '╰',
		BottomRight:  '╯',
		TopCross:     '┬',
		BottomCross:  '┴',
		ColSep:       '│',
		HeaderSep:    '─',
		HeaderLeft:   '├',
		HeaderRight:  '┤',
		HeaderCross:  '┼',
	}

	// TableBorderNormal uses standard sharp box corners: ┌─┬─┐ │ │ ├─┼─┤ └─┴─┘
	TableBorderNormal = TableBorder{
		HasOuter:     true,
		HasColSep:    true,
		HasHeaderSep: true,
		Top:          '─',
		Bottom:       '─',
		Left:         '│',
		Right:        '│',
		TopLeft:      '┌',
		TopRight:     '┐',
		BottomLeft:   '└',
		BottomRight:  '┘',
		TopCross:     '┬',
		BottomCross:  '┴',
		ColSep:       '│',
		HeaderSep:    '─',
		HeaderLeft:   '├',
		HeaderRight:  '┤',
		HeaderCross:  '┼',
	}

	// TableBorderClean has no outer border, but includes vertical column dividers and a horizontal header separator: Title │ Title ──┼──
	TableBorderClean = TableBorder{
		HasOuter:     false,
		HasColSep:    true,
		HasHeaderSep: true,
		ColSep:       '│',
		HeaderSep:    '─',
		HeaderCross:  '┼',
	}

	// TableBorderNone displays columns separated only by spaces without lines.
	TableBorderNone = TableBorder{
		HasOuter:     false,
		HasColSep:    false,
		HasHeaderSep: false,
	}
)

// TableCellRenderer renders an individual column cell.
// Return true if custom rendering was performed, or false to use default text rendering.
type TableCellRenderer func(buf *buffer.Buffer, col TableColumn, cellRect buffer.Rect, rowIndex int, colIndex int, selected bool) bool

// TableRowRenderer is the legacy row renderer callback for backward compatibility.
type TableRowRenderer func(buf *buffer.Buffer, cols []TableColumn, area buffer.Rect, rowIndex int, selected bool)

// VirtualTable provides ultra-fast table rendering for massive datasets (100k+ rows) with sticky headers,
// full border customization, flex-based column sizing, alignments, and zebra striping.
type VirtualTable struct {
	columns           []TableColumn
	totalRows         int
	selectedRow       int
	offset            int
	border            TableBorder
	borderStyle       style.Style
	headerStyle       style.Style
	headerBorderStyle style.Style
	rowStyle          style.Style
	alternateRowStyle style.Style
	selectedStyle     style.Style
	selectionPrefix   string
	zebra             bool
	padding           int

	// Data sources
	staticRows     [][]string
	rowProvider    func(rowIndex int) []string
	cellRenderer   TableCellRenderer
	legacyRenderer TableRowRenderer

	// Geometry cache to ensure 0 allocs/op during Draw
	cachedWidth int
	cachedColX  []int
	cachedColW  []int
}

// NewVirtualTable creates a new virtualized table (backward-compatible constructor).
func NewVirtualTable(cols []TableColumn, totalRows int, renderer TableRowRenderer) *VirtualTable {
	vt := NewTable(cols)
	vt.totalRows = totalRows
	vt.legacyRenderer = renderer
	return vt
}

// NewTable creates a new customizable Table widget.
func NewTable(cols []TableColumn) *VirtualTable {
	return &VirtualTable{
		columns:           cols,
		border:            TableBorderClean,
		borderStyle:       style.NewStyle().Foreground(cell.ColorHex("#555577")),
		headerStyle:       style.NewStyle().Bold(true).Foreground(cell.ColorHex("#00D2FF")),
		headerBorderStyle: style.NewStyle().Foreground(cell.ColorHex("#555577")),
		rowStyle:          style.NewStyle().Foreground(cell.DefaultColor()),
		alternateRowStyle: style.NewStyle().Foreground(cell.DefaultColor()).Background(cell.Color256(234)),
		selectedStyle:     style.NewStyle().Bold(true).Foreground(cell.ColorHex("#00FFAA")).Background(cell.Color256(236)),
		selectionPrefix:   "> ",
		zebra:             false,
		padding:           1,
	}
}

// SetBorder configures the table border style.
func (vt *VirtualTable) SetBorder(b TableBorder) *VirtualTable {
	vt.border = b
	return vt
}

// SetBorderStyle configures the style for table borders and column separators.
func (vt *VirtualTable) SetBorderStyle(s style.Style) *VirtualTable {
	vt.borderStyle = s
	return vt
}

// SetHeaderStyle configures the text style for header column titles.
func (vt *VirtualTable) SetHeaderStyle(s style.Style) *VirtualTable {
	vt.headerStyle = s
	return vt
}

// SetHeaderBorderStyle configures the style for the separator line under the header.
func (vt *VirtualTable) SetHeaderBorderStyle(s style.Style) *VirtualTable {
	vt.headerBorderStyle = s
	return vt
}

// SetRowStyle configures the default style for regular data rows.
func (vt *VirtualTable) SetRowStyle(s style.Style) *VirtualTable {
	vt.rowStyle = s
	return vt
}

// SetAlternateRowStyle configures the style for alternating rows when zebra striping is enabled.
func (vt *VirtualTable) SetAlternateRowStyle(s style.Style) *VirtualTable {
	vt.alternateRowStyle = s
	return vt
}

// SetSelectedStyle configures the style for the active selected row.
func (vt *VirtualTable) SetSelectedStyle(s style.Style) *VirtualTable {
	vt.selectedStyle = s
	return vt
}

// SetZebra enables or disables alternating row background colors.
func (vt *VirtualTable) SetZebra(enable bool) *VirtualTable {
	vt.zebra = enable
	return vt
}

// SetSelectionPrefix sets an optional prefix string (e.g. "> ") drawn in the first column of the selected row.
func (vt *VirtualTable) SetSelectionPrefix(prefix string) *VirtualTable {
	vt.selectionPrefix = prefix
	return vt
}

// SetPadding sets horizontal padding inside each cell (default is 1 space).
func (vt *VirtualTable) SetPadding(p int) *VirtualTable {
	if p < 0 {
		p = 0
	}
	vt.padding = p
	return vt
}

// SetTotalRows updates the total number of rows.
func (vt *VirtualTable) SetTotalRows(n int) *VirtualTable {
	if n < 0 {
		n = 0
	}
	vt.totalRows = n
	if vt.selectedRow >= n {
		vt.selectedRow = max(0, n-1)
	}
	return vt
}

// SetRows sets static data rows.
func (vt *VirtualTable) SetRows(rows [][]string) *VirtualTable {
	vt.staticRows = rows
	vt.totalRows = len(rows)
	if vt.selectedRow >= vt.totalRows {
		vt.selectedRow = max(0, vt.totalRows-1)
	}
	return vt
}

// SetRowProvider configures an on-demand row data provider for massive virtual datasets.
func (vt *VirtualTable) SetRowProvider(provider func(rowIndex int) []string) *VirtualTable {
	vt.rowProvider = provider
	return vt
}

// SetCellRenderer configures a custom cell renderer callback.
func (vt *VirtualTable) SetCellRenderer(renderer TableCellRenderer) *VirtualTable {
	vt.cellRenderer = renderer
	return vt
}

// Select sets selected row index with bounds clamping.
func (vt *VirtualTable) Select(row int) {
	if vt.totalRows == 0 {
		vt.selectedRow = 0
		vt.offset = 0
		return
	}
	if row < 0 {
		row = 0
	}
	if row >= vt.totalRows {
		row = vt.totalRows - 1
	}
	vt.selectedRow = row
}

// ScrollDown moves selected row down by n rows.
func (vt *VirtualTable) ScrollDown(n int) {
	vt.Select(vt.selectedRow + n)
}

// ScrollUp moves selected row up by n rows.
func (vt *VirtualTable) ScrollUp(n int) {
	vt.Select(vt.selectedRow - n)
}

// PageDown moves selection down by a page (default 10 rows if pageSize <= 0).
func (vt *VirtualTable) PageDown(pageSize int) {
	if pageSize <= 0 {
		pageSize = 10
	}
	vt.ScrollDown(pageSize)
}

// PageUp moves selection up by a page (default 10 rows if pageSize <= 0).
func (vt *VirtualTable) PageUp(pageSize int) {
	if pageSize <= 0 {
		pageSize = 10
	}
	vt.ScrollUp(pageSize)
}

// ScrollToTop moves selection to the very first row.
func (vt *VirtualTable) ScrollToTop() {
	vt.Select(0)
}

// ScrollToBottom moves selection to the very last row.
func (vt *VirtualTable) ScrollToBottom() {
	if vt.totalRows > 0 {
		vt.Select(vt.totalRows - 1)
	}
}

// HandleKey processes keyboard navigation for the table (Up/Down, PgUp/PgDn, Home/End, j/k/g/G).
func (vt *VirtualTable) HandleKey(k input.Key, pageSize int) bool {
	if pageSize <= 0 {
		pageSize = 10
	}

	switch k.Type {
	case input.KeyUp:
		vt.ScrollUp(1)
		return true
	case input.KeyDown:
		vt.ScrollDown(1)
		return true
	case input.KeyPgUp:
		vt.PageUp(pageSize)
		return true
	case input.KeyPgDown:
		vt.PageDown(pageSize)
		return true
	case input.KeyHome:
		vt.ScrollToTop()
		return true
	case input.KeyEnd:
		vt.ScrollToBottom()
		return true
	case input.KeyRune:
		if !k.HasCtrl() && !k.HasAlt() {
			switch k.Rune {
			case 'k':
				vt.ScrollUp(1)
				return true
			case 'j':
				vt.ScrollDown(1)
				return true
			case 'g':
				vt.ScrollToTop()
				return true
			case 'G':
				vt.ScrollToBottom()
				return true
			}
		}
	}
	return false
}

// SelectNext moves selection to the next row.
func (vt *VirtualTable) SelectNext() {
	vt.ScrollDown(1)
}

// SelectPrev moves selection to the previous row.
func (vt *VirtualTable) SelectPrev() {
	vt.ScrollUp(1)
}

// Selected returns the currently selected row index.
func (vt *VirtualTable) Selected() int {
	return vt.selectedRow
}

// SelectedRow returns the currently selected row index.
func (vt *VirtualTable) SelectedRow() int {
	return vt.selectedRow
}

// TotalRows returns the total count of rows in the table.
func (vt *VirtualTable) TotalRows() int {
	return vt.totalRows
}

// HandleMouse processes mouse events for the VirtualTable.
func (vt *VirtualTable) HandleMouse(msg tea.MouseMsg, area buffer.Rect) bool {
	if area.IsEmpty() || !area.Contains(msg.X, msg.Y) {
		return false
	}

	if msg.Button == input.MouseWheelUp {
		vt.ScrollUp(3)
		return true
	}
	if msg.Button == input.MouseWheelDown {
		vt.ScrollDown(3)
		return true
	}

	if msg.Button == input.MouseLeft && msg.Action == input.MousePress {
		row := vt.RowAt(area, msg.Y)
		if row != -1 {
			vt.Select(row)
			return true
		}
	}
	return false
}

// RowAt calculates the dataset row index under absolute screen coordinate y, or -1 if outside visible body rows.
func (vt *VirtualTable) RowAt(area buffer.Rect, y int) int {
	if area.IsEmpty() || vt.totalRows == 0 {
		return -1
	}

	topOffset := 0
	if vt.border.HasOuter {
		topOffset++
	}
	// Header row
	topOffset++
	if vt.border.HasHeaderSep {
		topOffset++
	}

	rowStartY := area.Y + topOffset
	relativeY := y - rowStartY
	if relativeY < 0 {
		return -1
	}

	bottomLimit := area.Bottom()
	if vt.border.HasOuter {
		bottomLimit--
	}
	if y >= bottomLimit {
		return -1
	}

	targetRow := vt.offset + relativeY
	if targetRow >= 0 && targetRow < vt.totalRows {
		return targetRow
	}
	return -1
}

func (vt *VirtualTable) calculateLayout(totalWidth int, colX, colW []int) {
	n := len(vt.columns)
	if n == 0 {
		return
	}

	startX := 0
	availWidth := totalWidth
	if vt.border.HasOuter {
		availWidth -= 2 // 1 left, 1 right border
		startX = 1
	}

	sepWidth := 0
	if vt.border.HasColSep {
		sepWidth = 1
	}
	totalSepWidth := (n - 1) * sepWidth
	contentWidth := availWidth - totalSepWidth
	if contentWidth < n {
		contentWidth = n
	}

	sumFixed := 0
	sumFlex := 0
	for _, col := range vt.columns {
		if col.Flex > 0 {
			sumFlex += col.Flex
		} else {
			w := col.Width
			if w <= 0 {
				w = 10
			}
			sumFixed += w
		}
	}

	remaining := contentWidth - sumFixed
	allocatedFlex := 0

	for i, col := range vt.columns {
		if col.Flex > 0 {
			w := 0
			if sumFlex > 0 && remaining > 0 {
				w = (remaining * col.Flex) / sumFlex
			}
			minW := col.MinWidth
			if minW <= 0 {
				minW = 4
			}
			if w < minW {
				w = minW
			}
			colW[i] = w
			allocatedFlex += w
		} else {
			w := col.Width
			if w <= 0 {
				w = 10
			}
			colW[i] = w
		}
	}

	// Distribute any rounding remainder across flex columns
	if sumFlex > 0 && remaining > allocatedFlex {
		diff := remaining - allocatedFlex
		for i, col := range vt.columns {
			if col.Flex > 0 && diff > 0 {
				colW[i]++
				diff--
			}
		}
	}

	currX := startX
	for i := range vt.columns {
		colX[i] = currX
		currX += colW[i] + sepWidth
	}
}

// resolveColumnLayout caches colX and colW slices to guarantee 0 allocs/op during Draw.
func (vt *VirtualTable) resolveColumnLayout(totalWidth int) ([]int, []int) {
	n := len(vt.columns)
	if n == 0 {
		return nil, nil
	}
	if len(vt.cachedColX) != n || len(vt.cachedColW) != n {
		vt.cachedColX = make([]int, n)
		vt.cachedColW = make([]int, n)
		vt.cachedWidth = -1
	}
	if vt.cachedWidth != totalWidth {
		vt.cachedWidth = totalWidth
		vt.calculateLayout(totalWidth, vt.cachedColX, vt.cachedColW)
	}
	return vt.cachedColX, vt.cachedColW
}

// ComputeColumnLayout resolves column widths and X positions given the available total width.
func (vt *VirtualTable) ComputeColumnLayout(totalWidth int) ([]int, []int) {
	n := len(vt.columns)
	if n == 0 {
		return nil, nil
	}
	colX := make([]int, n)
	colW := make([]int, n)
	vt.calculateLayout(totalWidth, colX, colW)
	return colX, colW
}

// Draw renders the virtual table into the buffer at the given area.
func (vt *VirtualTable) Draw(buf *buffer.Buffer, area buffer.Rect) {
	if area.IsEmpty() || area.Height < 1 || len(vt.columns) == 0 {
		return
	}

	colX, colW := vt.resolveColumnLayout(area.Width)
	nCols := len(vt.columns)

	currY := area.Y
	borderFg := vt.borderStyle.GetFg()
	borderBg := vt.borderStyle.GetBg()

	// 1. Draw Top Border (if HasOuter)
	if vt.border.HasOuter && currY < area.Bottom() {
		buf.SetRune(area.X, currY, vt.border.TopLeft, borderFg, borderBg, cell.AttrNone)
		for i := 0; i < nCols; i++ {
			baseX := area.X + colX[i]
			for cx := 0; cx < colW[i]; cx++ {
				buf.SetRune(baseX+cx, currY, vt.border.Top, borderFg, borderBg, cell.AttrNone)
			}
			if i < nCols-1 && vt.border.HasColSep {
				buf.SetRune(baseX+colW[i], currY, vt.border.TopCross, borderFg, borderBg, cell.AttrNone)
			}
		}
		buf.SetRune(area.Right()-1, currY, vt.border.TopRight, borderFg, borderBg, cell.AttrNone)
		currY++
	}

	// 2. Draw Sticky Header Row
	if currY < area.Bottom() {
		if vt.border.HasOuter {
			buf.SetRune(area.X, currY, vt.border.Left, borderFg, borderBg, cell.AttrNone)
			buf.SetRune(area.Right()-1, currY, vt.border.Right, borderFg, borderBg, cell.AttrNone)
		}

		for i, col := range vt.columns {
			baseX := area.X + colX[i]
			cellRect := buffer.NewRect(baseX, currY, colW[i], 1)
			align := col.HeaderAlign
			if align == AlignLeft && col.Align != AlignLeft {
				align = col.Align
			}
			RenderCellText(buf, cellRect, col.Title, align, vt.padding, vt.headerStyle.GetFg(), vt.headerStyle.GetBg(), vt.headerStyle.GetModifier())

			if i < nCols-1 && vt.border.HasColSep {
				buf.SetRune(baseX+colW[i], currY, vt.border.ColSep, borderFg, borderBg, cell.AttrNone)
			}
		}
		currY++
	}

	// 3. Draw Header Separator Line
	if vt.border.HasHeaderSep && currY < area.Bottom() {
		hBorderFg := vt.headerBorderStyle.GetFg()
		hBorderBg := vt.headerBorderStyle.GetBg()

		if vt.border.HasOuter {
			buf.SetRune(area.X, currY, vt.border.HeaderLeft, hBorderFg, hBorderBg, cell.AttrNone)
			buf.SetRune(area.Right()-1, currY, vt.border.HeaderRight, hBorderFg, hBorderBg, cell.AttrNone)
		}

		for i := 0; i < nCols; i++ {
			baseX := area.X + colX[i]
			for cx := 0; cx < colW[i]; cx++ {
				buf.SetRune(baseX+cx, currY, vt.border.HeaderSep, hBorderFg, hBorderBg, cell.AttrNone)
			}
			if i < nCols-1 && vt.border.HasColSep {
				buf.SetRune(baseX+colW[i], currY, vt.border.HeaderCross, hBorderFg, hBorderBg, cell.AttrNone)
			}
		}
		currY++
	}

	// 4. Calculate Body Viewport
	bottomLimit := area.Bottom()
	if vt.border.HasOuter {
		bottomLimit--
	}
	bodyHeight := bottomLimit - currY
	if bodyHeight <= 0 {
		// Draw bottom border if needed and return
		if vt.border.HasOuter && currY == bottomLimit {
			vt.drawBottomBorder(buf, area, colX, colW, currY, borderFg, borderBg)
		}
		return
	}

	// Clamp viewport scrolling offset
	if vt.selectedRow < vt.offset {
		vt.offset = vt.selectedRow
	} else if vt.selectedRow >= vt.offset+bodyHeight {
		vt.offset = vt.selectedRow - bodyHeight + 1
	}

	maxOffset := max(0, vt.totalRows-bodyHeight)
	if vt.offset > maxOffset {
		vt.offset = maxOffset
	}
	if vt.offset < 0 {
		vt.offset = 0
	}

	end := min(vt.totalRows, vt.offset+bodyHeight)

	// 5. Draw Visible Data Rows
	for r := vt.offset; r < end && currY < bottomLimit; r, currY = r+1, currY+1 {
		isSelected := (r == vt.selectedRow)

		// Determine row style
		st := vt.rowStyle
		if isSelected {
			st = vt.selectedStyle
		} else if vt.zebra && (r%2 == 1) {
			st = vt.alternateRowStyle
		}

		if vt.border.HasOuter {
			buf.SetRune(area.X, currY, vt.border.Left, borderFg, borderBg, cell.AttrNone)
			buf.SetRune(area.Right()-1, currY, vt.border.Right, borderFg, borderBg, cell.AttrNone)
		}

		// Backward-compatible TableRowRenderer
		if vt.legacyRenderer != nil {
			rowArea := buffer.NewRect(area.X, currY, area.Width, 1)
			vt.legacyRenderer(buf, vt.columns, rowArea, r, isSelected)
			continue
		}

		// Fetch row data strings if available
		var rowValues []string
		if vt.rowProvider != nil {
			rowValues = vt.rowProvider(r)
		} else if r < len(vt.staticRows) {
			rowValues = vt.staticRows[r]
		}

		for i, col := range vt.columns {
			baseX := area.X + colX[i]
			cellRect := buffer.NewRect(baseX, currY, colW[i], 1)

			handled := false
			if vt.cellRenderer != nil {
				handled = vt.cellRenderer(buf, col, cellRect, r, i, isSelected)
			}

			if !handled {
				val := ""
				if i < len(rowValues) {
					val = rowValues[i]
				}

				textToDraw := val
				if i == 0 && isSelected && vt.selectionPrefix != "" {
					textToDraw = vt.selectionPrefix + val
				}

				RenderCellText(buf, cellRect, textToDraw, col.Align, vt.padding, st.GetFg(), st.GetBg(), st.GetModifier())
			}

			if i < nCols-1 && vt.border.HasColSep {
				buf.SetRune(baseX+colW[i], currY, vt.border.ColSep, borderFg, borderBg, cell.AttrNone)
			}
		}
	}

	// 6. Clear Empty Remaining Rows in Viewport
	for ; currY < bottomLimit; currY++ {
		if vt.border.HasOuter {
			buf.SetRune(area.X, currY, vt.border.Left, borderFg, borderBg, cell.AttrNone)
			buf.SetRune(area.Right()-1, currY, vt.border.Right, borderFg, borderBg, cell.AttrNone)
		}
		for i := 0; i < nCols; i++ {
			baseX := area.X + colX[i]
			cellRect := buffer.NewRect(baseX, currY, colW[i], 1)
			for cx := 0; cx < cellRect.Width; cx++ {
				buf.SetRune(cellRect.X+cx, cellRect.Y, ' ', cell.DefaultColor(), cell.DefaultColor(), cell.AttrNone)
			}
			if i < nCols-1 && vt.border.HasColSep {
				buf.SetRune(baseX+colW[i], currY, vt.border.ColSep, borderFg, borderBg, cell.AttrNone)
			}
		}
	}

	// 7. Draw Bottom Border (if HasOuter)
	if vt.border.HasOuter && currY == bottomLimit {
		vt.drawBottomBorder(buf, area, colX, colW, currY, borderFg, borderBg)
	}
}

func (vt *VirtualTable) drawBottomBorder(buf *buffer.Buffer, area buffer.Rect, colX, colW []int, y int, fg, bg cell.Color) {
	nCols := len(vt.columns)
	buf.SetRune(area.X, y, vt.border.BottomLeft, fg, bg, cell.AttrNone)
	for i := 0; i < nCols; i++ {
		baseX := area.X + colX[i]
		for cx := 0; cx < colW[i]; cx++ {
			buf.SetRune(baseX+cx, y, vt.border.Bottom, fg, bg, cell.AttrNone)
		}
		if i < nCols-1 && vt.border.HasColSep {
			buf.SetRune(baseX+colW[i], y, vt.border.BottomCross, fg, bg, cell.AttrNone)
		}
	}
	buf.SetRune(area.Right()-1, y, vt.border.BottomRight, fg, bg, cell.AttrNone)
}

// RenderCellText draws aligned and padded text inside cellRect, clipping overflow and appending '…' if truncated.
// Guaranteed 0 B/op and 0 allocs/op on hot paths.
func RenderCellText(buf *buffer.Buffer, cellRect buffer.Rect, text string, align Alignment, padding int, fg, bg cell.Color, mod cell.Modifier) {
	cellW := cellRect.Width
	if cellW <= 0 {
		return
	}

	// Pre-fill background
	for cx := 0; cx < cellW; cx++ {
		buf.Set(cellRect.X+cx, cellRect.Y, cell.Cell{
			Rune:     ' ',
			Width:    1,
			Modifier: mod,
			FgType:   fg.Type,
			BgType:   bg.Type,
			Fg:       fg.Value,
			Bg:       bg.Value,
		})
	}

	padLeft := padding
	padRight := padding
	if padLeft+padRight >= cellW {
		padLeft = 0
		padRight = 0
	}
	availW := cellW - padLeft - padRight
	if availW <= 0 {
		return
	}

	textW := buffer.StringWidth(text)

	// Case 1: Fits within available cell width without truncation
	if textW <= availW {
		startX := cellRect.X + padLeft
		switch align {
		case AlignRight:
			startX = cellRect.X + cellRect.Width - padRight - textW
		case AlignCenter:
			startX = cellRect.X + padLeft + (availW-textW)/2
		}

		currX := startX
		for _, r := range text {
			rw := buffer.RuneWidth(r)
			if rw == 0 {
				continue
			}
			buf.SetRune(currX, cellRect.Y, r, fg, bg, mod)
			currX += rw
		}
		return
	}

	// Case 2: Exceeds available cell width: truncate and append '…'
	maxFit := availW - 1
	if maxFit < 0 {
		maxFit = 0
	}
	currX := cellRect.X + padLeft
	fittedW := 0
	for _, r := range text {
		rw := buffer.RuneWidth(r)
		if rw == 0 {
			continue
		}
		if fittedW+rw > maxFit {
			break
		}
		buf.SetRune(currX, cellRect.Y, r, fg, bg, mod)
		currX += rw
		fittedW += rw
	}
	if maxFit >= 0 && currX < cellRect.Right() {
		buf.SetRune(currX, cellRect.Y, '…', fg, bg, mod)
	}
}
