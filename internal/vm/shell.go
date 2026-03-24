package vm

import (
	"context"

	"github.com/pyrorhythm/spqt/internal/types"
	"github.com/pyrorhythm/spqt/pkg/reactive"
)

type Shell struct {
	Player *Player
	CurCtx *reactive.Prop[SidebarState]

	Client types.Client
}

func newShell() *Shell {
	sh := &Shell{
		Player: newPlayer(),
		CurCtx: reactive.NewProp(SbHome),
	}

	return sh
}

func (s *Shell) BindClient(ctx context.Context, c types.Client) {
	s.Client = c
	s.Player.BindClient(ctx, c)
}
