package vm

import (
	"context"

	"github.com/pyrorhythm/spqt/internal/types"
	"github.com/pyrorhythm/spqt/pkg/reactive"
)

type MetaVM string

func (m MetaVM) String() string {
	return string(m)
}

const (
	MetaAuth MetaVM = "auth"
	MetaMain MetaVM = "main"
)

func (m MetaVM) Index() int {
	switch m {
	case MetaAuth:
		return 0
	case MetaMain:
		return 1
	}

	panic("unreachable")
}

type App struct {
	Current *reactive.Prop[MetaVM]
	Auth    *Auth
	Shell   *Shell
}

func New(ctx context.Context, auth types.Authenticator, clientFactory func(types.Session) types.Client) *App {
	app := &App{Current: reactive.NewProp(MetaAuth)}

	app.Auth = newAuthVM(auth)
	app.Shell = newShell()

	app.Auth.State.OnExact(ASReady, func() {
		app.Shell.BindClient(ctx, clientFactory(app.Auth.Session))
		app.Current.Set(MetaMain)
	})

	return app
}
