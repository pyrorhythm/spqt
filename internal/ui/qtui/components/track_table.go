package components

import (
	"context"
	"fmt"
	"strings"

	qt "github.com/mappu/miqt/qt6"

	"github.com/pyrorhythm/spqt/internal/types"
	"github.com/pyrorhythm/spqt/internal/vm"
)

func BuildTrackTable(ctx context.Context, tl *vm.TrackList) *qt.QWidget {
	wrapper := qt.NewQWidget2()
	layout := qt.NewQVBoxLayout2()
	layout.SetContentsMargins(0, 0, 0, 0)
	layout.SetSpacing(4)
	wrapper.SetLayout(layout.QLayout)

	table := qt.NewQTableWidget2()
	table.SetColumnCount(5)
	table.SetHorizontalHeaderLabels([]string{"#", "Title", "Artist", "Album", "Duration"})
	table.SetSelectionBehavior(qt.QAbstractItemView__SelectRows)
	table.SetEditTriggers(qt.QAbstractItemView__NoEditTriggers)
	table.HorizontalHeader().SetStretchLastSection(true)
	table.VerticalHeader().SetVisible(false)
	table.SetShowGrid(false)
	table.SetAlternatingRowColors(true)
	layout.AddWidget(table.QWidget)

	loadBtn := qt.NewQPushButton3("Load liked tracks")
	layout.AddWidget(loadBtn.QWidget)

	statusLbl := qt.NewQLabel2()
	layout.AddWidget(statusLbl.QWidget)

	// Bindings
	table.OnCellClicked(func(row int, column int) {
		tl.Select(row)
	})

	loadBtn.OnClicked(func() {
		tl.LoadCmd.Execute(ctx)
	})

	tl.Loading.OnChange(func(loading bool) {
		loadBtn.SetEnabled(!loading)
	})

	tl.Status.OnChange(func(s string) {
		statusLbl.SetText(s)
	})

	tl.Tracks.OnChange(func() {
		fillTable(table, tl.Tracks.Items())
	})

	return wrapper
}

func fillTable(table *qt.QTableWidget, tracks []types.EnrichedTrack) {
	table.SetRowCount(len(tracks))

	for i, et := range tracks {
		t := et.Track

		table.SetItem(i, 0, qt.NewQTableWidgetItem2(fmt.Sprintf("%d", i+1)))
		table.SetItem(i, 1, qt.NewQTableWidgetItem2(t.Name))

		names := make([]string, 0, len(t.Artists))
		for _, a := range t.Artists {
			names = append(names, a.Name)
		}
		table.SetItem(i, 2, qt.NewQTableWidgetItem2(strings.Join(names, ", ")))
		table.SetItem(i, 3, qt.NewQTableWidgetItem2(t.Album.Name))

		secs := t.DurationMs / 1000
		table.SetItem(i, 4, qt.NewQTableWidgetItem2(fmt.Sprintf("%d:%02d", secs/60, secs%60)))
	}

	table.ResizeColumnsToContents()
}
