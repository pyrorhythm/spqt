package vm

import (
	"context"
	"math/big"
	"net/http"
	"strings"

	respot "github.com/devgianlu/go-librespot"

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
	Current   *reactive.CmpProp[types.EnrichedTrack, string]
	IsPlaying *reactive.Prop[bool]
	Progress  *reactive.Prop[int32]
	Volume    *reactive.Prop[uint32]
	CanNext   *reactive.Prop[bool]

	PlayerCmd *reactive.ECommand[playerCmd]

	c types.Client
}

func (p *Player) Exec(pc playerCmd) func() {
	return func() {
		p.PlayerCmd.Execute(pc)
	}
}

func newPlayer() *Player {
	return &Player{
		Current:   reactive.NewUProp(types.EnrichedTrack{}, types.TrackComparator{}),
		IsPlaying: reactive.NewProp(false),
		Progress:  reactive.NewProp[int32](0),
		Volume:    reactive.NewProp[uint32](50),
		CanNext:   reactive.NewProp(true),
		PlayerCmd: reactive.NewECommand[playerCmd](),
	}
}

func (p *Player) BindClient(ctx context.Context, c types.Client) {
	p.c = c

	p.PlayerCmd.Register(PCPlayPause, reactive.NewCommand(func() {
		if p.IsPlaying.Get() {
			c.Player().Pause()
			return
		}

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

	p.Current.OnChange(func(et types.EnrichedTrack) {
		parts := strings.Split(et.URI, ":")
		var i big.Int
		_, _ = i.SetString(parts[2], 62)

		sid := respot.SpotifyIdFromGid(respot.SpotifyIdTypeTrack, i.FillBytes(make([]byte, 16)))

		// log.Logger().Debug().Any("sid", sid).Bytes("id", sid.Id()).Send()

		str, err := p.c.Player().NewStream(ctx, &http.Client{}, sid, 320, 0)
		if err != nil {
			log.Logger().Error().Stack().Err(err).Send()
			return
		}

		err = p.c.Player().SetPrimaryStream(str.Source, false, false)
		if err != nil {
			log.Logger().Error().Stack().Err(err).Send()
		}
	})
}

const BASE62 = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func encode(input []byte) string {
	if len(input) == 0 {
		return ""
	}

	value := new(big.Int).SetBytes(input)
	var result strings.Builder

	for value.Cmp(big.NewInt(0)) > 0 {
		mod := new(big.Int)
		value.DivMod(value, big.NewInt(62), mod)
		result.WriteByte(BASE62[mod.Int64()])
	}

	for _, b := range input {
		if b == 0 {
			result.WriteByte(BASE62[0])
		} else {
			break
		}
	}

	encoded := result.String()
	runes := []rune(encoded)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func decode(input string) []byte {
	if len(input) == 0 {
		return []byte{}
	}

	value := big.NewInt(0)
	for _, c := range input {
		value.Mul(value, big.NewInt(62))
		value.Add(value, big.NewInt(int64(strings.IndexRune(BASE62, c))))
	}

	zb62 := []rune(BASE62)[0]

	decoded := value.Bytes()
	leadingZeroes := 0
	for _, c := range input {
		if c == zb62 {
			leadingZeroes++
		} else {
			break
		}
	}

	result := make([]byte, leadingZeroes+len(decoded))
	copy(result[leadingZeroes:], decoded)
	return result
}
