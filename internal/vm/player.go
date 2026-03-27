package vm

import (
	"context"
	"net/http"
	"time"

	respot "github.com/devgianlu/go-librespot"
	"github.com/devgianlu/go-librespot/player"
	metadatapb "github.com/devgianlu/go-librespot/proto/spotify/metadata"

	"github.com/pyrorhythm/spqt/internal/types"
	"github.com/pyrorhythm/spqt/pkg/log"
	"github.com/pyrorhythm/spqt/pkg/reactive"
)

type playerCmd int

const (
	PCPlayPause playerCmd = 2<<iota - 1
	PCStop
	PCNext
	PCPrev
)

type Player struct {
	Current   *reactive.CmpProp[*metadatapb.Track, string]
	IsPlaying *reactive.Prop[bool]
	Progress  *reactive.Prop[int64]
	Volume    *reactive.Prop[uint32]
	CanNext   *reactive.Prop[bool]

	PlayerCmd *reactive.ECommand[playerCmd]

	c types.Client
}

func (p *Player) SeekPos(posMs int64) {
	if p.c != nil {
		p.c.Player().SeekMs(posMs)
	}
}

func (p *Player) Exec(pc playerCmd) func() {
	return func() {
		p.PlayerCmd.Execute(pc)
	}
}

func newPlayer() *Player {
	return &Player{
		Current:   reactive.NewUProp(nil, types.TrackComparator{}),
		IsPlaying: reactive.NewProp(false),
		Progress:  reactive.NewProp[int64](0),
		Volume:    reactive.NewProp[uint32](50),
		CanNext:   reactive.NewProp(true),
		PlayerCmd: reactive.NewECommand[playerCmd](),
	}
}

func (p *Player) Client() types.Client {
	return p.c
}

func (p *Player) eventListener(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case e := <-p.c.Player().Receive():
			log.Trace(ctx).Int("type", int(e.Type)).Msg("event")
			switch e.Type {
			case player.EventTypePlay, player.EventTypeResume:
				p.IsPlaying.Set(true)
			case player.EventTypeStop, player.EventTypePause:
				p.IsPlaying.Set(false)
			case player.EventTypeNotPlaying:
				p.IsPlaying.Set(false)
				p.Current.Set(nil)
			}
		}
	}
}

func (p *Player) BindClient(ctx context.Context, c types.Client) {
	ctx = log.Span(ctx, "player")

	p.c = c
	log.Trace(ctx).Msg("registering commands")

	p.PlayerCmd.Register(PCPlayPause, reactive.NewCommand(func() {
		if p.IsPlaying.Get() {
			log.Trace(ctx).Msg("cmd: pause")
			c.Player().Pause()
			return
		}
		log.Trace(ctx).Msg("cmd: play")
		c.Player().Play()
	}, nil))

	p.PlayerCmd.Register(PCStop, reactive.NewCommand(func() {
		c.Player().Stop()
	}, nil))

	p.PlayerCmd.Register(PCNext, reactive.NewCommand(func() {
		// TODO: integrate with track queue
	}, nil))

	p.PlayerCmd.Register(PCPrev, reactive.NewCommand(func() {
		// TODO: integrate with track queue
	}, nil))

	p.Progress.Poll(ctx, time.Second, c.Player().PositionMs)

	p.Current.OnChange(func(pb *metadatapb.Track) { p.setStream(ctx, pb) })

	log.Trace(ctx).Msg("starting event listener")
	go p.eventListener(ctx)
}

func (p *Player) setStream(ctx context.Context, pb *metadatapb.Track) {
	if pb == nil {
		return
	}
	sid := respot.SpotifyIdFromGid(respot.SpotifyIdTypeTrack, pb.Gid)
	log.Trace(ctx).Str("track", pb.GetName()).Str("sid", sid.Uri()).Msg("new stream")
	str, err := p.c.Player().NewStream(ctx, &http.Client{}, sid, 320, 0)
	if err != nil {
		log.Error(ctx).Stack().Err(err).Msg("new stream failed")
		return
	}
	log.Trace(ctx).Msg("setting primary stream")
	err = p.c.Player().SetPrimaryStream(str.Source, false, false)
	if err != nil {
		log.Error(ctx).Stack().Err(err).Msg("set primary stream failed")
	}
}
