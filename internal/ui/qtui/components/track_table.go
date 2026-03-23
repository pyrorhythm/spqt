package components

import (
	"context"
	"fmt"
	"strings"

	qt "github.com/mappu/miqt/qt6"

	"github.com/pyrorhythm/spqt/internal/types"
	"github.com/pyrorhythm/spqt/internal/vm"
	"github.com/pyrorhythm/spqt/pkg/qtw"
)

func BuildTrackTable(ctx context.Context, tl *vm.TrackList) *qt.QWidget {
	wrapper := qtw.Widget()

	table := qtw.Table().
		Cols("#", "Title", "Artist", "Album", "Duration").
		SelectRows().ReadOnly().NoGrid().AlternatingRows().
		HStretchLast().NoVertHeader().
		OnRowClick(func(row int) { tl.Select(row) }).
		Build()

	loadBtn := qtw.Button("Load liked tracks").
		OnClick(func() { tl.LoadCmd.Execute(ctx) }).
		Build()

	statusLbl := qtw.EmptyLabel().Build()

	wrapper.SetLayout(qtw.VBox().NoMargins().Spacing(4).Items(
		table, loadBtn, statusLbl,
	))

	// Bindings
	tl.Loading.OnChange(func(loading bool) {
		loadBtn.SetEnabled(!loading)
	})

	tl.Status.OnChange(func(s string) {
		statusLbl.SetText(s)
	})

	qtw.BindTable(tl.Tracks, table, func(et types.EnrichedTrack, i int) []string {
		t := et.Track
		names := make([]string, 0, len(t.Artists))
		for _, a := range t.Artists {
			names = append(names, a.Name)
		}
		return []string{
			fmt.Sprintf("%d", i+1),
			t.Name,
			strings.Join(names, ", "),
			t.Album.Name,
			qtw.FormatDuration(t.DurationMs),
		}
	})

	return wrapper
}
