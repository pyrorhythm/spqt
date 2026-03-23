package qtui

import (
	"context"

	qt "github.com/mappu/miqt/qt6"

	"github.com/pyrorhythm/spqt/internal/vm"
	"github.com/pyrorhythm/spqt/pkg/log"
	"github.com/pyrorhythm/spqt/pkg/qtw"
)

func buildAuthPage(ctx context.Context, avm *vm.Auth) *qt.QWidget {
	page := qtw.Widget()

	status := qtw.EmptyLabel().Align(qt.AlignCenter).Font(font).Build()
	retryBtn := qtw.Button("Retry").Visible(false).
		OnClick(func() { avm.LoginCmd.Execute(ctx) }).
		Build()

	page.SetLayout(qtw.VBox().Items(status, retryBtn))

	avm.State.OnChange(func(s vm.AuthState) {
		log.Logger().Trace().Any("s", s).Msg("got state change")
		switch s {
		case vm.ASChecking:
			status.SetText("Connecting...")
			retryBtn.QWidget.SetVisible(false)
		case vm.ASNeedsLogin:
			status.SetText("Please, auth in opened window")
			retryBtn.QWidget.SetVisible(false)
		case vm.ASAuthorizing:
			status.SetText("Authorizing...")
			retryBtn.QWidget.SetVisible(false)
		case vm.ASAuthError:
			status.SetText("Error: " + avm.Error.Get().Error())
			retryBtn.QWidget.SetVisible(true)
		case vm.ASReady:
			status.SetText("Good to go!")
			retryBtn.QWidget.SetVisible(false)
		}
	})

	return page
}
