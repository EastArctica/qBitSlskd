package main

import (
	"fmt"
	"net/http"

	"github.com/EastArctica/qbitslskd/config"
	"github.com/EastArctica/qbitslskd/models"
	qbittorrent "github.com/EastArctica/qbitslskd/qBitTorrent"
	"github.com/EastArctica/qbitslskd/torznab"
)

var cache models.Cache = models.Cache{
	Search: make(map[string]models.SearchCacheEntry),
}

// var categories map[string]Category = map[string]Category{
// 	"lidarr": {
// 		Name:     "lidarr",
// 		SavePath: "",
// 	},
// }

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
		break
	case "/api/v2/torrents/add":
		break
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
		// This is NOT a real torznab function and is custom to qBitSlskd
		torznab.CustomDownloadHandler(w, req, cache_ptr)
		return
	}

}

func main() {
	config.Init()
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v2/", HanndleQbtApiRequests)
	mux.HandleFunc("/api", HanndleTorznabApiRequests)

	// mux.Handle("/", logAll())

	fmt.Printf("qBitSlskd started on port %s!\n", config.PORT)
	http.ListenAndServe(fmt.Sprintf(":%s", config.PORT), mux)
}
