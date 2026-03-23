package types

import (
	"context"
	"net/http"
	"net/url"

	"github.com/devgianlu/go-librespot/ap"
	"github.com/devgianlu/go-librespot/audio"
	"github.com/devgianlu/go-librespot/dealer"
	"github.com/devgianlu/go-librespot/mercury"
	"github.com/devgianlu/go-librespot/player"
	"github.com/devgianlu/go-librespot/spclient"
)

type Session interface {
	Close()
	Username() string
	StoredCredentials() []byte

	Spclient() *spclient.Spclient
	Events() player.EventManager
	AudioKey() *audio.KeyProvider
	Dealer() *dealer.Dealer
	Accesspoint() *ap.Accesspoint
	Mercury() *mercury.Client

	WebApi(ctx context.Context, method string, path string, query url.Values, header http.Header,
		body []byte) (*http.Response, error)
}
