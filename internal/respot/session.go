package respot

import (
	"context"
	"net/http"
	"net/url"

	"github.com/devgianlu/go-librespot/ap"
	"github.com/devgianlu/go-librespot/apresolve"
	"github.com/devgianlu/go-librespot/audio"
	"github.com/devgianlu/go-librespot/dealer"
	"github.com/devgianlu/go-librespot/login5"
	"github.com/devgianlu/go-librespot/mercury"
	"github.com/devgianlu/go-librespot/player"
	devicespb "github.com/devgianlu/go-librespot/proto/spotify/connectstate/devices"
	"github.com/devgianlu/go-librespot/spclient"
)

type sessionImpl struct {
	deviceType  devicespb.DeviceType
	deviceId    string
	clientToken string

	client *http.Client

	resolver *apresolve.ApResolver
	login5   *login5.Login5

	ap       *ap.Accesspoint
	hg       *mercury.Client
	sp       *spclient.Spclient
	dealer   *dealer.Dealer
	audioKey *audio.KeyProvider
	events   player.EventManager
}

func (s *sessionImpl) Close() {
	s.events.Close()
	s.audioKey.Close()
	s.hg.Close()
	s.dealer.Close()
	s.ap.Close()
}

func (s *sessionImpl) Username() string {
	return s.ap.Username()
}

func (s *sessionImpl) StoredCredentials() []byte {
	return s.ap.StoredCredentials()
}

func (s *sessionImpl) Spclient() *spclient.Spclient {
	return s.sp
}

func (s *sessionImpl) Events() player.EventManager {
	return s.events
}

func (s *sessionImpl) AudioKey() *audio.KeyProvider {
	return s.audioKey
}

func (s *sessionImpl) Dealer() *dealer.Dealer {
	return s.dealer
}

func (s *sessionImpl) Accesspoint() *ap.Accesspoint {
	return s.ap
}

func (s *sessionImpl) Mercury() *mercury.Client {
	return s.hg
}

func (s *sessionImpl) WebApi(ctx context.Context, method string, path string, query url.Values, header http.Header,
	body []byte) (*http.Response, error) {
	return s.sp.WebApiRequest(ctx, method, path, query, header, body)
}
