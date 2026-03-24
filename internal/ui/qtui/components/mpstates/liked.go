package mpstates

import (
	"context"
	"strings"

	qt "github.com/mappu/miqt/qt6"

	"github.com/pyrorhythm/spqt/internal/types"
	"github.com/pyrorhythm/spqt/internal/vm"
	"github.com/pyrorhythm/spqt/pkg/qtw"
)

func artistsStr(ar []types.ArtistRef) string {
	var ans = make([]string, len(ar))

	for _, a := range ar {
		ans = append(ans, a.Name)
	}

	return strings.Join(ans, ", ")
}

func BuildLikedTracks(ctx context.Context, sh *vm.Shell) func() *qt.QWidget {
	return func() *qt.QWidget {
		cr, _ := sh.Client.LikedTracks(ctx)
		ts, _ := sh.Client.EnrichPage(ctx, cr, 0)

		tbl := qtw.
			Table().Name("Liked tracks").
			NoVertHeader().Cols("Title", "Artist", "Album", "Duration").
			OnRowClick(func(row int) { sh.Player.Current.Set(ts[row]) }).
			Build()

		qtw.FillTable(tbl, len(ts), func(row int) []string {
			et := ts[row]

			if et.Track == nil {
				return nil
			}

			return []string{
				et.Track.Name,
				artistsStr(et.Track.Artists),
				et.Album.Name,
				qtw.FormatDuration(et.DurationMs),
			}
		})

		return qtw.Widget().Layout(
			qtw.VBox().Margins(24, 24, 24, 24).Spacing(16).Items(
				qtw.Label("Liked Songs").
					Property("heading", qt.NewQVariant14("true")).
					Build(),
				qtw.Stretch(),
				tbl,
			),
		).Build()
	}
}
