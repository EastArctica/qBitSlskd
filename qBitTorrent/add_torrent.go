package qbittorrent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/EastArctica/qbitslskd/config"
	"github.com/EastArctica/qbitslskd/models"
	"github.com/jackpal/bencode-go"
)

func AddTorrentHandler(w http.ResponseWriter, req *http.Request, cache *models.Cache) {
	if req.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	apiKey, ok := APIKey(req)
	if !ok {
		fmt.Printf("AddTorrentHandler: no SID cookie or basic auth credentials\n")
		http.Error(w, "Fails.", 200)
		return
	}

	req.ParseMultipartForm(100 << 10) // 100MB

	// We only support file upload directly, no url or any of that
	file, _, err := req.FormFile("torrents")
	if err != nil {
		fmt.Printf("req.FormFile fail: %s\n", err)
		http.Error(w, "Fails.", 200)
		return
	}

	fileData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var torrentFile models.TorrentFile
	err = bencode.Unmarshal(bytes.NewReader(fileData), &torrentFile)
	if err != nil {
		fmt.Printf("Invalid file: %s\n", err)
		http.Error(w, "file decode Fails.", 200)
		return
	}

	cache.Mutex.Lock()
	cacheEntry, ok := cache.Search[torrentFile.Info.CacheId]
	cache.Mutex.Unlock()

	if !ok {
		fmt.Printf("Cache Fail\n")
		http.Error(w, "Fails.", 200)
		return
	}

	// Lidarr only imports a download that comes back tagged with the category
	// it grabbed under, and slskd knows nothing about categories, so the value
	// has to be remembered here and replayed by /api/v2/torrents/info.
	category := req.PostFormValue("category")
	cacheEntry.Category = category

	cache.Mutex.Lock()
	cache.Search[torrentFile.Info.CacheId] = cacheEntry
	if cacheEntry.InfoHash != "" {
		cache.Search[cacheEntry.InfoHash] = cacheEntry
	}
	// qBittorrent creates an unknown category on add, and Lidarr checks that
	// its category exists before it will import.
	if category != "" {
		if _, exists := cache.Categories[category]; !exists {
			cache.Categories[category] = models.Category{
				Name:     category,
				SavePath: req.PostFormValue("savepath"),
			}
		}
	}
	cache.Mutex.Unlock()

	filesRaw, err := json.Marshal(cacheEntry.Files)
	if err != nil {
		fmt.Printf("addTorrentHandler: Failed to marshal cacheEntry.files: %s\n", err)
		http.Error(w, "Fails.", 200)
		return
	}

	// Download files
	client := &http.Client{}
	downloadReq, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v0/transfers/downloads/%s", config.SLSKD_ROOT, cacheEntry.Username), bytes.NewBuffer(filesRaw))
	if err != nil {
		fmt.Printf("addTorrentHandler: Failed to create download request: %s\n", err)
		http.Error(w, "Fails.", 200)
		return
	}

	downloadReq.Header.Add("X-API-Key", apiKey)
	downloadReq.Header.Add("Content-Type", "application/json")

	downloadRes, err := client.Do(downloadReq)
	if err != nil {
		fmt.Printf("addTorrentHandler: Failed to send download request: %s\n", err)
		http.Error(w, "Fails.", 200)
		return
	}

	if downloadRes.StatusCode == http.StatusUnauthorized || downloadRes.StatusCode == http.StatusForbidden {
		http.Error(w, "Fails.", 200)
		return
	}

	if downloadRes.StatusCode == http.StatusCreated {
		fmt.Fprintf(w, "Ok.")
		return
	}

	resBody, err := io.ReadAll(downloadRes.Body)
	if err != nil {
		fmt.Printf("addTorrentHandler: Unable to read body from slskd download: %d\n", downloadRes.StatusCode)
		http.Error(w, "Fails.", 200)
		return
	}

	fmt.Printf("addTorrentHandler: Bad response from slskd download: %d - %s\n", downloadRes.StatusCode, string(resBody))
	http.Error(w, "Fails.", 200)
}
