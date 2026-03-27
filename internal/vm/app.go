package vm

import (
	"context"

	metadatapb "github.com/devgianlu/go-librespot/proto/spotify/metadata"
	"github.com/dgraph-io/badger/v4"
	"google.golang.org/protobuf/proto"

	"github.com/pyrorhythm/spqt/internal/types"
	"github.com/pyrorhythm/spqt/pkg/cache"
	"github.com/pyrorhythm/spqt/pkg/log"
	"github.com/pyrorhythm/spqt/pkg/reactive"
)

type AppState string

func (m AppState) String() string {
	return string(m)
}

const (
	StateAuth AppState = "auth"
	StateMain AppState = "main"
)

func (m AppState) Index() int {
	switch m {
	case StateAuth:
		return 0
	case StateMain:
		return 1
	}

	panic("unreachable")
}

type App struct {
	State        *reactive.Prop[AppState]
	Auth         *Auth
	Player       *Player
	Nav          *reactive.Prop[NavState]
	Client       types.Client
	Images       *ImageService
	LikedTracks  *TrackListVM
	SearchResults *TrackListVM
}

func New(ctx context.Context, auth types.Authenticator, clientFactory func(context.Context, types.Session) types.Client, db *badger.DB) *App {
	ctx = log.Span(ctx, "vm")

	trackLRU := cache.NewLRU(db, "tracks:liked:", 50,
		func(t *metadatapb.Track) []byte { b, _ := proto.Marshal(t); return b },
		func(b []byte) *metadatapb.Track { var t metadatapb.Track; _ = proto.Unmarshal(b, &t); return &t },
	)

	app := &App{
		State:        reactive.NewProp(StateAuth),
		Player:       newPlayer(),
		Nav:          reactive.NewProp(NavHome),
		Images:       NewImageService(ctx, db),
		LikedTracks:  NewTrackListVM(trackLRU),
		SearchResults: NewTrackListVM(cache.NewLRU(db, "tracks:search:", 200,
			func(t *metadatapb.Track) []byte { b, _ := proto.Marshal(t); return b },
			func(b []byte) *metadatapb.Track { var t metadatapb.Track; _ = proto.Unmarshal(b, &t); return &t },
		)),
	}

	app.Auth = newAuthVM(ctx, auth)

	app.Auth.State.OnExact(ASReady, func() {
		app.Client = clientFactory(ctx, app.Auth.Session)
		app.Player.BindClient(ctx, app.Client)
		app.State.Set(StateMain)
	})

	return app
}
