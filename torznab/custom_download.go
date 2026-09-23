package torznab

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"net/http"
	"time"

	"github.com/EastArctica/qbitslskd/album_lookup"
	"github.com/EastArctica/qbitslskd/models"
	"github.com/jackpal/bencode-go"
)

// Nothing is ever transferred over BitTorrent, so the value itself is arbitrary,
// but the piece count derived from it has to agree with the total size.
const torrentPieceLength = 262144

func CustomDownloadHandler(w http.ResponseWriter, req *http.Request, cache *models.Cache) {
	if !req.URL.Query().Has("id") {
		http.Error(w, "Missing id param", http.StatusBadRequest)
		return
	}

	// Ensure we have an api key
	apiKey := req.URL.Query().Get("apikey")
	if apiKey == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// TODO: Verify key? uuid is kinda secure
	cacheId := req.URL.Query().Get("id")

	cache.Mutex.Lock()
	cacheEntry, ok := cache.Search[cacheId]
	cache.Mutex.Unlock()
	if !ok {
		http.Error(w, "id param does not map to a valid cache entry", http.StatusBadRequest)
		return
	}

	// A release with no files bencodes to a torrent with an empty file list,
	// which nothing downstream can act on.
	if len(cacheEntry.Files) == 0 {
		fmt.Printf("CustomDownloadHandler: cache entry %s has no files\n", cacheId)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Generate a fake sha1 hash
	hasher := sha1.New()
	// hash.Hash promises that Write never returns an error.
	hasher.Write([]byte(cacheEntry.Username + cacheEntry.Name + cacheEntry.SearchedAt.String()))

	hashBytes := hasher.Sum(nil)

	torrentFiles := make([]models.TorrentFileInfoFile, 0, len(cacheEntry.Files))
	var totalSize int64

	for _, file := range cacheEntry.Files {
		totalSize += file.Size

		torrentFiles = append(torrentFiles, models.TorrentFileInfoFile{
			Length: file.Size,
			// path is a list of path components, not one string, and it is
			// relative to info.name. Every file in a release comes from the
			// same Soulseek directory (see groupByDirectory) and name is
			// already that directory, so each file sits directly beneath it.
			Path: []string{album_lookup.SlskdBasename(file.Filename)},
		})
	}

	// pieces is one sha1 per piece, so a parser that checks its length against
	// the total size rejects a single digest outright. Nothing ever hashes a
	// piece, so the same fake digest stands in for all of them.
	pieceCount := (totalSize + torrentPieceLength - 1) / torrentPieceLength
	if pieceCount == 0 {
		pieceCount = 1
	}

	pieces := make([]byte, 0, pieceCount*sha1.Size)
	for i := int64(0); i < pieceCount; i++ {
		pieces = append(pieces, hashBytes...)
	}

	torrentFile := models.TorrentFile{
		Announce:     "udp://example.com:1337/announce",
		AnnounceList: [][]string{{"udp://example.com:1337/announce"}},
		Comment:      "qBitSlskd fake torrent file",
		CreatedBy:    "qBitSlskd",
		CreationDate: int(time.Now().Unix()),
		Info: models.TorrentFileInfo{
			Files:       torrentFiles,
			Name:        cacheEntry.Name,
			PieceLength: torrentPieceLength,
			Pieces:      pieces,
			Private:     0,
			Username:    cacheEntry.Username,
			CacheId:     cacheId,
		},
		UrlList: []string{},
	}

	// Lidarr tracks the download by the infohash it computes from this file,
	// not by the id in the URL, and then waits for a torrent carrying that hash
	// to appear in /api/v2/torrents/info. Indexing the entry under it here is
	// what lets that endpoint resolve the grab back to a Soulseek user and path.
	var infoData bytes.Buffer
	err := bencode.Marshal(&infoData, torrentFile.Info)
	if err != nil {
		fmt.Printf("CustomDownloadHandler: Failed to marshal torrent info: %s\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	infoHash := fmt.Sprintf("%x", sha1.Sum(infoData.Bytes()))

	// Buffered so that a marshal failure is still a clean error response rather
	// than an error page appended to a half written torrent.
	var torrentData bytes.Buffer
	err = bencode.Marshal(&torrentData, torrentFile)
	if err != nil {
		fmt.Printf("CustomDownloadHandler: Failed to marshal torrent file: %s\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Both keys point at the same release, so both need the infohash: the
	// release-hash entry is the one /api/v2/torrents/info reaches (it can only
	// derive a release hash from the slskd username and directory), and that is
	// where the hash Lidarr is waiting for has to be readable from.
	cacheEntry.InfoHash = infoHash

	cache.Mutex.Lock()
	cache.Search[cacheId] = cacheEntry
	cache.Search[infoHash] = cacheEntry
	cache.Mutex.Unlock()

	w.Header().Set("Content-Type", "application/x-bittorrent")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", cacheEntry.Name+".torrent"))
	w.Write(torrentData.Bytes())
}
