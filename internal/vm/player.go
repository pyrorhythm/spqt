package vm

import (
	"context"
)

type Player struct {
}

func (v Player) Create(ctx context.Context) *Player {
	pvm := &Player{}

	return pvm
}
