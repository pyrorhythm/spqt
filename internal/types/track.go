package types

import (
	"encoding/hex"
	"strings"
	"time"

	librespot "github.com/devgianlu/go-librespot"
	metadatapb "github.com/devgianlu/go-librespot/proto/spotify/metadata"
	playlist4pb "github.com/devgianlu/go-librespot/proto/spotify/playlist4"

	"github.com/pyrorhythm/spqt/pkg/reactive"
)

// Image is a CDN image extracted from proto Image messages.
type Image struct {
	URL    string
	Width  int32
	Height int32
}

// ImagesFromProto converts proto Image messages to CDN-url Images.
func ImagesFromProto(pbs []*metadatapb.Image) []Image {
	if len(pbs) == 0 {
		return nil
	}
	imgs := make([]Image, 0, len(pbs))
	for _, img := range pbs {
		imgs = append(imgs, Image{
			URL:    "https://i.scdn.co/image/" + hex.EncodeToString(img.GetFileId()),
			Width:  img.GetWidth(),
			Height: img.GetHeight(),
		})
	}
	return imgs
}

// TrackURI extracts the URI from a track proto.
func TrackURI(pb *metadatapb.Track) string {
	if pb == nil {
		return ""
	}
	if pb.GetCanonicalUri() != "" {
		return pb.GetCanonicalUri()
	}
	if len(pb.Gid) > 0 {
		return librespot.SpotifyIdFromGid(librespot.SpotifyIdTypeTrack, pb.Gid).Uri()
	}
	return ""
}

func AlbumURI(pb *metadatapb.Album) string {
	if pb == nil || len(pb.Gid) == 0 {
		return ""
	}
	return "spotify:album:" + librespot.GidToBase62(pb.Gid)
}

func ArtistURI(pb *metadatapb.Artist) string {
	if pb == nil || len(pb.Gid) == 0 {
		return ""
	}
	return "spotify:artist:" + librespot.GidToBase62(pb.Gid)
}

func AlbumCovers(pb *metadatapb.Album) []Image {
	if pb == nil {
		return nil
	}
	imgs := ImagesFromProto(pb.Cover)
	if len(imgs) == 0 && pb.CoverGroup != nil {
		imgs = ImagesFromProto(pb.CoverGroup.GetImage())
	}
	return imgs
}

func ArtistNames(pb *metadatapb.Track) string {
	if pb == nil {
		return "<*metadatapb.Track=nil>"
	}
	names := make([]string, 0, len(pb.GetArtist()))
	for _, a := range pb.GetArtist() {
		names = append(names, a.GetName())
	}
	return strings.Join(names, ", ")
}

type TrackComparator struct{}

func (TrackComparator) Compare(a, b *metadatapb.Track) int {
	ua, ub := TrackURI(a), TrackURI(b)
	if ua == ub {
		return 0
	}
	if ua == "" {
		return 1
	}
	if ub == "" {
		return -1
	}
	if ua < ub {
		return -1
	}
	return 1
}

func (TrackComparator) Key(a *metadatapb.Track) string {
	return TrackURI(a)
}

var _ reactive.Comparator[*metadatapb.Track, string] = (*TrackComparator)(nil)

// Playlist is a lightweight representation of a playlist.
type Playlist struct {
	URI         string
	Name        string
	Description string
	Owner       string
	TrackCount  int
	CreatedAt   time.Time
}

func (p *Playlist) FromProto(pb *playlist4pb.SelectedListContent) {
	*p = Playlist{
		Owner:      pb.GetOwnerUsername(),
		TrackCount: int(pb.GetLength()),
	}
	if pb.Attributes != nil {
		p.Name = pb.Attributes.GetName()
		p.Description = pb.Attributes.GetDescription()
	}
	if pb.CreatedAt != nil {
		p.CreatedAt = time.Unix(*pb.CreatedAt, 0)
	}
}
