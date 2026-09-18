package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/EastArctica/qbitslskd/config"
	"github.com/EastArctica/qbitslskd/models"
)

type Cache struct {
	CacheMutex  sync.Mutex
	SearchCache map[string]models.SearchCacheEntry
}

var cache Cache = Cache{
	SearchCache: make(map[string]models.SearchCacheEntry),
}

func HanndleQbtApiRequests(w http.ResponseWriter, req *http.Request) {
	// cache_ptr := &cache

	switch req.URL.Path {
	case "/api/v2/app/webapiVersion":
		break
	case "/api/v2/auth/login":
		break
	case "/api/v2/app/preferences":
		break
	case "/api/v2/torrents/categories":
		break
	case "/api/v2/torrents/createCategory":
		break
	case "/api/v2/torrents/info":
		break
	case "/api/v2/torrents/add":
		break
	case "/api/v2/torrents/delete":
		break
	}
}

func HanndleTorznabApiRequests(w http.ResponseWriter, req *http.Request) {
	// cache_ptr := &cache

	//

}

func main() {
	config.Init()
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v2/", HanndleQbtApiRequests)
	mux.HandleFunc("/api/", HanndleTorznabApiRequests)

	// mux.Handle("/", logAll())

	fmt.Printf("qBitSlskd started on port %s!\n", config.PORT)
	http.ListenAndServe(fmt.Sprintf(":%s", config.PORT), mux)
}
