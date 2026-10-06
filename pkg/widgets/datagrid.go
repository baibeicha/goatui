package widgets

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
)

// DataGridColumn describes a single table column in the data grid.
type DataGridColumn struct {
	Name       string
	DataType   string
	IsPK       bool
	IsNullable bool
}

// DataGridTheme contains styling colors for the DataGrid table view.
type DataGridTheme struct {
	Background cell.Color
	Foreground cell.Color
	HeaderBg   cell.Color
	HeaderFg   cell.Color
	SelectedBg cell.Color
	SelectedFg cell.Color
	BorderFg   cell.Color
	AccentFg   cell.Color
	MutedFg    cell.Color
}

// DefaultDataGridTheme returns standard colors for DataGrid rendering.
func DefaultDataGridTheme() DataGridTheme {
	return DataGridTheme{
		Background: cell.RGB(24, 24, 37),
		Foreground: cell.RGB(205, 214, 244),
		HeaderBg:   cell.RGB(30, 30, 46),
		HeaderFg:   cell.RGB(137, 180, 250),
		SelectedBg: cell.RGB(69, 71, 90),
		SelectedFg: cell.RGB(255, 255, 255),
		BorderFg:   cell.RGB(49, 50, 68),
		AccentFg:   cell.RGB(249, 226, 175),
		MutedFg:    cell.RGB(108, 112, 134),
	}
}

// DataGridWidget provides an interactive tabular grid viewer.
type DataGridWidget struct {
	TableName    string
	Columns      []DataGridColumn
	FKColumns    map[string]bool
	AllRows      [][]string
	FilteredRows [][]string
	Page         int
	PageSize     int
	SelectedRow  int
	FilterText   string
	FilterActive bool
}

// NewDataGridWidget creates a new DataGrid widget with specified columns and rows.
func NewDataGridWidget(tableName string, cols []DataGridColumn, rows [][]string) *DataGridWidget {
	if len(cols) == 0 {
		cols = []DataGridColumn{
			{Name: "id", DataType: "uuid", IsPK: true},
			{Name: "name", DataType: "varchar(100)"},
		}
	}
	if rows == nil {
		rows = make([][]string, 0)
	}

	return &DataGridWidget{
		TableName:    tableName,
		Columns:      cols,
		FKColumns:    make(map[string]bool),
		AllRows:      rows,
		FilteredRows: rows,
		Page:         0,
		PageSize:     16,
		SelectedRow:  0,
	}
}

// TotalPages returns total number of pages based on filtered row count.
func (dg *DataGridWidget) TotalPages() int {
	if len(dg.FilteredRows) == 0 {
		return 1
	}
	pages := len(dg.FilteredRows) / dg.PageSize
	if len(dg.FilteredRows)%dg.PageSize != 0 {
		pages++
	}
	return pages
}

// NextPage advances to the next page of rows.
func (dg *DataGridWidget) NextPage() {
	totalPages := dg.TotalPages()
	if dg.Page < totalPages-1 {
		dg.Page++
		dg.SelectedRow = 0
	}
}

// PrevPage goes back to the previous page of rows.
func (dg *DataGridWidget) PrevPage() {
	if dg.Page > 0 {
		dg.Page--
		dg.SelectedRow = 0
	}
}

// SelectNextRow moves the row cursor down.
func (dg *DataGridWidget) SelectNextRow() {
	pageRows := dg.CurrentPageRows()
	if dg.SelectedRow < len(pageRows)-1 {
		dg.SelectedRow++
	}
}

// SelectPrevRow moves the row cursor up.
func (dg *DataGridWidget) SelectPrevRow() {
	if dg.SelectedRow > 0 {
		dg.SelectedRow--
	}
}

