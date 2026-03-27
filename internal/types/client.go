package types

import (
	"context"

	"github.com/devgianlu/go-librespot/player"
	metadatapb "github.com/devgianlu/go-librespot/proto/spotify/metadata"
	"github.com/devgianlu/go-librespot/spclient"
)

type Client interface {
	Close()
	Player() *player.Player
	Search(ctx context.Context, query string) (*spclient.ContextResolver, error)
	LikedTracks(ctx context.Context, artistID ...string) (*spclient.ContextResolver, error)
	ResolveContext(ctx context.Context, uri string) (*spclient.ContextResolver, error)
	FetchTrackMetadata(ctx context.Context, uri string) (*metadatapb.Track, error)
	FetchAlbumMetadata(ctx context.Context, uri string) (*metadatapb.Album, error)
	FetchArtistMetadata(ctx context.Context, uri string) (*metadatapb.Artist, error)

	FetchLikesEnrich(ctx context.Context, onBatch func([]*metadatapb.Track)) error

	EnrichPage(ctx context.Context, cr *spclient.ContextResolver, pageIdx int, onBatch func([]*metadatapb.Track)) error
}
