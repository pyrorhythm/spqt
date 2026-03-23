package qtw

import (
	qt "github.com/mappu/miqt/qt6"

	"github.com/pyrorhythm/spqt/pkg/reactive"
)

type TableBuilder struct{ t *qt.QTableWidget }

func Table() *TableBuilder {
	return &TableBuilder{t: qt.NewQTableWidget2()}
}

func (b *TableBuilder) Cols(headers ...string) *TableBuilder {
	b.t.SetColumnCount(len(headers))
	b.t.SetHorizontalHeaderLabels(headers)
	return b
}

func (b *TableBuilder) SelectRows() *TableBuilder {
	b.t.SetSelectionBehavior(qt.QAbstractItemView__SelectRows)
	return b
}

func (b *TableBuilder) ReadOnly() *TableBuilder {
	b.t.SetEditTriggers(qt.QAbstractItemView__NoEditTriggers)
	return b
}

func (b *TableBuilder) NoGrid() *TableBuilder {
	b.t.SetShowGrid(false)
	return b
}

func (b *TableBuilder) AlternatingRows() *TableBuilder {
	b.t.SetAlternatingRowColors(true)
	return b
}

func (b *TableBuilder) HStretchLast() *TableBuilder {
	b.t.HorizontalHeader().SetStretchLastSection(true)
	return b
}

func (b *TableBuilder) NoVertHeader() *TableBuilder {
	b.t.VerticalHeader().SetVisible(false)
	return b
}

func (b *TableBuilder) Name(n string) *TableBuilder {
	b.t.SetObjectName(*qt.NewQAnyStringView3(n))
	return b
}

func (b *TableBuilder) OnRowClick(fn func(row int)) *TableBuilder {
	b.t.OnCellClicked(func(row int, _ int) { fn(row) })
	return b
}

func (b *TableBuilder) Build() *qt.QTableWidget { return b.t }

// FillTable clears and fills a QTableWidget. render returns column values for each row.
func FillTable(table *qt.QTableWidget, rows int, render func(row int) []string) {
	table.SetRowCount(rows)
	for i := 0; i < rows; i++ {
		for j, val := range render(i) {
			table.SetItem(i, j, qt.NewQTableWidgetItem2(val))
		}
	}
	table.ResizeColumnsToContents()
}

// BindTable syncs a reactive.List to a QTableWidget. render returns column values for each item.
func BindTable[T any](list *reactive.List[T], table *qt.QTableWidget, render func(item T, index int) []string) {
	sync := func() {
		items := list.Items()
		table.SetRowCount(len(items))
		for i, item := range items {
			for j, val := range render(item, i) {
				table.SetItem(i, j, qt.NewQTableWidgetItem2(val))
			}
		}
		table.ResizeColumnsToContents()
	}
	sync()
	list.OnChange(sync)
}
