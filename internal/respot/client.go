package respot

import (
	"context"
	"fmt"
	"strings"

	librespot "github.com/devgianlu/go-librespot"
	connectpb "github.com/devgianlu/go-librespot/proto/spotify/connectstate"
	extmetadatapb "github.com/devgianlu/go-librespot/proto/spotify/extendedmetadata"
	metadatapb "github.com/devgianlu/go-librespot/proto/spotify/metadata"
	"github.com/devgianlu/go-librespot/spclient"

	"github.com/pyrorhythm/spqt/internal/types"
	"github.com/pyrorhythm/spqt/pkg/log"
)

type clientImpl struct {
	sess types.Session
}

func (c *clientImpl) Close() {
	c.sess.Close()
}

func NewClient(sess types.Session) *clientImpl {
	return &clientImpl{sess: sess}
}

func (c *clientImpl) resolve(ctx context.Context, sc *connectpb.Context) (*spclient.ContextResolver, error) {
	return spclient.NewContextResolver(ctx, FromContext(ctx), c.sess.Spclient(), sc)
}

// Request the context for an uri
//
// All [SpotifyId] uris are supported in addition to the following special uris:
// - liked songs:
//   - all: `spotify:user:<user_id>:collection`
//   - of artist: `spotify:user:<user_id>:collection:artist:<artist_id>`
// - search: `spotify:search:<search+query>` (whitespaces are replaced with `+`)
//
// ## Query params found in the wild:
// - include_video=true
//
// ## Known results of uri types:
// - uris of type `track`
//   - returns a single page with a single track
//   - when requesting a single track with a query in the request, the returned track uri
//     **will** contain the query
// - uris of type `artist`
//   - returns 2 pages with tracks: 10 most popular tracks and latest/popular album
//   - remaining pages are artist albums sorted by popularity (only provided as page_url)
// - uris of type `search`
//   - is massively influenced by the provided query
//   - the query result shown by the search expects no query at all
//   - uri looks like `spotify:search:never+gonna`

func (c *clientImpl) Search(ctx context.Context, query string) (*spclient.ContextResolver, error) {
	encq := strings.Join(strings.Fields(query), "+")
	sc, err := c.sess.Spclient().
		ContextResolve(ctx, fmt.Sprintf("spotify:search:%s", encq))
	if err != nil {
		return nil, fmt.Errorf("failed to perform (%s) search: %w", query, err)
	}

	cr, err := c.resolve(ctx, sc)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve search context: %w", err)
	}

	return cr, nil
}

func (c *clientImpl) LikedTracks(ctx context.Context, artistId ...string) (*spclient.ContextResolver, error) {
	uri := fmt.Sprintf("spotify:user:%s:%s", c.sess.Username(), "collection")
	if len(artistId) > 0 {
		uri += fmt.Sprintf(":artist:%s", artistId[0])
	}

	sc, err := c.sess.Spclient().
		ContextResolve(ctx, uri)
	if err != nil {
		return nil, fmt.Errorf("failed to get users' liked tracks: %w", err)
	}

	cr, err := c.resolve(ctx, sc)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve liked tracks context: %w", err)
	}

	return cr, nil
}

func (c *clientImpl) ResolveContext(ctx context.Context, uri string) (*spclient.ContextResolver, error) {
	sc, err := c.sess.Spclient().ContextResolve(ctx, uri)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve context %s: %w", uri, err)
	}

	return c.resolve(ctx, sc)
}

func (c *clientImpl) FetchTrackMetadata(ctx context.Context, uri string) (*types.Track, error) {
	id, err := librespot.SpotifyIdFromUri(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid track uri: %w", err)
	}

	var pb metadatapb.Track
	if err := c.sess.Spclient().ExtendedMetadataSimple(
		ctx, *id, extmetadatapb.ExtensionKind_TRACK_V4, &pb,
	); err != nil {
		return nil, fmt.Errorf("failed fetching track metadata for %s: %w", uri, err)
	}

	var t types.Track
	t.FromProto(&pb)
	return &t, nil
}

func (c *clientImpl) FetchAlbumMetadata(ctx context.Context, uri string) (*types.Album, error) {
	id, err := librespot.SpotifyIdFromUri(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid album uri: %w", err)
	}

	var pb metadatapb.Album
	if err := c.sess.Spclient().ExtendedMetadataSimple(
		ctx, *id, extmetadatapb.ExtensionKind_ALBUM_V4, &pb,
	); err != nil {
		return nil, fmt.Errorf("failed fetching album metadata for %s: %w", uri, err)
	}

	var a types.Album
	a.FromProto(&pb)
	return &a, nil
}

func (c *clientImpl) FetchArtistMetadata(ctx context.Context, uri string) (*types.Artist, error) {
	id, err := librespot.SpotifyIdFromUri(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid artist uri: %w", err)
	}

	var pb metadatapb.Artist
	if err := c.sess.Spclient().ExtendedMetadataSimple(
		ctx, *id, extmetadatapb.ExtensionKind_ARTIST_V4, &pb,
	); err != nil {
		return nil, fmt.Errorf("failed fetching artist metadata for %s: %w", uri, err)
	}

	var a types.Artist
	a.FromProto(&pb)
	return &a, nil
}

