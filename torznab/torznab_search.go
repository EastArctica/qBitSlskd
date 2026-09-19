package torznab

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/EastArctica/qbitslskd/config"
	"github.com/EastArctica/qbitslskd/models"
	"github.com/google/uuid"
)

func SearchHandler(w http.ResponseWriter, req *http.Request, cache *models.Cache) {
	startSearchResponseData, http_err := initSearch(req)
	if http_err != nil {
		http.Error(w, http_err.Err, http_err.Code)
		return
	}

	var startSearchResponse models.SearchResponse
	err := json.Unmarshal(startSearchResponseData, &startSearchResponse)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	apiKey := req.URL.Query().Get("apikey")
	if apiKey == "" {
		println("No api key")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var search_response *models.SearchResponse = nil

	search_response, http_err = blockUntilSearchComplete(startSearchResponse.ID, apiKey)
	if http_err != nil {
		http.Error(w, http_err.Err, http_err.Code)
		return
	}

	fmt.Println(search_response.ID)

	var search_results []models.SearchResult = make([]models.SearchResult, 0)

	search_results, http_err = getSearchResults(search_response.ID, apiKey)
	if http_err != nil {
		http.Error(w, http_err.Err, http_err.Code)
		return
	}

	items := buildItems(search_results, cache, apiKey)

	var searchResults = models.TorznabSearchResults{
		Version: "2.0",
		Atom:    "http://www.w3.org/2005/Atom",
		Torznab: "http://torznab.com/schemas/2015/feed",
		Channel: models.TorznabSearchChannel{
			Link: models.TorznabChannelLink{
				Href: req.URL.String(),
				Rel:  "self",
				Type: "application/rss+xml",
			},
			Title:       "qBitSlskd - Search Results",
			Description: "Search results for '" + req.URL.Query().Get("q") + "' from qBitSlskd",
			// I'm not sure if something should be done related to this...
			Language:  "en-us",
			WebMaster: "example@example.com",
			Category:  "search",
			Image: models.TorznabChannelImage{
				URL:         "https://avatars.githubusercontent.com/u/76762370",
				Title:       "qBitSlskd",
				Link:        "https://github.com/EastArctica/qBitSlskd",
				Description: "slskd Logo",
			},
			// Cache time (minutes)
			Ttl: "30",
			Response: models.TorznabChannelResponse{
				Newznab: "http://www.newznab.com/DTD/2010/feeds/attributes/",
				// TODO: Offset (Is this possible with slskd?)
				Offset: "0",
				Total:  strconv.Itoa(len(items)),
			},
			Item: items,
		},
	}

	data, err := xml.Marshal(searchResults)
	if err != nil {
		fmt.Printf("Failed to marshal search results %s\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// fmt.Printf("Finished search for %s\n", )

	w.Header().Set("response-type", "application/xml")
	fmt.Fprint(w, string(data))
}

func initSearch(req *http.Request) ([]byte, *models.HttpError) {
	searchId, err := uuid.NewUUID()
	if err != nil {
		fmt.Printf("Failed to generate new UUID: %s\n", err.Error())
		return nil, models.HttpError_From("Internal server error", http.StatusInternalServerError)
	}

	responseLimit := 250
	if req.URL.Query().Has("limit") {
		queryLimit, err := strconv.Atoi(req.URL.Query().Get("limit"))
		if err == nil {
			responseLimit = queryLimit
		}
	}

	// Ensure we have an api key
	apiKey := req.URL.Query().Get("apikey")
	if apiKey == "" {
		println("No api key")
		return nil, models.HttpError_From("Unauthorized", http.StatusUnauthorized)
	}

	// Search text, it can't be empty and prowlarr tests it with an empty string
	searchText := req.URL.Query().Get("q")
	if searchText == "" {
		searchText = "mu"
	}

	// Actually search :tear:
	searchReq := models.SearchRequest{
		Id:              searchId.String(),
		FileLimit:       100000,
		FilterResponses: true,
		// TODO: Queue length env variable (actually env variable for all this)
		MaximumPeerQueueLength:   1500,
		MinimumPeerUploadSpeed:   0,
		MinimumResponseFileCount: 1,
		ResponseLimit:            responseLimit,
		SearchText:               searchText,
		// idk 30 second timeout?
		SearchTimeout: 30000,
	}

	searchReqData, err := json.Marshal(searchReq)
	if err != nil {
		fmt.Printf("Failed to marshal searchReq: %s\n", err)
		return nil, models.HttpError_From("Internal server error", http.StatusInternalServerError)
	}

	// Search with query
	client := &http.Client{}
	startSearchReq, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v0/searches", config.SLSKD_ROOT), bytes.NewBuffer(searchReqData))
	if err != nil {
		fmt.Printf("Failed to create search request: %s\n", err)
		return nil, models.HttpError_From("Internal Server Error", http.StatusInternalServerError)
	}

	startSearchReq.Header.Add("X-API-Key", apiKey)
	startSearchReq.Header.Add("Content-Type", "application/json")

	startSearchRes, err := client.Do(startSearchReq)
	if err != nil {
		// TODO: Actually check the error
		return nil, models.HttpError_From("Internal server error", http.StatusInternalServerError)
	}

	if startSearchRes.StatusCode == 401 || startSearchRes.StatusCode == 403 {
		return nil, models.HttpError_From("Forbidden", http.StatusForbidden)
	}

	startSearchResData, err := io.ReadAll(startSearchRes.Body)
	if err != nil {
		fmt.Printf("Failed to read search response %s\n", err)
		return nil, models.HttpError_From("Internal Server Error", http.StatusInternalServerError)
	}

	return startSearchResData, nil
}

func blockUntilSearchComplete(id string, apiKey string) (*models.SearchResponse, *models.HttpError) {
	client := &http.Client{}

	statusURL := fmt.Sprintf("%s/api/v0/searches/%s", config.SLSKD_ROOT, id)
	statusReq, err := http.NewRequest("GET", statusURL, nil)
	if err != nil {
		return nil, models.HttpError_From("Internal server error", http.StatusInternalServerError)
	}

	// include API key header
	statusReq.Header.Add("X-API-Key", apiKey)

	for {

		statusRes, err := client.Do(statusReq)
		if err != nil {
			return nil, models.HttpError_From("Internal server error", http.StatusInternalServerError)
		}

		if statusRes.StatusCode == 403 {
			return nil, models.HttpError_From("Forbidden", http.StatusForbidden)
		}

		statusBody, err := io.ReadAll(statusRes.Body)
		statusRes.Body.Close()
		if err != nil {
			return nil, models.HttpError_From("Internal server error", http.StatusInternalServerError)
		}

		// parse into SearchStateResponse
		var state models.SearchResponse
		if err := json.Unmarshal(statusBody, &state); err != nil {
			return nil, models.HttpError_From("Internal server error", http.StatusInternalServerError)
		}

		if state.IsComplete {
			return &state, nil
		}

		time.Sleep(1 * time.Second)
	}
}

func getSearchResults(search_id string, api_key string) ([]models.SearchResult, *models.HttpError) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v0/searches/%s/responses", config.SLSKD_ROOT, search_id), nil)
	if err != nil {
		return nil, models.HttpError_From("Internal server error", http.StatusInternalServerError)
	}
	req.Header.Add("X-API-Key", api_key)

	res, err := client.Do(req)
	if err != nil {
		return nil, models.HttpError_From("Internal server error", http.StatusInternalServerError)
	}
	defer res.Body.Close()

	respBody2, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, models.HttpError_From("Internal server error", http.StatusInternalServerError)
	}

	// parse into a slice of SearchResult
	var apiResults []models.SearchResult
	if err := json.Unmarshal(respBody2, &apiResults); err != nil {
		return nil, models.HttpError_From("Internal server error", http.StatusInternalServerError)
	}

	return apiResults, nil
}

// Newznab category for Audio. This also has to appear in the caps <categories>
// block before Lidarr will accept releases carrying it.
const CATEGORY_AUDIO = "3000"

// slskd answers per user: one SearchResult is everything a single user matched,
// across however many directories. What anyone actually downloads is an album,
// which on Soulseek is a directory, so each user's file list gets split into the
// directories those files live in. One user offering three albums becomes three
// releases.
type releaseCandidate struct {
	username    string
	directory   string
	files       []models.SearchResultFile
	totalSize   int64
	queueLength int
	freeSlot    bool
}

func groupByDirectory(results []models.SearchResult) []releaseCandidate {
	candidates := make([]releaseCandidate, 0)

	for _, result := range results {
		// Insertion order is tracked separately so that two identical searches
		// produce identically ordered output; map iteration order is randomised.
		order := make([]string, 0)
		directories := make(map[string]*releaseCandidate)

		for _, file := range result.Files {
			if file.IsLocked {
				continue
			}

			directory := slskdDirectory(file.Filename)

			candidate, ok := directories[directory]
			if !ok {
				candidate = &releaseCandidate{
					username:    result.Username,
					directory:   directory,
					files:       make([]models.SearchResultFile, 0),
					queueLength: result.QueueLength,
					freeSlot:    result.HasFreeUploadSlot,
				}

				directories[directory] = candidate
				order = append(order, directory)
			}

			candidate.files = append(candidate.files, file)
			candidate.totalSize += file.Size
		}

		for _, directory := range order {
			candidates = append(candidates, *directories[directory])
		}
	}

	return candidates
}

func buildItems(results []models.SearchResult, cache *models.Cache, apiKey string) []models.TorznabChannelItem {
	candidates := groupByDirectory(results)
	items := make([]models.TorznabChannelItem, 0, len(candidates))
	pubDate := time.Now().Format(time.RFC1123Z)

	for _, candidate := range candidates {
		// TODO: Run this through the album namer. Lidarr identifies a release by
		// parsing its title, so a raw Soulseek directory name matches nothing and
		// the release gets discarded.
		name := slskdBasename(candidate.directory)

		// The hash is the release's identity for the rest of the pipeline: it is
		// the guid, the search cache key, the id in the download URL, and later
		// the torrent hash reported back by /api/v2/torrents/info. They all have
		// to agree or the grab cannot be matched to a Soulseek user and path.
		hash := releaseHash(candidate.username, candidate.directory)

		cache.CacheMutex.Lock()
		cache.SearchCache[hash] = models.SearchCacheEntry{
			SearchedAt: time.Now(),
			Files:      candidate.files,
			Username:   candidate.username,
			Name:       name,
		}
		cache.CacheMutex.Unlock()

		downloadURL := fmt.Sprintf("%s/api?t=custom_download&id=%s&apikey=%s",
			config.QBITSLSKD_ROOT, hash, url.QueryEscape(apiKey))

		// Nothing here really seeds, but a release reporting zero seeders is
		// treated as unavailable and dropped, so every release gets at least one.
		// Peers with a free upload slot get two so they sort above queued ones.
		seeders := 1
		if candidate.freeSlot {
			seeders = 2
		}

		size := strconv.FormatInt(candidate.totalSize, 10)

		items = append(items, models.TorznabChannelItem{
			Title:       name,
			Guid:        models.TorznabGuid{Text: hash, IsPermaLink: "false"},
			Link:        downloadURL,
			PubDate:     pubDate,
			Category:    CATEGORY_AUDIO,
			Description: name,
			Enclosure: models.TorznabChannelItemEnclosure{
				URL:    downloadURL,
				Length: size,
				Type:   "application/x-bittorrent",
			},
			Attr: []models.TorznabChannelItemAttr{
				{Name: "category", Value: CATEGORY_AUDIO},
				{Name: "size", Value: size},
				{Name: "files", Value: strconv.Itoa(len(candidate.files))},
				{Name: "seeders", Value: strconv.Itoa(seeders)},
				{Name: "peers", Value: strconv.Itoa(candidate.queueLength + seeders)},
				// A Soulseek transfer can never seed, so stop Lidarr holding
				// completed items against ratio goals that will never be met.
				{Name: "downloadvolumefactor", Value: "0"},
				{Name: "uploadvolumefactor", Value: "0"},
				{Name: "minimumseedtime", Value: "1"},
			},
		})
	}

	return items
}

func releaseHash(username string, directory string) string {
	hasher := sha1.New()
	// hash.Hash promises that Write never returns an error.
	hasher.Write([]byte(username + directory))

	return fmt.Sprintf("%x", hasher.Sum(nil))
}

// slskd always separates shared paths with backslashes, whatever the host OS.
func slskdDirectory(filename string) string {
	index := strings.LastIndex(filename, "\\")
	if index < 0 {
		return ""
	}

	return filename[:index]
}

func slskdBasename(path string) string {
	index := strings.LastIndex(path, "\\")
	if index < 0 {
		return path
	}

	return path[index+1:]
}
