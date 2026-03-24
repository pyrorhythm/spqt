package respot

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/devgianlu/go-librespot/ap"
	"github.com/devgianlu/go-librespot/apresolve"
	"github.com/devgianlu/go-librespot/audio"
	"github.com/devgianlu/go-librespot/dealer"
	"github.com/devgianlu/go-librespot/events"
	"github.com/devgianlu/go-librespot/login5"
	"github.com/devgianlu/go-librespot/mercury"
	devicespb "github.com/devgianlu/go-librespot/proto/spotify/connectstate/devices"
	credentialspb "github.com/devgianlu/go-librespot/proto/spotify/login5/v3/credentials"
	"github.com/devgianlu/go-librespot/session"
	"github.com/devgianlu/go-librespot/spclient"
	"golang.org/x/oauth2"

	"github.com/pyrorhythm/spqt/internal/types"
	"github.com/pyrorhythm/spqt/pkg/file"
	"github.com/pyrorhythm/spqt/pkg/log"
)

func Authorize(ctx context.Context) <-chan types.AuthEvent {
	ch := make(chan types.AuthEvent, 4)

	go func() {
		var (
			si       *sessionImpl
			err      error
			deviceId string
			isNew    bool
			creds    *session.StoredCredentials
		)

		if deviceId, isNew = getDeviceID(); isNew {
			_ = file.Write(cacheDir, deviceIDFile, []byte(deviceId))
			goto link
		}

		if creds, err = file.ReadJSON[session.StoredCredentials](cacheDir, credentialsFile); err != nil {
			log.Ctx(ctx).Warn().Msg("couldnt find cached credentials, proceeding with OAuth")
			err = nil
			goto link
		}

		if si, err = sessionInit(ctx, deviceId, func(ap *ap.Accesspoint) error {
			return ap.ConnectStored(ctx, creds.Username, creds.Data)
		}); err != nil {
			log.Ctx(ctx).Warn().Msg("failed to login with cached credentials, proceeding with OAuth")
			err = nil
			goto link
		}

		goto sessOK

	link:
		si, err = sessionInit(ctx, deviceId, oauthApFn(ctx, ch))
		if err != nil {
			ch <- types.FailedEvent{Error: err}
			close(ch)
			return
		}

	sessOK:
		ch <- types.SessionAuthorizedEvent{Session: si}
		close(ch)

		if err = file.WriteJSON(cacheDir, credentialsFile, session.StoredCredentials{
			Username: si.Username(),
			Data:     si.StoredCredentials(),
		}); err != nil {
			log.Ctx(ctx).Warn().Err(err).Msg("failed to save credentials")
		}
	}()

	return ch
}

func oauthApFn(ctx context.Context, eventChan chan types.AuthEvent) func(*ap.Accesspoint) error {
	return func(ap *ap.Accesspoint) error {
		var (
			code string
			tok  *oauth2.Token
			pkce = oauth2.GenerateVerifier()
		)

		sctx, scl := context.WithCancel(ctx)

		_, resCh, err := session.NewOAuth2Server(sctx, FromContext(ctx), 9292)
		if err != nil {
			scl()
			return fmt.Errorf("failed to open oauth2 server: %w", err)
		}

		eventChan <- types.LinkEvent{Link: oa2Conf.AuthCodeURL("", oauth2.S256ChallengeOption(pkce))}

		code = <-resCh
		scl()

		eventChan <- types.CodeReceivedEvent{}

		tok, err = oa2Conf.Exchange(ctx, code, oauth2.VerifierOption(pkce))
		if err != nil {
			return fmt.Errorf("failed exchanging oauth2 code: %w", err)
		}

		return ap.ConnectSpotifyToken(ctx, tok.Extra("username").(string), tok.AccessToken)
	}
}

func getDeviceID() (id string, new bool) {
	rawhex, err := file.Read(cacheDir, deviceIDFile)
	if err != nil {
		log.Logger().Warn().Err(err).Msg("failed to read device ID, generating a random one")
	} else {
		return string(rawhex), false
	}

	bytes := make([]byte, 20)
	_, _ = rand.Read(bytes)

	return hex.EncodeToString(bytes), true
}

func sessionInit(ctx context.Context, deviceId string, authApFn func(*ap.Accesspoint) error) (*sessionImpl, error) {
	var err error

	lg := FromContext(ctx)

	s := &sessionImpl{
		deviceId:   deviceId,
		deviceType: devicespb.DeviceType_COMPUTER,
		client:     &http.Client{Timeout: 30 * time.Second},
	}

	s.clientToken, err = retrieveClientToken(s.client, s.deviceId)
	if err != nil {
		return nil, fmt.Errorf("failed obtaining client token: %w", err)
	}

	lg.Debugf("obtained new client token: %s", s.clientToken)

	s.resolver = apresolve.NewApResolver(lg, s.client)
	s.login5 = login5.NewLogin5(lg, s.client, s.deviceId, s.clientToken)
	apAddr, err := s.resolver.GetAccesspoint(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed getting accesspoint from resolver: %w", err)
	}

	s.ap = ap.NewAccesspoint(lg, apAddr, s.deviceId)

	if err := authApFn(s.ap); err != nil {
		return nil, fmt.Errorf("failed authenticating accesspoint: %w", err)
	}

	if err := s.login5.Login(ctx, &credentialspb.StoredCredential{
		Username: s.ap.Username(),
		Data:     s.ap.StoredCredentials(),
	}); err != nil {
		return nil, fmt.Errorf("failed authenticating with login5: %w", err)
	}

	if spAddr, err := s.resolver.GetSpclient(ctx); err != nil {
		return nil, fmt.Errorf("failed getting spclient from resolver: %w", err)
	} else if s.sp, err = spclient.NewSpclient(ctx, lg, s.client, spAddr, s.login5.AccessToken(), s.deviceId,
		s.clientToken); err != nil {
		return nil, fmt.Errorf("failed initializing spclient: %w", err)
	}

	dealerAddr, err := s.resolver.GetDealer(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed getting dealer from resolver: %w", err)
	}
	s.dealer = dealer.NewDealer(lg, s.client, dealerAddr, s.login5.AccessToken())
	s.hg = mercury.NewClient(lg, s.ap)
	s.audioKey = audio.NewAudioKeyProvider(lg, s.ap)

	s.events, err = events.Plugin.NewEventManager(lg, nil, s.hg, s.sp, s.ap.Username())
	if err != nil {
		return nil, fmt.Errorf("failed initializing event sender: %w", err)
	}

	return s, nil
}
