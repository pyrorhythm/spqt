package components

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	qt "github.com/mappu/miqt/qt6"

	"github.com/pyrorhythm/spqt/internal/types"
	"github.com/pyrorhythm/spqt/internal/vm"
	"github.com/pyrorhythm/spqt/pkg/log"
)

func BuildPlayerBar(pvm *vm.Player) *qt.QWidget {
	bar := qt.NewQWidget2()
	bar.SetObjectName(*qt.NewQAnyStringView3("playerBar"))
	bar.SetFixedHeight(80)

	root := qt.NewQHBoxLayout2()
	root.SetContentsMargins(4, 2, 4, 2)
	bar.SetLayout(root.QLayout)

	// --- Left: cover + meta ---
	coverLbl := qt.NewQLabel2()
	coverLbl.SetFixedSize2(56, 56)
	coverLbl.SetScaledContents(true)
	placeholder := qt.NewQPixmap2(56, 56)
	placeholder.Fill()
	coverLbl.SetPixmap(placeholder)
	coverLbl.SetAlignment(qt.AlignLeft)
	root.AddWidget(coverLbl.QWidget)

	metaBox := qt.NewQVBoxLayout2()
	metaBox.SetSpacing(2)

	trackLbl := qt.NewQLabel3("—")
	trackLbl.SetObjectName(*qt.NewQAnyStringView3("playerTrackName"))

	artistLbl := qt.NewQLabel2()
	artistLbl.SetObjectName(*qt.NewQAnyStringView3("playerArtistName"))

	metaBox.AddWidget(trackLbl.QWidget)
	metaBox.AddWidget(artistLbl.QWidget)
	metaBox.AddStretch()
	root.AddLayout(metaBox.QLayout)

	// --- Center: controls + progress ---
	center := qt.NewQVBoxLayout2()
	center.SetSpacing(2)

	controls := qt.NewQHBoxLayout2()
	controls.SetSpacing(12)

	playBtn := qt.NewQPushButton3("▶")
	playBtn.SetFixedSize2(36, 36)

	nextBtn := qt.NewQPushButton3("⏭")
	nextBtn.SetFixedSize2(36, 36)

	controls.AddStretch()
	controls.AddWidget(playBtn.QWidget)
	controls.AddWidget(nextBtn.QWidget)
	controls.AddStretch()

	progressRow := qt.NewQHBoxLayout2()
	progressRow.SetSpacing(6)

	posLbl := qt.NewQLabel3("0:00")
	slider := qt.NewQSlider3(qt.Horizontal)
	slider.SetRange(0, 1000)
	durLbl := qt.NewQLabel3("0:00")

	progressRow.AddWidget(posLbl.QWidget)
	progressRow.AddWidget(slider.QWidget)
	progressRow.AddWidget(durLbl.QWidget)

	center.AddLayout(controls.QLayout)
	center.AddLayout(progressRow.QLayout)

	root.AddLayout(center.QLayout)
	root.SetStretch(2, 1)

	// --- Bindings ---
	pvm.Current.OnChange(func(et types.EnrichedTrack) {
		if et.Track == nil {
			trackLbl.SetText("—")
			artistLbl.SetText("")
			durLbl.SetText("0:00")
			slider.SetValue(0)
			placeholder2 := qt.NewQPixmap2(56, 56)
			placeholder2.Fill()
			coverLbl.SetPixmap(placeholder2)
			return
		}

		trackLbl.SetText(et.Name)

		names := make([]string, 0, len(et.Artists))
		for _, a := range et.Artists {
			names = append(names, a.Name)
		}
		artistLbl.SetText(strings.Join(names, ", "))

		dur := et.DurationMs / 1000
		durLbl.SetText(fmt.Sprintf("%d:%02d", dur/60, dur%60))

		// Load album art
		if et.FullAlbum != nil && len(et.FullAlbum.Covers) > 0 {
			url := et.FullAlbum.Covers[0].URL
			go func() {
				data, err := fetchImage(url)
				if err != nil {
					log.Logger().Warn().Err(err).Str("url", url).Msg("failed to fetch cover")
					return
				}
				pm := qt.NewQPixmap()
				if pm.LoadFromDataWithData(data) {
					coverLbl.SetPixmap(pm)
				}
			}()
		}
	})

	pvm.IsPlaying.OnChange(func(playing bool) {
		if playing {
			playBtn.SetText("⏸")
		} else {
			playBtn.SetText("▶")
		}
	})

	pvm.Progress.OnChange(func(pct float64) {
		slider.SetValue(int(pct * 1000))

		cur := pvm.Current.Get()
		if cur.Track != nil {
			pos := int32(pct * float64(cur.DurationMs) / 1000)
			posLbl.SetText(fmt.Sprintf("%d:%02d", pos/60, pos%60))
		}
	})

	if pvm.CanNext != nil {
		pvm.CanNext.OnChange(func(can bool) {
			nextBtn.SetEnabled(can)
		})
	}

	playBtn.OnClicked(func() {
		if pvm.PlayCmd != nil {
			pvm.PlayCmd.Execute()
		}
	})

	nextBtn.OnClicked(func() {
		if pvm.NextCmd != nil {
			pvm.NextCmd.Execute()
		}
	})

	return bar
}

func fetchImage(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
