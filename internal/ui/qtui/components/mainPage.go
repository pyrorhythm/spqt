package components

import (
	"context"

	qt "github.com/mappu/miqt/qt6"

	"github.com/pyrorhythm/spqt/internal/ui/qtui/components/mpstates"
	"github.com/pyrorhythm/spqt/internal/vm"
	"github.com/pyrorhythm/spqt/pkg/qtw"
)

func BuildMainPage(ctx context.Context, shell *vm.Shell) *qt.QWidget {
	pages := qtw.NewPages()

	builders := map[vm.SidebarState]func() *qt.QWidget{
		vm.SbHome:        mpstates.BuildHome,
		vm.SbSearch:      mpstates.BuildSearch,
		vm.SbLikedTracks: mpstates.BuildLikedTracks(ctx, shell),
	}

	built := map[vm.SidebarState]bool{}

	show := func(state vm.SidebarState) {
		name := state.String()
		if !built[state] {
			if builder, ok := builders[state]; ok {
				pages.Page(name, builder())
				built[state] = true
			}
		}
		pages.Show(name)
	}

	// Build the initial page
	show(shell.CurCtx.Get())

	shell.CurCtx.OnChange(show)

	return pages.Widget()
}
