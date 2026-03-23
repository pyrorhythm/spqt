package vm

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"sync/atomic"

	"github.com/pyrorhythm/spqt/internal/types"
	"github.com/pyrorhythm/spqt/pkg/log"
	"github.com/pyrorhythm/spqt/pkg/reactive"
)

type AuthState string

const (
	ASChecking    AuthState = "Checking"
	ASNeedsLogin  AuthState = "Needs login"
	ASAuthorizing AuthState = "Authorizing"
	ASReady       AuthState = "Ready"
	ASAuthError   AuthState = "Authorization error"
)

type Auth struct {
	State    *reactive.Prop[AuthState]
	Error    *reactive.Prop[error]
	LoginCmd *reactive.CtxCommand
	Session  types.Session

	auth               types.Authenticator
	eventReaderRunning atomic.Bool
}

func openWithBrowser(url string) error {
	switch runtime.GOOS {
	case "linux":
		return exec.Command("xdg-open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		return fmt.Errorf("runtime.GOOS %s is not supported", runtime.GOOS)
	}
}

func (v *Auth) spawnEventReader(ech <-chan types.Event) {
	go func() {
		for e := range ech {
			log.Logger().Trace().Any("e", e).Msg("got event")
			switch typedE := e.(type) {
			case types.SessionAuthorizedEvent:
				v.Session = typedE.Session
				v.State.Set(ASReady)
			case types.LinkEvent:
				v.State.Set(ASNeedsLogin)
				openWithBrowser(typedE.Link)
			case types.CodeReceivedEvent:
				v.State.Set(ASAuthorizing)
			case types.FailedEvent:
				v.Error.Set(typedE.Error)
				v.State.Set(ASAuthError)
			}
		}
	}()
}

func newAuthVM(auth types.Authenticator) *Auth {
	v := &Auth{
		State: reactive.NewProp[AuthState](ASChecking),
		Error: reactive.NewProp[error](nil),
		auth:  auth,
	}

	v.LoginCmd = reactive.NewCtxCommand(
		func(ctx context.Context) { v.spawnEventReader(v.auth.Authorize(ctx)) },
		func(context.Context) bool { return !v.eventReaderRunning.Load() },
	)

	return v
}
