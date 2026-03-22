package rs

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"

	librespot "github.com/devgianlu/go-librespot"
	"github.com/devgianlu/go-librespot/ap"
	"github.com/devgianlu/go-librespot/apresolve"
	"github.com/devgianlu/go-librespot/audio"
	"github.com/devgianlu/go-librespot/dealer"
	"github.com/devgianlu/go-librespot/login5"
	"github.com/devgianlu/go-librespot/mercury"
	"github.com/devgianlu/go-librespot/player"
	devicespb "github.com/devgianlu/go-librespot/proto/spotify/connectstate/devices"
	"github.com/devgianlu/go-librespot/spclient"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/spotify"

	"github.com/pyrorhythm/spqt/pkg/log"
)

type Session struct {
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

func (s *Session) Close() {
	s.events.Close()
	s.audioKey.Close()
	s.hg.Close()
	s.dealer.Close()
	s.ap.Close()
}

func (s *Session) Username() string {
	return s.ap.Username()
}

func (s *Session) StoredCredentials() []byte {
	return s.ap.StoredCredentials()
}

func (s *Session) Spclient() *spclient.Spclient {
	return s.sp
}

func (s *Session) Events() player.EventManager {
	return s.events
}

func (s *Session) AudioKey() *audio.KeyProvider {
	return s.audioKey
}

func (s *Session) Dealer() *dealer.Dealer {
	return s.dealer
}

func (s *Session) Accesspoint() *ap.Accesspoint {
	return s.ap
}

func (s *Session) Mercury() *mercury.Client {
	return s.hg
}

func (s *Session) WebApi(ctx context.Context, method string, path string, query url.Values, header http.Header,
	body []byte) (*http.Response, error) {
	return s.sp.WebApiRequest(ctx, method, path, query, header, body)
}

type Event interface {
	iAmAnEvent()
}

type FailedEvent struct{ Error error }
type SessionAuthorizedEvent struct{ Session *Session }
type LinkEvent struct{ Link string }
type CodeReceivedEvent struct{}

func (s SessionAuthorizedEvent) iAmAnEvent() {}
func (l LinkEvent) iAmAnEvent()              {}
func (f FailedEvent) iAmAnEvent()            {}
func (l CodeReceivedEvent) iAmAnEvent()      {}

const (
	cacheDir        = "./cache"
	deviceIDFile    = "deviceID"
	credentialsFile = "spotifyCredentials"
)

var oa2Conf = &oauth2.Config{
	ClientID:    librespot.ClientIdHex,
	RedirectURL: fmt.Sprintf("http://127.0.0.1:9292/login"),
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

func getDeviceID() (id string, new bool) {
	rawhex, err := read(deviceIDFile)
	if err != nil {
		log.Logger().Warn().Err(err).Msg("failed to read device ID, generating a random one")
	} else {
		return string(rawhex), false
	}

	bytes := make([]byte, 20)
	_, _ = rand.Read(bytes)

	return hex.EncodeToString(bytes), true
}

func dirNe(dir string) {
	fi, err := os.Stat(dir)
	if fi == nil || err != nil {
		os.MkdirAll(dir, 0755)
	}

}

func read(file string) ([]byte, error) {
	dirNe(cacheDir)
	f, err := os.OpenFile(
		path.Join(cacheDir, file),
		os.O_RDONLY,
		0o700)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()

	return io.ReadAll(f)
}

func readJSON[T any](file string) (*T, error) {
	data, err := read(file)
	if err != nil {
		return nil, err
	}

	var t T
	if err = json.Unmarshal(data, &t); err != nil {
		return nil, err
	}

	return &t, nil
}

func write(file string, pl []byte) error {
	f, err := os.OpenFile(
		path.Join(cacheDir, file),
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		0o700)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	_, err = f.Write(pl)

	return err
}

func writeJSON(file string, v any) error {
	pl, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return write(file, pl)
}
