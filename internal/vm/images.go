package vm

import (
	"context"
	"io"
	"net/http"
	"sync"
	"time"

	metadatapb "github.com/devgianlu/go-librespot/proto/spotify/metadata"
	"github.com/dgraph-io/badger/v4"
	qt "github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/mainthread"
	"golang.org/x/sync/singleflight"

	"github.com/pyrorhythm/spqt/internal/types"
	"github.com/pyrorhythm/spqt/pkg/log"
	"github.com/pyrorhythm/spqt/pkg/qtw"
)

const imageCacheMax = 200

// sizeSuffix returns the Spotify CDN size suffix for the given target pixel size.
func sizeSuffix(targetPx int) string {
	switch {
	case targetPx <= 64:
		return "-64"
	case targetPx <= 128:
		return "-128"
	case targetPx <= 256:
		return "-256"
	case targetPx <= 320:
		return "-320"
	case targetPx <= 640:
		return "-640"
	default:
		return ""
	}
}

type ImageService struct {
	ctx context.Context
	db  *badger.DB
	mem map[string]*qt.QPixmap
	// head is the index of the oldest entry in the ring.
	// A new []string allocation is used for eviction to keep order slice bounded.
	order []string
	mu    sync.Mutex
	sf    singleflight.Group
}

func NewImageService(ctx context.Context, db *badger.DB) *ImageService {
	return &ImageService{
		ctx:   log.Span(ctx, "images"),
		db:    db,
		mem:   make(map[string]*qt.QPixmap, imageCacheMax),
		order: make([]string, 0, imageCacheMax),
	}
}

func BestURL(images []types.Image, targetPx int) string {
	if len(images) == 0 {
		return ""
	}
	var best types.Image
	bestSet := false
	var largest types.Image
	largestSize := int32(0)

	for _, img := range images {
		sz := max(img.Height, img.Width)

		if sz > largestSize {
			largestSize = sz
			largest = img
		}

		if sz >= int32(targetPx) {
			if !bestSet || sz < best.Width && sz < best.Height {
				best = img
				bestSet = true
			}
		}

	}
	if bestSet {
		return best.URL
	}
	return largest.URL
}

func (s *ImageService) LoadCover(album *metadatapb.Album, targetPx int, onLoad func(*qt.QPixmap)) {
	covers := types.AlbumCovers(album)
	url := BestURL(covers, targetPx)
	if url == "" {
		return
	}
	url += sizeSuffix(targetPx)
	s.Load(url, onLoad)
}

func (s *ImageService) Load(url string, onLoad func(*qt.QPixmap)) {
	if url == "" {
		return
	}

	s.mu.Lock()
	if pm, ok := s.mem[url]; ok {
		s.mu.Unlock()
		onLoad(pm)
		return
	}
	s.mu.Unlock()

	go func() {
		val, err, _ := s.sf.Do(url, func() (any, error) {
			return s.fetch(url)
		})
		if err != nil || val == nil {
			return
		}
		qtw.T(onLoad, val.(*qt.QPixmap))
	}()
}

func (s *ImageService) fetch(url string) (*qt.QPixmap, error) {
	var data []byte
	var fromCache bool

fetch:
	if s.db != nil && !fromCache {
		_ = s.db.View(func(txn *badger.Txn) error {
			item, err := txn.Get([]byte("img:" + url))
			if err != nil {
				return err
			}
			data, err = item.ValueCopy(nil)
			return err
		})
		if data != nil {
			fromCache = true
		}
	}

	if data == nil {
		req, err := http.NewRequestWithContext(s.ctx, http.MethodGet, url, nil)
		if err != nil {

			log.Warn(s.ctx).
				Err(err).Str("url", url).
				Msg("fetch failed")

			return nil, err
		}
		httpClient := &http.Client{Timeout: 10 * time.Second}
		resp, err := httpClient.Do(req)
		if err != nil {

			log.Warn(s.ctx).
				Err(err).Str("url", url).
				Msg("fetch failed")

			return nil, err
		}
		defer resp.Body.Close()
		data, err = io.ReadAll(resp.Body)
		if err != nil {

			log.Warn(s.ctx).
				Err(err).Str("url", url).
				Msg("fetch failed")

			return nil, err
		}

		log.Trace(s.ctx).Str("url", url).Int("bytes", len(data)).Msg("http fetch")

		if s.db != nil {
			_ = s.db.Update(func(txn *badger.Txn) error {
				return txn.Set([]byte("img:"+url), data)
			})
		}
	}

	var pm *qt.QPixmap
	mainthread.Wait(func() {
		p := qt.NewQPixmap()
		if p.LoadFromDataWithData(data) {
			pm = p
		}
	})

	if pm == nil {
		log.Warn(s.ctx).Str("url", url).Bool("fromCache", fromCache).Msg("decode failed")
		if fromCache && s.db != nil {
			// Corrupt cache entry - delete and re-fetch from network
			_ = s.db.Update(func(txn *badger.Txn) error {
				return txn.Delete([]byte("img:" + url))
			})
			data = nil
			fromCache = false
			goto fetch
		}
		return nil, nil
	}

	s.mu.Lock()
	s.memPut(url, pm)
	s.mu.Unlock()

	return pm, nil
}

func (s *ImageService) memPut(url string, pm *qt.QPixmap) {
	if len(s.order) >= imageCacheMax {
		evict := s.order[0]
		s.order = append(s.order[:0], s.order[1:]...)
		delete(s.mem, evict)
	}
	s.mem[url] = pm
	s.order = append(s.order, url)
}

func (s *ImageService) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mem = make(map[string]*qt.QPixmap)
	s.order = s.order[:0]
}

