package qtw

import (
	qt "github.com/mappu/miqt/qt6"

	"github.com/pyrorhythm/spqt/pkg/reactive"
)

// Bind subscribes to a Prop and calls setter whenever it changes.
// Also calls setter immediately with the current value.
func Bind[T comparable](prop *reactive.Prop[T], setter func(T)) {
	setter(prop.Get())
	prop.OnChange(setter)
}

// BindText binds a Prop[string] to a QLabel's text.
func BindText(prop *reactive.Prop[string], label *qt.QLabel) {
	Bind(prop, label.SetText)
}

// BindEnabled binds a Prop[bool] to a QWidget's enabled state.
func BindEnabled(prop *reactive.Prop[bool], widget *qt.QWidget) {
	Bind(prop, widget.SetEnabled)
}

// BindVisible binds a Prop[bool] to a QWidget's visibility.
func BindVisible(prop *reactive.Prop[bool], widget *qt.QWidget) {
	Bind(prop, widget.SetVisible)
}

// BindCommand connects a Command to a QPushButton: click→Execute, CanExecute→SetEnabled.
func BindCommand(cmd *reactive.Command, btn *qt.QPushButton) {
	btn.QWidget.SetEnabled(cmd.CanExecute())
	btn.OnClicked(func() { cmd.Execute() })
	cmd.OnCanExecuteChanged(func(canExec bool) {
		btn.QWidget.SetEnabled(canExec)
	})
}

// BindList rebuilds a QListWidget whenever the List changes.
func BindList[T any](list *reactive.List[T], widget *qt.QListWidget, render func(T) string) {
	sync := func() {
		widget.Clear()
		for _, item := range list.Items() {
			widget.AddItem(render(item))
		}
	}
	sync()
	list.OnChange(sync)
}
