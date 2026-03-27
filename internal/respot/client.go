package respot

import (
	"context"
	"fmt"
	"strings"

	librespot "github.com/devgianlu/go-librespot"
	"github.com/devgianlu/go-librespot/player"
	connectpb "github.com/devgianlu/go-librespot/proto/spotify/connectstate"
	extmetadatapb "github.com/devgianlu/go-librespot/proto/spotify/extendedmetadata"
	metadatapb "github.com/devgianlu/go-librespot/proto/spotify/metadata"
	"github.com/devgianlu/go-librespot/spclient"
	"github.com/pkg/errors"

	"github.com/pyrorhythm/fn"
	"github.com/pyrorhythm/spqt/internal/types"
	"github.com/pyrorhythm/spqt/pkg/log"
)

const enrichBatchSize = 50

var _ types.Client = (*clientImpl)(nil)

type clientImpl struct {
	ctx  context.Context
	sess types.Session
	p    *player.Player
}

func (c *clientImpl) Player() *player.Player { return c.p }
func (c *clientImpl) Close()                 { c.sess.Close() }

func NewClient(ctx context.Context, sess types.Session) types.Client {
	ctx = log.Span(ctx, "respot")
	log.Trace(ctx).Msg("creating client")
	pl, err := newPlayer(sess)
	if err != nil {
		log.Error(ctx).Stack().Err(err).Msg("failed to create new player")
	}
	log.Trace(ctx).Msg("client ready")
	return &clientImpl{ctx: ctx, sess: sess, p: pl}
}

func sanitize(str string) string {
	return strings.Join(strings.Fields(str), "+")
}

func searchURI(str string) string {
	return fmt.Sprintf("spotify:search:%s", sanitize(str))
}

func likedTracksURI(username string, artistID *string) string {
	return fmt.Sprintf(
		"spotify:user:%s:collection%s",
		username, fn.If(artistID != nil, fmt.Sprintf(":artist:%s", *artistID), ""),
	)
}

func firstOrNil[T any](els ...T) *T {
	if len(els) > 0 {
		return &els[0]
	}
	return nil
}

func w[T any](v T, e error) func(string) (T, error) {
	return func(s string) (T, error) {
		return v, errors.Wrap(e, s)
	}
}

func (c *clientImpl) Search(ctx context.Context, query string) (*spclient.ContextResolver, error) {
	return w(c.ResolveContext(ctx, searchURI(query)))("search")
}

func (c *clientImpl) LikedTracks(ctx context.Context, artistID ...string) (*spclient.ContextResolver, error) {
	return w(c.ResolveContext(ctx, likedTracksURI(c.sess.Username(), firstOrNil(artistID...))))("liked tracks")
}

func (c *clientImpl) ResolveContext(ctx context.Context, uri string) (*spclient.ContextResolver, error) {
	log.Ctx(c.ctx).Trace().Str("uri", uri).Msg("resolve context")
	sc, err := c.sess.Spclient().ContextResolve(ctx, uri)
	if err != nil {
		return nil, fmt.Errorf("resolve %s: %w", uri, err)
	}
	return spclient.NewContextResolver(ctx, fromContext(ctx), c.sess.Spclient(), sc)
}

func (c *clientImpl) FetchTrackMetadata(ctx context.Context, uri string) (*metadatapb.Track, error) {
	log.Ctx(c.ctx).Trace().Str("uri", uri).Msg("fetch track metadata")
	id, err := librespot.SpotifyIdFromUri(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid track uri: %w", err)
	}
	var pb metadatapb.Track
	if err := c.sess.Spclient().ExtendedMetadataSimple(ctx, *id, extmetadatapb.ExtensionKind_TRACK_V4, &pb); err != nil {
		return nil, fmt.Errorf("fetch track %s: %w", uri, err)
	}
	return &pb, nil
}

func (c *clientImpl) FetchAlbumMetadata(ctx context.Context, uri string) (*metadatapb.Album, error) {
	log.Ctx(c.ctx).Trace().Str("uri", uri).Msg("fetch album metadata")
	id, err := librespot.SpotifyIdFromUri(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid album uri: %w", err)
	}
	var pb metadatapb.Album
	if err := c.sess.Spclient().ExtendedMetadataSimple(ctx, *id, extmetadatapb.ExtensionKind_ALBUM_V4, &pb); err != nil {
		return nil, fmt.Errorf("fetch album %s: %w", uri, err)
	}
	return &pb, nil
}

