package qbittorrent

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/EastArctica/qbitslskd/config"
	"github.com/EastArctica/qbitslskd/models"
	"github.com/EastArctica/qbitslskd/slskd"
	"github.com/EastArctica/qbitslskd/utils"
)

func TorrentsInfoHandler(w http.ResponseWriter, req *http.Request, cache *models.Cache) {
	fmt.Printf("Request to path %s\n", req.URL.Path)

	// There's a lot of scawwy parameters to this :3
	// TODO: filter, category, tag, sort, reverse, limit, offset, hashes

	apiKey, ok := APIKey(req)
	if !ok {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	users, err := slskd.GetDownloads(apiKey)
	if err != nil {
		fmt.Printf("torrentsInfoHandler: Failed to get slskd downloads: %s\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// qBittorrent filters the list server side, and Lidarr asks for its own
	// category.
	categoryFilter := req.URL.Query().Get("category")

	// Convert downloads to torrents info
	// TODO: This doesn't need to be a slice, we can determine the size by summing the dirs
	var torrents = []models.Transate_QBTorrentInfo{}

	for _, user := range users {
		for _, dir := range user.Directories {
			var firstAddedAt int64
			var totalBytes int
			var bytesRemaining int
			// epoch seconds, -1 if incomplete
			var completionTime *time.Time = nil
			var downloadSpeed int64 = 0
			var status models.Status = models.StatusDownloading

			// The release hash is the only handle we have on a slskd download,
			// so it is what ties the transfer back to the search result Lidarr
			// grabbed, and with it the release name and the infohash.
			release := releaseHash(user.Username, dir.Directory)

			cache.Mutex.Lock()
			cacheEntry, known := cache.Search[release]
			cache.Mutex.Unlock()

			if categoryFilter != "" && cacheEntry.Category != categoryFilter {
				continue
			}

			// Lidarr waits for a torrent carrying the infohash it computed from
			// the .torrent we served; reporting the release hash instead leaves
			// the queue item unmatched, which is why no album details show up
			// while a download is running.
			hash := cacheEntry.InfoHash
			if hash == "" {
				hash = release
			}

			// Falling back to the directory keeps an unknown release visible
			// with something readable rather than a blank row.
			name := cacheEntry.Name
			if name == "" {
				dirParts := strings.Split(dir.Directory, "\\")
				name = dirParts[len(dirParts)-1]
			}

			if !known {
				fmt.Printf("torrentsInfoHandler: no cached release for %s%s\n", user.Username, dir.Directory)
			}

			for _, file := range dir.Files {
				totalBytes += file.Size
				bytesRemaining += file.BytesRemaining

				// This should get the latest downloading file's speed
				if file.AverageSpeed != 0 {
					downloadSpeed = int64(file.AverageSpeed)
				}

				// Find earliest enqueued file
				enqueuedTime, err := time.Parse(time.RFC3339, file.EnqueuedAt)
				if err == nil {
					enq := enqueuedTime.Unix()
					if firstAddedAt == 0 || enq < firstAddedAt {
						firstAddedAt = enq
					}
				}

				// Track the latest completion time for when all files are done
				if !file.EndedAt.IsZero() {
					if completionTime == nil || file.EndedAt.After(*completionTime) {
						completionTime = &file.EndedAt
					}
				}
			}

			// TODO: Implement more types of state
			// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#get-torrent-list
			if bytesRemaining == 0 {
				status = models.StatusCompleted
			}

			if utils.Some(dir.Files, func(file slskd.SlskdDownloadsFiles) bool { return utils.Includes(file.State, "Errored") }) {
				status = models.StatusError
			}

			// Get the full content path
			// Note: This is not a bug, but more of a note. If slskd ever switches off always using backslashes for directories, this will break
			// TODO: Get the actual download directory from slskd instead of forcing the user to specify it
			splitPath := strings.Split(dir.Directory, "\\")
			path := splitPath[len(splitPath)-1]
			if bytesRemaining == 0 {
				path = config.COMPLETE_DIR + path
			} else {
				path = config.INCOMPLETE_DIR + path
			}

			t := models.QbtTorrent{
				Hash:           hash,
				Name:           name,
				Size:           int64(totalBytes),
				CompletedBytes: int64(totalBytes - bytesRemaining),
				DownloadSpeed:  int64(downloadSpeed),
				Status:         status,
				AddedAt:        time.Unix(firstAddedAt, 0),
				CompletedAt:    completionTime,
				Category:       cacheEntry.Category,
				SavePath:       path,
				SourceUser:     user.Username,
			}

			torrent := models.QBTorrentInfoFromCore(t)

			torrents = append(torrents, torrent)
		}
	}

	// Marshal torrents and send it back
	torrentsData, err := json.Marshal(torrents)
	if err != nil {
		fmt.Printf("torrentsInfoHandler: Failed to marshal torrents\n")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.Write(torrentsData)
}

// TODO: Move these two functions elsewhere (somewhere more accessible)
func findAudioFile(files []slskd.SlskdDownloadsFiles) (*slskd.SlskdDownloadsFiles, error) {
	for _, file := range files {
		if isAudioFile(file.Filename) {
			return &file, nil
		}
	}

	return nil, fmt.Errorf("no audio files found")
}

func isAudioFile(path string) bool {
	for _, extension := range AUDIO_EXTENSIONS {
		if bytes.HasSuffix([]byte(path), []byte(extension)) {
			return true
		}
	}

	return false
}

var AUDIO_EXTENSIONS = []string{
	".mp3",
	".aac",
	".ogg",
	".wma",
	".opus",
	".m4a",
	".mp2",
	".ac3",
	".eac3",
	".dts",
	".amr",
	".awb",
	".ra",
	".ram",
	".flac",
	".alac",
	".ape",
	".wv",
	".tta",
	".shn",
	".wav",
	".aiff",
	".aif",
	".pcm",
	".au",
	".bwf",
	".rf64",
	".mid",
	".midi",
	".kar",
	".rmi",
	".mod",
	".s3m",
	".xm",
	".it",
	".mtm",
	".umx",
	".cda",
	".dss",
	".ds2",
	".dvf",
	".msv",
	".gsm",
	".vox",
	".sln",
	".voc",
	".iff",
	".svx",
}

func sha1Hash(str string) (string, error) {
	hasher := sha1.New()
	_, err := hasher.Write([]byte(str))
	if err != nil {
		return "", fmt.Errorf("failed to write to hasher: %w", err)
	}

	hashBytes := hasher.Sum(nil)
	hashHex := fmt.Sprintf("%x", hashBytes)

	return hashHex, nil
}

func releaseHash(username string, directory string) string {
	hasher := sha1.New()
	// hash.Hash promises that Write never returns an error.
	hasher.Write([]byte(username + directory))

	return fmt.Sprintf("%x", hasher.Sum(nil))
}
