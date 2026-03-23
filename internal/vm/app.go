package vm

import (
	"github.com/pyrorhythm/spqt/internal/respot"
	"github.com/pyrorhythm/spqt/internal/types"
	"github.com/pyrorhythm/spqt/pkg/reactive"
)

type CurrentViewModel string

const (
	CVMAuth   CurrentViewModel = "auth"
	CVMPlayer CurrentViewModel = "player"
)

func (cvm CurrentViewModel) Index() int {
	switch cvm {
	case CVMAuth:
		return 0
	case CVMPlayer:
		return 1
	}

	panic("unreachable")
}

type App struct {
	Current   *reactive.Prop[CurrentViewModel]
	Auth      *Auth
	Player    *Player
	TrackList *TrackList
}

func New(auth types.Authenticator) *App {
	app := &App{Current: reactive.NewProp[CurrentViewModel](CVMAuth)}

	app.Auth = newAuthVM(auth)
	app.Player = newPlayerVM()
	app.TrackList = newTrackListVM(app.Player)

	app.Auth.State.OnExact(ASReady, func() {
		app.Current.Set(CVMPlayer)
		app.TrackList.SetClient(respot.NewClient(app.Auth.Session))
	})

	return app
}
