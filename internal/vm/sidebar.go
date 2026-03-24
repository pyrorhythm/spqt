package vm

type SidebarState string

const (
	SbHome        SidebarState = "Home"
	SbSearch      SidebarState = "Search"
	SbLikedTracks SidebarState = "Liked Songs"
)

func (s SidebarState) String() string {
	return (string)(s)
}

func SidebarStateFrom(s string) SidebarState {
	switch ss := SidebarState(s); ss {
	case SbHome, SbSearch, SbLikedTracks:
		return ss
	default:
		return SbHome
	}
}
