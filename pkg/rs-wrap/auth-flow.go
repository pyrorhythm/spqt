package rs

import (
	"context"
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
	rssess "github.com/devgianlu/go-librespot/session"
	"github.com/devgianlu/go-librespot/spclient"
	"golang.org/x/oauth2"

	"github.com/pyrorhythm/spqt/pkg/log"
)

func aapFn(ctx context.Context, eventChan chan Event) func(*ap.Accesspoint) error {
	return func(ap *ap.Accesspoint) error {
		var (
			code string
			tok  *oauth2.Token
			pkce = oauth2.GenerateVerifier()
		)

		sctx, scl := context.WithCancel(ctx)

		_, resCh, err := rssess.NewOAuth2Server(sctx, FromContext(ctx), 9292)
		if err != nil {
			scl()
			return fmt.Errorf("failed to open oauth2 server: %w", err)
		}

		eventChan <- LinkEvent{oa2Conf.AuthCodeURL("", oauth2.S256ChallengeOption(pkce))}

		code = <-resCh
		scl()

		eventChan <- CodeReceivedEvent{}

		tok, err = oa2Conf.Exchange(ctx, code, oauth2.VerifierOption(pkce))
		if err != nil {
			return fmt.Errorf("failed exchanging oauth2 code: %w", err)
		}

		return ap.ConnectSpotifyToken(ctx, tok.Extra("username").(string), tok.AccessToken)
	}
}

func Authorize(ctx context.Context) <-chan Event {
	events := make(chan Event, 4)

	go func() {
		var (
			session  *Session
			err      error
			deviceId string
			isNew    bool
			creds    *rssess.StoredCredentials
		)

		if deviceId, isNew = getDeviceID(); isNew {
			_ = write(deviceIDFile, []byte(deviceId))

			goto link
		}

		if creds, err = readJSON[rssess.StoredCredentials](credentialsFile); err != nil {
			log.Ctx(ctx).Warn().Msg("couldnt find cached credentials, proceeding with OAuth")
			err = nil

			goto link
		}

		if session, err = sessionInit(ctx, deviceId, func(ap *ap.Accesspoint) error {
			return ap.ConnectStored(ctx, creds.Username, creds.Data)
		}); err != nil {
			log.Ctx(ctx).Warn().Msg("failed to login with cached credentials, proceeding with OAuth")
			err = nil

			goto link
		}

		goto sessOK

	link:
		session, err = sessionInit(ctx, deviceId, aapFn(ctx, events))
		if err != nil {
			events <- FailedEvent{err}
			close(events)

			return
		}

		goto sessOK

	sessOK:
		events <- SessionAuthorizedEvent{session}
		close(events)

		if err = writeJSON(credentialsFile, rssess.StoredCredentials{
			Username: session.Username(),
			Data:     session.StoredCredentials(),
		}); err != nil {
			log.Ctx(ctx).Warn().Err(err).Msg("failed to save credentials")
		}
	}()

	return events
}

func sessionInit(ctx context.Context, deviceId string, authApFn func(*ap.Accesspoint) error) (*Session, error) {
	var err error

	log := FromContext(ctx)

	s := &Session{
		deviceId:   deviceId,
		deviceType: devicespb.DeviceType_COMPUTER,
		client:     &http.Client{Timeout: 30 * time.Second},
	}

	s.clientToken, err = retrieveClientToken(s.client, s.deviceId)
	if err != nil {
		return nil, fmt.Errorf("failed obtaining client token: %w", err)
	}

	log.Debugf("obtained new client token: %s", s.clientToken)

	s.resolver = apresolve.NewApResolver(log, s.client)
	s.login5 = login5.NewLogin5(log, s.client, s.deviceId, s.clientToken)
	apAddr, err := s.resolver.GetAccesspoint(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed getting accesspoint from resolver: %w", err)
	}

	s.ap = ap.NewAccesspoint(log, apAddr, s.deviceId)

	if err := authApFn(s.ap); err != nil {
		return nil, fmt.Errorf("failed authenticating accesspoint with stored credentials: %w", err)
	}

	if err := s.login5.Login(ctx, &credentialspb.StoredCredential{
		Username: s.ap.Username(),
		Data:     s.ap.StoredCredentials(),
	}); err != nil {
		return nil, fmt.Errorf("failed authenticating with login5: %w", err)
	}

	if spAddr, err := s.resolver.GetSpclient(ctx); err != nil {
		return nil, fmt.Errorf("failed getting spclient from resolver: %w", err)
	} else if s.sp, err = spclient.NewSpclient(ctx, log, s.client, spAddr, s.login5.AccessToken(), s.deviceId,
		s.clientToken); err != nil {
		return nil, fmt.Errorf("failed initializing spclient: %w", err)
	}
	dealerAddr, err := s.resolver.GetDealer(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed getting dealer from resolver: %w", err)
	}
	s.dealer = dealer.NewDealer(log, s.client, dealerAddr, s.login5.AccessToken())
	s.hg = mercury.NewClient(log, s.ap)
	s.audioKey = audio.NewAudioKeyProvider(log, s.ap)

	s.events, err = events.Plugin.NewEventManager(log, nil, s.hg, s.sp, s.ap.Username())
	if err != nil {
		return nil, fmt.Errorf("failed initializing event sender: %w", err)
	}

	return s, nil
}
