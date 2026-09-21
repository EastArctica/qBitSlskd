package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/EastArctica/qbitslskd/config"
	"github.com/EastArctica/qbitslskd/models"
	qbittorrent "github.com/EastArctica/qbitslskd/qBitTorrent"
	"github.com/EastArctica/qbitslskd/torznab"
)

var cache models.Cache = models.Cache{
	Search:     make(map[string]models.SearchCacheEntry),
	Categories: make(map[string]models.Category),
}

func HanndleQbtApiRequests(w http.ResponseWriter, req *http.Request) {
	cache_ptr := &cache

	switch req.URL.Path {
	case "/api/v2/app/webapiVersion":
		qbittorrent.Version(w)
	case "/api/v2/auth/login":
		qbittorrent.Login(w, req)
	case "/api/v2/app/preferences":
		qbittorrent.Preferences(w)
	case "/api/v2/torrents/categories":
		qbittorrent.CategoriesHandler(w, req, cache_ptr)
	case "/api/v2/torrents/createCategory":
		qbittorrent.CreateCategoryHandler(w, req, cache_ptr)
	case "/api/v2/torrents/info":
		qbittorrent.TorrentsInfoHandler(w, req, cache_ptr)
	case "/api/v2/torrents/add":
		qbittorrent.AddTorrentHandler(w, req, cache_ptr)
	case "/api/v2/torrents/delete":
		break
	}
}

func HanndleTorznabApiRequests(w http.ResponseWriter, req *http.Request) {
	cache_ptr := &cache

	functionType := req.URL.Query().Get("t")
	switch functionType {
	case "caps":
		torznab.CapabilitiesHandler(w, req)
		return
	case "search":
		torznab.SearchHandler(w, req, cache_ptr)
		return
	case "custom_download":
		torznab.CustomDownloadHandler(w, req, cache_ptr) // Note: This is NOT a real torznab function and is custom to qBitSlskd
		return
	}

}

func main() {
	config.Init()
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v2/", HanndleQbtApiRequests)
	mux.HandleFunc("/api", HanndleTorznabApiRequests)

	fmt.Printf("qBitSlskd started on port %s!\n", config.PORT)
	http.ListenAndServe(fmt.Sprintf(":%s", config.PORT), logAll(mux))
}

func logAll(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Print request line
		fmt.Printf("%s %s %s\n", req.Method, req.URL.RequestURI(), req.Proto)
		// Print headers
		for name, values := range req.Header {
			for _, v := range values {
				fmt.Printf("%s: %s\n", name, v)
			}
		}
		fmt.Println()

		// Print body
		if req.Body != nil {
			bodyBytes, err := io.ReadAll(req.Body)
			if err != nil {
				fmt.Printf("Error reading body: %v\n", err)
			} else if len(bodyBytes) > 0 {
				fmt.Printf("%s\n", string(bodyBytes))
			}
			// Restore Body so handlers down‐stream (if any) can still read it
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}

		next.ServeHTTP(w, req)
	})
}
