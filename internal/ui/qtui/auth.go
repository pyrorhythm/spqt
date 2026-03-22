package qtui

import (
	"context"

	qt "github.com/mappu/miqt/qt6"

	"github.com/pyrorhythm/spqt/internal/vm"
)

func buildAuthPage(ctx context.Context, avm *vm.Auth) *qt.QWidget {
	page := qt.NewQWidget2()
	layout := qt.NewQVBoxLayout2()
	page.SetLayout(layout.QLayout)

	status := qt.NewQLabel(page)
	retryBtn := qt.NewQPushButton5("Retry", page)

	layout.AddWidget(status.QWidget)
	layout.AddWidget(retryBtn.QWidget)

	avm.State.OnChange(func(s vm.AuthState) {
		switch s {
		case vm.ASChecking:
			status.SetText("Connecting...")
			retryBtn.QWidget.SetVisible(false)
		case vm.ASNeedsLogin:
			status.SetText("Please, authenticate in opened window")
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

	retryBtn.OnClicked(func() {
		avm.LoginCmd.Execute(ctx)
	})

	return page
}
