package vm

import (
	"github.com/pyrorhythm/spqt/internal/types"
	"github.com/pyrorhythm/spqt/pkg/reactive"
)

type Player struct {
	Current   *reactive.CmpProp[types.EnrichedTrack, string]
	IsPlaying *reactive.Prop[bool]
	Progress  *reactive.Prop[float64]
	CanNext   *reactive.Prop[bool]

	PlayCmd *reactive.Command
	NextCmd *reactive.Command
}

func newPlayerVM() *Player {
	return &Player{
		Current:   reactive.NewUProp[types.EnrichedTrack](types.EnrichedTrack{}, types.TrackComparator{}),
		IsPlaying: reactive.NewProp[bool](false),
		Progress:  reactive.NewProp[float64](0),
		CanNext:   reactive.NewProp[bool](true),
	}
}
