package respot

import (
	librespot "github.com/devgianlu/go-librespot"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/spotify"
)

const (
	cacheDir        = "./.spqt"
	deviceIDFile    = "deviceID"
	credentialsFile = "spotifyCredentials"
)

var oa2Conf = &oauth2.Config{
	ClientID:    librespot.ClientIdHex,
	RedirectURL: "http://127.0.0.1:9292/login",
	Scopes: []string{
		"app-remote-control",
		"playlist-modify",
		"playlist-modify-private",
		"playlist-modify-public",
		"playlist-read",
		"playlist-read-collaborative",
		"playlist-read-private",
		"streaming",
		"ugc-image-upload",
		"user-follow-modify",
		"user-follow-read",
		"user-library-modify",
		"user-library-read",
		"user-modify",
		"user-modify-playback-state",
		"user-modify-private",
		"user-personalized",
		"user-read-birthdate",
		"user-read-currently-playing",
		"user-read-email",
		"user-read-play-history",
		"user-read-playback-position",
		"user-read-playback-state",
		"user-read-private",
		"user-read-recently-played",
		"user-top-read",
	},
	Endpoint: spotify.Endpoint,
}