func (c *clientImpl) FetchArtistMetadata(ctx context.Context, uri string) (*metadatapb.Artist, error) {
	log.Ctx(c.ctx).Trace().Str("uri", uri).Msg("fetch artist metadata")
	id, err := librespot.SpotifyIdFromUri(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid artist uri: %w", err)
	}
	var pb metadatapb.Artist
	if err := c.sess.Spclient().ExtendedMetadataSimple(ctx, *id, extmetadatapb.ExtensionKind_ARTIST_V4, &pb); err != nil {
		return nil, fmt.Errorf("fetch artist %s: %w", uri, err)
	}
	return &pb, nil
}

// EnrichPage fetches a page of tracks (TRACK_V4 only) and streams results
// in batches via onBatch. Album/artist enrichment is done lazily by callers.
func (c *clientImpl) EnrichPage(
	ctx context.Context,
	cr *spclient.ContextResolver,
	pageIdx int,
	onBatch func([]*metadatapb.Track),
) error {
	log.Trace(ctx).Int("page", pageIdx).Msg("enrich page start")
	ctxTracks, err := cr.Page(ctx, pageIdx)
	if err != nil {
		return fmt.Errorf("page %d: %w", pageIdx, err)
	}
	log.Trace(ctx).Int("tracks", len(ctxTracks)).Msg("page resolved")
	if len(ctxTracks) == 0 {
		return nil
	}

	pts := make([]*connectpb.ProvidedTrack, 0, len(ctxTracks))
	for _, ct := range ctxTracks {
		pts = append(pts, librespot.ContextTrackToProvidedTrack(cr.Type(), ct))
	}

	for start := 0; start < len(pts); start += enrichBatchSize {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		end := min(start+enrichBatchSize, len(pts))

		log.Trace(ctx).Int("start", start).Int("end", end).Msg("fetching batch")
		batch, err := c.fetchTrackBatch(ctx, pts[start:end])
		if err != nil {
			return fmt.Errorf("batch [%d:%d]: %w", start, end, err)
		}
		log.Trace(ctx).Int("count", len(batch)).Msg("batch ready, dispatching")
		onBatch(batch)
	}

	return nil
}

// fetchTrackBatch issues a single batched TRACK_V4 request.
func (c *clientImpl) fetchTrackBatch(
	ctx context.Context,
	pts []*connectpb.ProvidedTrack,
) ([]*metadatapb.Track, error) {

	entities := make([]*extmetadatapb.EntityRequest, 0, len(pts))
	for _, pt := range pts {
		entities = append(entities, &extmetadatapb.EntityRequest{
			EntityUri: pt.Uri,
			Query:     []*extmetadatapb.ExtensionQuery{{ExtensionKind: extmetadatapb.ExtensionKind_TRACK_V4}},
		})
	}

	resp, err := c.sess.Spclient().ExtendedMetadata(ctx, &extmetadatapb.BatchedEntityRequest{
		EntityRequest: entities,
	})
	if err != nil {
		return nil, err
	}

	trackPBs := make(map[string]*metadatapb.Track, len(pts))
	for _, arr := range resp.ExtendedMetadata {
		for _, ext := range arr.ExtensionData {
			if ext.Header.StatusCode != 200 {
				log.Warn(ctx).Str("uri", ext.EntityUri).Int32("status", ext.Header.StatusCode).Msg("track non-200")
				continue
			}
			var pb metadatapb.Track
			if err := ext.ExtensionData.UnmarshalTo(&pb); err != nil {
				log.Warn(ctx).Err(err).Str("uri", ext.EntityUri).Msg("unmarshal track")
				continue
			}
			trackPBs[ext.EntityUri] = &pb
		}
	}

	result := make([]*metadatapb.Track, 0, len(pts))
	for _, pt := range pts {
		if pb, ok := trackPBs[pt.Uri]; ok {
			result = append(result, pb)
		} else {
			log.Warn(ctx).Str("uri", pt.Uri).Msg("track missing from batch")
		}
	}
	return result, nil
}

func (c *clientImpl) FetchLikesEnrich(
	ctx context.Context,
	onBatch func([]*metadatapb.Track),
) error {
	cr, err := c.LikedTracks(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to fetch liked tracks")
	}

	return c.EnrichPage(ctx, cr, 0, onBatch)
}