// EnrichPage fetches a page from the resolver, then issues a single batched
// ExtendedMetadata request for all tracks, albums, and artists on that page.
// Returns enriched tracks with full metadata attached.
func (c *clientImpl) EnrichPage(
	ctx context.Context,
	cr *spclient.ContextResolver,
	pageIdx int,
) ([]types.EnrichedTrack, error) {
	ctxTracks, err := cr.Page(ctx, pageIdx)
	if err != nil {
		return nil, fmt.Errorf("failed fetching page %d: %w", pageIdx, err)
	}
	if len(ctxTracks) == 0 {
		return nil, nil
	}

	// Convert to ProvidedTracks and collect all URIs we need metadata for.
	pts := make([]*connectpb.ProvidedTrack, 0, len(ctxTracks))
	trackURIs := make([]string, 0, len(ctxTracks))
	albumURISet := make(map[string]struct{})
	artistURISet := make(map[string]struct{})

	for _, ct := range ctxTracks {
		pt := librespot.ContextTrackToProvidedTrack(cr.Type(), ct)
		pts = append(pts, pt)
		trackURIs = append(trackURIs, pt.Uri)

		if pt.AlbumUri != "" {
			albumURISet[pt.AlbumUri] = struct{}{}
		}
		if pt.ArtistUri != "" {
			artistURISet[pt.ArtistUri] = struct{}{}
		}
	}

	// Build one batch request with all entities.
	var entities []*extmetadatapb.EntityRequest
	for _, uri := range trackURIs {
		entities = append(entities, &extmetadatapb.EntityRequest{
			EntityUri: uri,
			Query:     []*extmetadatapb.ExtensionQuery{{ExtensionKind: extmetadatapb.ExtensionKind_TRACK_V4}},
		})
	}
	for uri := range albumURISet {
		entities = append(entities, &extmetadatapb.EntityRequest{
			EntityUri: uri,
			Query:     []*extmetadatapb.ExtensionQuery{{ExtensionKind: extmetadatapb.ExtensionKind_ALBUM_V4}},
		})
	}
	for uri := range artistURISet {
		entities = append(entities, &extmetadatapb.EntityRequest{
			EntityUri: uri,
			Query:     []*extmetadatapb.ExtensionQuery{{ExtensionKind: extmetadatapb.ExtensionKind_ARTIST_V4}},
		})
	}

	resp, err := c.sess.Spclient().ExtendedMetadata(ctx, &extmetadatapb.BatchedEntityRequest{
		EntityRequest: entities,
	})
	if err != nil {
		return nil, fmt.Errorf("batch metadata request failed: %w", err)
	}

	// Index responses by (kind, uri) for O(1) lookup.
	tracks := make(map[string]*metadatapb.Track)
	albums := make(map[string]*metadatapb.Album)
	artists := make(map[string]*metadatapb.Artist)

	lg := log.Ctx(ctx)

	for _, arr := range resp.ExtendedMetadata {
		for _, ext := range arr.ExtensionData {
			if ext.Header.StatusCode != 200 {
				lg.Warn().
					Str("uri", ext.EntityUri).
					Int32("status", ext.Header.StatusCode).
					Msg("metadata entity returned non-200")
				continue
			}

			switch arr.ExtensionKind {
			case extmetadatapb.ExtensionKind_TRACK_V4:
				var pb metadatapb.Track
				if err := ext.ExtensionData.UnmarshalTo(&pb); err != nil {
					lg.Warn().Err(err).Str("uri", ext.EntityUri).Msg("failed to unmarshal track")
					continue
				}
				tracks[ext.EntityUri] = &pb

			case extmetadatapb.ExtensionKind_ALBUM_V4:
				var pb metadatapb.Album
				if err := ext.ExtensionData.UnmarshalTo(&pb); err != nil {
					lg.Warn().Err(err).Str("uri", ext.EntityUri).Msg("failed to unmarshal album")
					continue
				}
				albums[ext.EntityUri] = &pb

			case extmetadatapb.ExtensionKind_ARTIST_V4:
				var pb metadatapb.Artist
				if err := ext.ExtensionData.UnmarshalTo(&pb); err != nil {
					lg.Warn().Err(err).Str("uri", ext.EntityUri).Msg("failed to unmarshal artist")
					continue
				}
				artists[ext.EntityUri] = &pb
			}
		}
	}

	// Assemble enriched tracks. Tracks that came back from the batch
	// may reference album/artist URIs we didn't know about from the
	// ProvidedTrack metadata — those are filled from the Track proto itself.
	result := make([]types.EnrichedTrack, 0, len(pts))

	for _, pt := range pts {
		pbTrack, ok := tracks[pt.Uri]
		if !ok {
			lg.Warn().Str("uri", pt.Uri).Msg("track missing from batch response, skipping")
			continue
		}

		var t types.Track
		t.FromProto(pbTrack)
		et := types.EnrichedTrack{Track: &t}

		// Album — prefer the URI from ProvidedTrack, fall back to parsed track
		albumURI := pt.AlbumUri
		if albumURI == "" {
			albumURI = t.Album.URI
		}
		if pbAlbum, ok := albums[albumURI]; ok {
			var a types.Album
			a.FromProto(pbAlbum)
			et.FullAlbum = &a
		}

		// Artists
		seen := make(map[string]struct{})
		var artURIs []string
		if pt.ArtistUri != "" {
			seen[pt.ArtistUri] = struct{}{}
			artURIs = append(artURIs, pt.ArtistUri)
		}
		for _, ar := range t.Artists {
			if ar.URI != "" {
				if _, dup := seen[ar.URI]; !dup {
					seen[ar.URI] = struct{}{}
					artURIs = append(artURIs, ar.URI)
				}
			}
		}
		for _, uri := range artURIs {
			if pbArt, ok := artists[uri]; ok {
				var a types.Artist
				a.FromProto(pbArt)
				et.FullArtists = append(et.FullArtists, &a)
			}
		}

		result = append(result, et)
	}

	return result, nil
}
