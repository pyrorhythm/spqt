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

// EnrichPage fetches a page from the resolver, then issues batched metadata
// requests. Phase 1: fetch all track metadata. Phase 2: extract album/artist
// URIs from the track protos and batch-fetch those (ProvidedTrack metadata
// from context-resolve often lacks album_uri/artist_uri).
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

	lg := log.Ctx(ctx)

	// Convert to ProvidedTracks, collect track URIs.
	pts := make([]*connectpb.ProvidedTrack, 0, len(ctxTracks))
	for _, ct := range ctxTracks {
		pts = append(pts, librespot.ContextTrackToProvidedTrack(cr.Type(), ct))
	}

	// --- Phase 1: batch-fetch all tracks ---
	trackEntities := make([]*extmetadatapb.EntityRequest, 0, len(pts))
	for _, pt := range pts {
		trackEntities = append(trackEntities, &extmetadatapb.EntityRequest{
			EntityUri: pt.Uri,
			Query:     []*extmetadatapb.ExtensionQuery{{ExtensionKind: extmetadatapb.ExtensionKind_TRACK_V4}},
		})
	}

	trackResp, err := c.sess.Spclient().ExtendedMetadata(ctx, &extmetadatapb.BatchedEntityRequest{
		EntityRequest: trackEntities,
	})
	if err != nil {
		return nil, fmt.Errorf("batch track metadata request failed: %w", err)
	}

	trackPBs := make(map[string]*metadatapb.Track, len(pts))
	for _, arr := range trackResp.ExtendedMetadata {
		for _, ext := range arr.ExtensionData {
			if ext.Header.StatusCode != 200 {
				lg.Warn().Str("uri", ext.EntityUri).Int32("status", ext.Header.StatusCode).Msg("track metadata non-200")
				continue
			}
			var pb metadatapb.Track
			if err := ext.ExtensionData.UnmarshalTo(&pb); err != nil {
				lg.Warn().Err(err).Str("uri", ext.EntityUri).Msg("failed to unmarshal track")
				continue
			}
			trackPBs[ext.EntityUri] = &pb
		}
	}

	// --- Phase 2: collect album/artist URIs from track protos, batch-fetch ---
	albumURISet := make(map[string]struct{})
	artistURISet := make(map[string]struct{})

	for _, pb := range trackPBs {
		if pb.Album != nil && len(pb.Album.Gid) > 0 {
			albumURISet["spotify:album:"+librespot.GidToBase62(pb.Album.Gid)] = struct{}{}
		}
		for _, a := range pb.Artist {
			if len(a.Gid) > 0 {
				artistURISet["spotify:artist:"+librespot.GidToBase62(a.Gid)] = struct{}{}
			}
		}
	}

	albums := make(map[string]*metadatapb.Album)
	artists := make(map[string]*metadatapb.Artist)

	if len(albumURISet)+len(artistURISet) > 0 {
		var auxEntities []*extmetadatapb.EntityRequest
		for uri := range albumURISet {
			auxEntities = append(auxEntities, &extmetadatapb.EntityRequest{
				EntityUri: uri,
				Query:     []*extmetadatapb.ExtensionQuery{{ExtensionKind: extmetadatapb.ExtensionKind_ALBUM_V4}},
			})
		}
		for uri := range artistURISet {
			auxEntities = append(auxEntities, &extmetadatapb.EntityRequest{
				EntityUri: uri,
				Query:     []*extmetadatapb.ExtensionQuery{{ExtensionKind: extmetadatapb.ExtensionKind_ARTIST_V4}},
			})
		}

		auxResp, err := c.sess.Spclient().ExtendedMetadata(ctx, &extmetadatapb.BatchedEntityRequest{
			EntityRequest: auxEntities,
		})
		if err != nil {
			lg.Warn().Err(err).Msg("batch album/artist metadata request failed, continuing without enrichment")
		} else {
			for _, arr := range auxResp.ExtendedMetadata {
				for _, ext := range arr.ExtensionData {
					if ext.Header.StatusCode != 200 {
						lg.Warn().Str("uri", ext.EntityUri).Int32("status", ext.Header.StatusCode).Msg("aux metadata non-200")
						continue
					}
					switch arr.ExtensionKind {
					case extmetadatapb.ExtensionKind_ALBUM_V4:
						var pb metadatapb.Album
						if err := ext.ExtensionData.UnmarshalTo(&pb); err == nil {
							albums[ext.EntityUri] = &pb
						}
					case extmetadatapb.ExtensionKind_ARTIST_V4:
						var pb metadatapb.Artist
						if err := ext.ExtensionData.UnmarshalTo(&pb); err == nil {
							artists[ext.EntityUri] = &pb
						}
					}
				}
			}
		}
	}

	// --- Assemble enriched tracks ---
	result := make([]types.EnrichedTrack, 0, len(pts))

	for _, pt := range pts {
		pbTrack, ok := trackPBs[pt.Uri]
		if !ok {
			lg.Warn().Str("uri", pt.Uri).Msg("track missing from batch response, skipping")
			continue
		}

		var t types.Track
		t.FromProto(pbTrack)
		et := types.EnrichedTrack{Track: &t}

		// Album
		if pbAlbum, ok := albums[t.Album.URI]; ok {
			var a types.Album
			a.FromProto(pbAlbum)
			et.FullAlbum = &a
		}

		// Artists
		for _, ar := range t.Artists {
			if pbArt, ok := artists[ar.URI]; ok {
				var a types.Artist
				a.FromProto(pbArt)
				et.FullArtists = append(et.FullArtists, &a)
			}
		}

		result = append(result, et)
	}

	return result, nil
}
