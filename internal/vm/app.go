package vm

import (
	"context"

	"github.com/pyrorhythm/spqt/pkg/reactive"
)

type CurrentViewModel string

const (
	CVMAuth   CurrentViewModel = "auth"
	CVMPlayer CurrentViewModel = "player"
)

type App struct {
	Current *reactive.Prop[CurrentViewModel]
	//
	Auth   *Auth
	Player *Player
}

func (App) Create(ctx context.Context) *App {
	app := &App{Current: reactive.NewProp[CurrentViewModel](CVMAuth)}

	app.Auth = Auth{}.Create(ctx)
	app.Player = Player{}.Create(ctx)

	return app
}
