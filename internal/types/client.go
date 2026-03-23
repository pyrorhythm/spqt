package types

import (
	"context"

	"github.com/devgianlu/go-librespot/spclient"
)

type Client interface {
	Close()
	Search(ctx context.Context, query string) (*spclient.ContextResolver, error)
	LikedTracks(ctx context.Context, artistId ...string) (*spclient.ContextResolver, error)
	ResolveContext(ctx context.Context, uri string) (*spclient.ContextResolver, error)
	FetchTrackMetadata(ctx context.Context, uri string) (*Track, error)
	FetchAlbumMetadata(ctx context.Context, uri string) (*Album, error)
	FetchArtistMetadata(ctx context.Context, uri string) (*Artist, error)
	EnrichPage(ctx context.Context, cr *spclient.ContextResolver, index int) ([]EnrichedTrack, error)
}
