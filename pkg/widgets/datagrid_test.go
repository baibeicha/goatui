package widgets

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/baibeicha/goatui/pkg/core/buffer"
)

func TestDataGridWidget_PagingAndFiltering(t *testing.T) {
	cols := []DataGridColumn{
		{Name: "id", DataType: "int", IsPK: true},
		{Name: "user", DataType: "varchar"},
	}
	rows := [][]string{
		{"1", "alice"},
		{"2", "bob"},
		{"3", "charlie"},
	}

	dg := NewDataGridWidget("users", cols, rows)
	dg.PageSize = 2

	if dg.TotalPages() != 2 {
		t.Fatalf("expected 2 pages, got %d", dg.TotalPages())
	}

	dg.NextPage()
	if dg.Page != 1 {
		t.Fatalf("expected page 1, got %d", dg.Page)
	}

	dg.PrevPage()
	if dg.Page != 0 {
		t.Fatalf("expected page 0, got %d", dg.Page)
	}

	// Filter
	dg.SetFilter("bob")
	if len(dg.FilteredRows) != 1 || dg.FilteredRows[0][1] != "bob" {
		t.Fatalf("expected 1 filtered row for bob, got %v", dg.FilteredRows)
	}

	// CSV Export
	tmpFile := filepath.Join(t.TempDir(), "test_export.csv")
	if err := dg.ExportCSV(tmpFile); err != nil {
		t.Fatalf("export CSV failed: %v", err)
	}
	data, err := os.ReadFile(tmpFile)
	if err != nil || len(data) == 0 {
		t.Fatalf("read exported CSV failed: %v", err)
	}

	// Render
	buf := buffer.NewBuffer(80, 24)
	dg.Render(buf, 0, 0, 80, 24, DefaultDataGridTheme())
	if buf.Cell(4, 0) == nil {
		t.Fatal("expected rendered cell in datagrid")
	}
}
