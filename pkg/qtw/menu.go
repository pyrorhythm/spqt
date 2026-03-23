package qtw

import qt "github.com/mappu/miqt/qt6"

// MenuBuilder constructs a QMenu via chaining.
type MenuBuilder struct{ menu *qt.QMenu }

// Menu creates a new menu with the given title.
func Menu(title string) *MenuBuilder {
	return &MenuBuilder{menu: qt.NewQMenu3(title)}
}

// Action adds a menu item with a callback.
func (b *MenuBuilder) Action(text string, fn func()) *MenuBuilder {
	act := qt.NewQAction2(text)
	act.OnTriggered(fn)
	b.menu.AddAction(act)
	return b
}

// ActionIcon adds a menu item with an icon and callback.
func (b *MenuBuilder) ActionIcon(icon *qt.QIcon, text string, fn func()) *MenuBuilder {
	act := qt.NewQAction3(icon, text)
	act.OnTriggered(fn)
	b.menu.AddAction(act)
	return b
}

// Separator adds a separator line.
func (b *MenuBuilder) Separator() *MenuBuilder {
	b.menu.AddSeparator()
	return b
}

// Sub adds a submenu.
func (b *MenuBuilder) Sub(sub *MenuBuilder) *MenuBuilder {
	b.menu.AddMenu(sub.menu)
	return b
}

func (b *MenuBuilder) Build() *qt.QMenu { return b.menu }

// MenuBar creates a menu bar and adds the given menus.
func MenuBar(menus ...*MenuBuilder) *qt.QMenuBar {
	mb := qt.NewQMenuBar2()
	for _, m := range menus {
		mb.AddMenu(m.menu)
	}
	return mb
}
