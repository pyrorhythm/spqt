package vm

type NavState string

const (
	NavHome        NavState = "Home"
	NavSearch      NavState = "Search"
	NavLikedTracks NavState = "Liked Songs"
)

func (s NavState) String() string {
	return (string)(s)
}

func NavStateFrom(s string) NavState {
	switch ss := NavState(s); ss {
	case NavHome, NavSearch, NavLikedTracks:
		return ss
	default:
		return NavHome
	}
}
