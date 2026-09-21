package qbittorrent

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/EastArctica/qbitslskd/slskd"
)

func DeleteTorrentHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != "DELETE" && req.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := req.ParseForm(); err != nil {
		fmt.Printf("deleteTorrentHandler: Failed to read form: %s\n", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if !req.Form.Has("hashes") {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// TODO: Check this, delete files from disk if true
	if !req.Form.Has("deleteFiles") {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	apiKey, ok := APIKey(req)
	if !ok {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	hashesStr := req.Form.Get("hashes")
	hashes := strings.Split(hashesStr, "|")

	// TODO: multithread slskd deletions
	downloads, err := slskd.GetDownloads(apiKey)
	if err != nil {
		fmt.Printf("deleteTorrentHandler: Failed to get slskd downloads: %s\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	for _, hash := range hashes {
		var foundIds []string
		var username string

	DownloadsLoop:
		for _, user := range downloads {
			for _, dir := range user.Directories {
				newHash, err := sha1Hash(user.Username + dir.Directory)
				if err != nil {
					continue
				}

				if hash == newHash {
					username = user.Username
					for _, file := range dir.Files {
						foundIds = append(foundIds, file.ID)
					}

					break DownloadsLoop
				}
			}
		}

		for _, id := range foundIds {
			err := slskd.DeleteFile(apiKey, username, id, true)
			if err != nil {
				fmt.Printf("deleteTorrentHandler: Failed to create delete files: %s\n", err)
				continue
			}
		}
	}
}