// SetFilter filters table rows matching text across all column cells.
func (dg *DataGridWidget) SetFilter(text string) {
	dg.FilterText = text
	clean := strings.ToLower(strings.TrimSpace(text))
	if clean == "" {
		dg.FilteredRows = dg.AllRows
		dg.Page = 0
		dg.SelectedRow = 0
		return
	}

	res := make([][]string, 0)
	for _, row := range dg.AllRows {
		match := false
		for _, val := range row {
			if strings.Contains(strings.ToLower(val), clean) {
				match = true
				break
			}
		}
		if match {
			res = append(res, row)
		}
	}
	dg.FilteredRows = res
	dg.Page = 0
	dg.SelectedRow = 0
}

// CurrentPageRows returns the slice of rows visible on the active page.
func (dg *DataGridWidget) CurrentPageRows() [][]string {
	if len(dg.FilteredRows) == 0 {
		return nil
	}
	start := dg.Page * dg.PageSize
	if start >= len(dg.FilteredRows) {
		start = 0
		dg.Page = 0
	}
	end := start + dg.PageSize
	if end > len(dg.FilteredRows) {
		end = len(dg.FilteredRows)
	}
	return dg.FilteredRows[start:end]
}

// ExportCSV writes the entire dataset or filtered rows to a CSV file.
func (dg *DataGridWidget) ExportCSV(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	var sb strings.Builder
	colNames := make([]string, len(dg.Columns))
	for i, c := range dg.Columns {
		colNames[i] = fmt.Sprintf("%q", c.Name)
	}
	sb.WriteString(strings.Join(colNames, ",") + "\n")

	for _, row := range dg.FilteredRows {
		escaped := make([]string, len(row))
		for i, v := range row {
			escaped[i] = fmt.Sprintf("%q", v)
		}
		sb.WriteString(strings.Join(escaped, ",") + "\n")
	}

	return os.WriteFile(path, []byte(sb.String()), 0644)
}

// Render draws the DataGrid tabular view into a GoatUI buffer.
func (dg *DataGridWidget) Render(buf *buffer.Buffer, startX, startY, width, height int, th DataGridTheme) {
	if buf == nil || width < 10 || height < 4 {
		return
	}

	// Fill background
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			buf.SetRune(startX+x, startY+y, ' ', th.Foreground, th.Background, cell.AttrNone)
		}
	}

	// Header row
	headerY := startY
	colW := width / (len(dg.Columns) + 1)
	if colW < 8 {
		colW = 8
	}

	currX := startX + 4
	for _, col := range dg.Columns {
		badge := "  "
		if col.IsPK {
			badge = "🔑"
		} else if dg.FKColumns[col.Name] {
			badge = "🔗"
		}
		hdr := fmt.Sprintf("%s %s", badge, col.Name)
		for i, r := range []rune(hdr) {
			if currX+i < startX+width {
				buf.SetRune(currX+i, headerY, r, th.HeaderFg, th.HeaderBg, cell.AttrBold)
			}
		}
		currX += colW
	}

	// Rows
	pageRows := dg.CurrentPageRows()
	for rIdx, row := range pageRows {
		rowY := startY + 2 + rIdx
		if rowY >= startY+height-1 {
			break
		}

		isSel := (rIdx == dg.SelectedRow)
		rowBg := th.Background
		rowFg := th.Foreground
		if isSel {
			rowBg = th.SelectedBg
			rowFg = th.SelectedFg
		}

		// Row number
		rNum := fmt.Sprintf("%2d", dg.Page*dg.PageSize+rIdx+1)
		for i, r := range rNum {
			buf.SetRune(startX+1+i, rowY, r, th.MutedFg, rowBg, cell.AttrNone)
		}

		cPos := startX + 4
		for cIdx, val := range row {
			if cIdx >= len(dg.Columns) {
				break
			}
			for i, r := range []rune(val) {
				if cPos+i < startX+width && i < colW-2 {
					buf.SetRune(cPos+i, rowY, r, rowFg, rowBg, cell.AttrNone)
				}
			}
			cPos += colW
		}
	}
}
