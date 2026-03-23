package types

import (
	"encoding/hex"
	"time"

	librespot "github.com/devgianlu/go-librespot"
	metadatapb "github.com/devgianlu/go-librespot/proto/spotify/metadata"
	playlist4pb "github.com/devgianlu/go-librespot/proto/spotify/playlist4"

	"github.com/pyrorhythm/spqt/pkg/reactive"
)

type Image struct {
	URL    string
	Width  int32
	Height int32
}

type Track struct {
	URI        string
	Name       string
	Album      Album
	Artists    []ArtistRef
	TrackNum   int32
	DiscNum    int32
	DurationMs int32
	Popularity int32
	Explicit   bool
}

type TrackComparator struct{}

func (TrackComparator) Compare(a, b EnrichedTrack) int {
	if a.Track == nil && b.Track == nil {
		return 0
	}

	if a.Track == nil {
		return 1
	}

	if b.Track == nil {
		return -1
	}

	if a.URI != b.URI {
		if a.URI < b.URI {
			return -1
		}

		return 1
	}

	return 0
}

func (TrackComparator) Key(a EnrichedTrack) string {
	return a.URI
}

var _ reactive.Comparator[EnrichedTrack, string] = (*TrackComparator)(nil)

type ArtistRef struct {
	URI  string
	Name string
}

type Album struct {
	URI        string
	Name       string
	Artists    []ArtistRef
	Type       string
	Label      string
	Year       int
	Covers     []Image
	TrackCount int
	Popularity int32
}

type Artist struct {
	URI        string
	Name       string
	Popularity int32
	Portraits  []Image
	Genres     []string
}

type Playlist struct {
	URI         string
	Name        string
	Description string
	Owner       string
	TrackCount  int
	CreatedAt   time.Time
}

// --- Converters ---

func (t *Track) FromProto(pb *metadatapb.Track) {
	*t = Track{
		Name:       pb.GetName(),
		TrackNum:   pb.GetNumber(),
		DiscNum:    pb.GetDiscNumber(),
		DurationMs: pb.GetDuration(),
		Popularity: pb.GetPopularity(),
		Explicit:   pb.GetExplicit(),
	}

	if pb.GetCanonicalUri() != "" {
		t.URI = pb.GetCanonicalUri()
	} else if len(pb.Gid) > 0 {
		t.URI = librespot.SpotifyIdFromGid(librespot.SpotifyIdTypeTrack, pb.Gid).Uri()
	}

	if pb.Album != nil {
		t.Album = Album{}
		t.Album.FromProto(pb.Album)
	}

	for _, a := range pb.Artist {
		t.Artists = append(t.Artists, artistRefFromProto(a))
	}
}

func (a *Album) FromProto(pb *metadatapb.Album) {
	*a = Album{
		Name:       pb.GetName(),
		Label:      pb.GetLabel(),
		Popularity: pb.GetPopularity(),
		Type:       pb.GetTypeStr(),
	}

	if len(pb.Gid) > 0 {
		a.URI = librespot.SpotifyIdFromGid(librespot.SpotifyIdTypeTrack, pb.Gid).Uri()
		// NOTE: SpotifyIdType for albums doesn't exist in go-librespot,
		// construct URI manually if needed: "spotify:album:" + librespot.GidToBase62(pb.Gid)
	}

	if pb.Date != nil {
		a.Year = int(pb.Date.GetYear())
	}

	for _, ar := range pb.Artist {
		a.Artists = append(a.Artists, artistRefFromProto(ar))
	}

	for _, d := range pb.Disc {
		a.TrackCount += len(d.GetTrack())
	}

	a.Covers = imagesFromProto(pb.Cover)
}

func (a *Artist) FromProto(pb *metadatapb.Artist) {
	*a = Artist{
		Name:       pb.GetName(),
		Popularity: pb.GetPopularity(),
	}

	if len(pb.Gid) > 0 {
		a.URI = "spotify:artist:" + librespot.GidToBase62(pb.Gid)
	}

	a.Portraits = imagesFromProto(pb.Portrait)
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

func artistRefFromProto(pb *metadatapb.Artist) ArtistRef {
	ref := ArtistRef{Name: pb.GetName()}

	if len(pb.Gid) > 0 {
		ref.URI = "spotify:artist:" + librespot.GidToBase62(pb.Gid)
	}

	return ref
}

func imagesFromProto(pbs []*metadatapb.Image) []Image {
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

type EnrichedTrack struct {
	*Track
	FullAlbum   *Album
	FullArtists []*Artist
}
