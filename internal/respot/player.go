package respot

import (
	"github.com/devgianlu/go-librespot/player"

	"github.com/pyrorhythm/spqt/internal/types"
)

func newPlayer(sess types.Session) (*player.Player, error) {
	return player.NewPlayer(&player.Options{
		Spclient:     sess.Spclient(),
		AudioKey:     sess.AudioKey(),
		Events:       sess.Events(),
		Log:          logger(),
		CountryCode:  new("us"),
		FlacEnabled:  false,
		AudioBackend: "audio-toolbox",
		// AudioDevice:      "1__2",
		MixerControlName: "Master",
		VolumeUpdate:     make(chan float32, 1),
	})
}
