package torznab

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/EastArctica/qbitslskd/album_lookup"
	"github.com/EastArctica/qbitslskd/config"
	"github.com/EastArctica/qbitslskd/models"
	"github.com/google/uuid"
)

// Every call out to slskd goes through one client with a timeout, so a slskd
// that accepts a connection and then stalls cannot pin a handler forever.
var slskdClient = &http.Client{Timeout: 60 * time.Second}

const (
	// slskd caps the search itself at SearchRequest.SearchTimeout, so a search
	// still incomplete well past that is never going to finish and the poll
	// loop has to give up rather than spin.
	searchPollTimeout  = 60 * time.Second
	searchPollInterval = 1 * time.Second
)

func SearchHandler(w http.ResponseWriter, req *http.Request, cache *models.Cache) {
	startSearchResponse, http_err := initSearch(req)
	if http_err != nil {
		http.Error(w, http_err.Err, http_err.Code)
		return
	}

	// initSearch has already rejected the request if this is missing.
	apiKey := req.URL.Query().Get("apikey")

	search_response, http_err := blockUntilSearchComplete(startSearchResponse.ID, apiKey)
	if http_err != nil {
		http.Error(w, http_err.Err, http_err.Code)
		return
	}

	search_results, http_err := getSearchResults(search_response.ID, apiKey)
	if http_err != nil {
		http.Error(w, http_err.Err, http_err.Code)
		return
	}

	if config.DELETE_SEARCHES {
		deleteSearch(search_response.ID, apiKey)
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

	w.Header().Set("Content-Type", "application/xml")
	fmt.Fprint(w, xml.Header)
	fmt.Fprint(w, string(data))
}

func initSearch(req *http.Request) (*models.SearchResponse, *models.HttpError) {
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
	startSearchReq, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v0/searches", config.SLSKD_ROOT), bytes.NewBuffer(searchReqData))
	if err != nil {
		fmt.Printf("Failed to create search request: %s\n", err)
		return nil, models.HttpError_From("Internal Server Error", http.StatusInternalServerError)
	}

	startSearchReq.Header.Add("X-API-Key", apiKey)
	startSearchReq.Header.Add("Content-Type", "application/json")

	startSearchRes, err := slskdClient.Do(startSearchReq)
	if err != nil {
		// TODO: Actually check the error
		return nil, models.HttpError_From("Internal server error", http.StatusInternalServerError)
	}
	defer startSearchRes.Body.Close()

	if startSearchRes.StatusCode == 401 || startSearchRes.StatusCode == 403 {
		return nil, models.HttpError_From("Forbidden", http.StatusForbidden)
	}

	startSearchResData, err := io.ReadAll(startSearchRes.Body)
	if err != nil {
		fmt.Printf("Failed to read search response %s\n", err)
		return nil, models.HttpError_From("Internal Server Error", http.StatusInternalServerError)
	}

	var startSearchResponse models.SearchResponse
	if err := json.Unmarshal(startSearchResData, &startSearchResponse); err != nil {
		fmt.Printf("Failed to unmarshal search response %s\n", err)
		return nil, models.HttpError_From("Internal Server Error", http.StatusInternalServerError)
	}

	return &startSearchResponse, nil
}

func blockUntilSearchComplete(id string, apiKey string) (*models.SearchResponse, *models.HttpError) {
	statusURL := fmt.Sprintf("%s/api/v0/searches/%s", config.SLSKD_ROOT, id)

	ctx, cancel := context.WithTimeout(context.Background(), searchPollTimeout)
	defer cancel()

	for {
		// A request carrying a context cannot be replayed once that context is
		// done, so each poll builds its own.
		statusReq, err := http.NewRequestWithContext(ctx, "GET", statusURL, nil)
		if err != nil {
			return nil, models.HttpError_From("Internal server error", http.StatusInternalServerError)
		}

		// include API key header
		statusReq.Header.Add("X-API-Key", apiKey)

		statusRes, err := slskdClient.Do(statusReq)
		if err != nil {
			if ctx.Err() != nil {
				fmt.Printf("Timed out waiting for search %s to complete\n", id)
				return nil, models.HttpError_From("Gateway timeout", http.StatusGatewayTimeout)
			}

			return nil, models.HttpError_From("Internal server error", http.StatusInternalServerError)
		}

		if statusRes.StatusCode == 403 {
			statusRes.Body.Close()
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

		select {
		case <-ctx.Done():
			fmt.Printf("Timed out waiting for search %s to complete\n", id)
			return nil, models.HttpError_From("Gateway timeout", http.StatusGatewayTimeout)
		case <-time.After(searchPollInterval):
		}
	}
}

func getSearchResults(search_id string, api_key string) ([]models.SearchResult, *models.HttpError) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v0/searches/%s/responses", config.SLSKD_ROOT, search_id), nil)
	if err != nil {
		return nil, models.HttpError_From("Internal server error", http.StatusInternalServerError)
	}
	req.Header.Add("X-API-Key", api_key)

	res, err := slskdClient.Do(req)
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

// slskd holds on to every search until it is told not to, so drop it once the
// responses have been read. Best effort: a search that outlives its results
// wastes memory in slskd but is not worth failing the feed over.
func deleteSearch(search_id string, api_key string) {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/v0/searches/%s", config.SLSKD_ROOT, search_id), nil)
	if err != nil {
		fmt.Printf("Failed to build delete request for search %s: %s\n", search_id, err)
		return
	}
	req.Header.Add("X-API-Key", api_key)

	res, err := slskdClient.Do(req)
	if err != nil {
		fmt.Printf("Failed to delete search %s: %s\n", search_id, err)
		return
	}
	res.Body.Close()
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
			// A file sitting at the root of a share has no directory to name a
			// release after, and an item with an empty title is something Lidarr
			// can never match, so it is dropped rather than published blank.
			if directory == "" {
				continue
			}

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
	pubDate := time.Now().Format(time.RFC1123Z)

	items := make([]models.TorznabChannelItem, 0, len(candidates))
	var itemsMutex sync.Mutex
	var wg sync.WaitGroup

	for _, candidate := range candidates {
		wg.Go(func() {
			// Lidarr identifies a release by parsing its title, so a raw Soulseek
			// directory base_name matches nothing and the release gets discarded.
			// The required format for Lidarr is `Artist - Album (Year) [Quality]`
			// Quality is usually the file extension.

			album_ptr := album_lookup.LookupAlbum(candidate.directory)
			if album_ptr == nil {
				return
			}

			// Insert trust me bro
			album := *album_ptr

			file_ext := getLidarrCompatibleFileExtension(candidate)

			name := fmt.Sprintf("%s - %s (%s) [%s]", album.Artist, album.Title, album.Year, file_ext)

			// The hash is the release's identity for the rest of the pipeline: it is
			// the guid, the search cache key, the id in the download URL, and later
			// the torrent hash reported back by /api/v2/torrents/info. They all have
			// to agree or the grab cannot be matched to a Soulseek user and path.
			hash := releaseHash(candidate.username, candidate.directory)

			cache.Mutex.Lock()
			cache.Search[hash] = models.SearchCacheEntry{
				SearchedAt: time.Now(),
				Files:      candidate.files,
				Username:   candidate.username,
				Name:       name,
			}
			cache.Mutex.Unlock()

			downloadURL := fmt.Sprintf("%s/api?t=custom_download&id=%s&apikey=%s",
				config.QBITSLSKD_ROOT, hash, url.QueryEscape(apiKey))

			// Nothing here really seeds, but a release reporting zero seeders is
			// treated as unavailable and dropped, so every release gets at least one.
			// Peers with a free upload slot get two so they sort above queued ones.
			seeders := 0
			if candidate.freeSlot {
				seeders = 100
			}

			size := strconv.FormatInt(candidate.totalSize, 10)

			itemsMutex.Lock()
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
			itemsMutex.Unlock()
		})
	}

	wg.Wait()

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

func getLidarrCompatibleFileExtension(candidate releaseCandidate) string {
	file_split := strings.Split(candidate.files[0].Filename, ".")
	file_ext := strings.ToLower(file_split[len(file_split)-1])

	lidarr_quality := file_ext

	sample_file := candidate.files[0]

	vbr := false
	if sample_file.IsVariableBitRate != nil {
		vbr = *sample_file.IsVariableBitRate
	}

	switch file_ext {
	case "flac":
		if sample_file.BitDepth >= 24 {
			lidarr_quality = "flac 24bit"
		} else {
			lidarr_quality = "flac"
		}
	case "mp3":
		if vbr && sample_file.BitRate >= 255 {
			lidarr_quality = "mp3 vbr v0"
		} else if vbr {
			lidarr_quality = "mp3 vbr v2"
		} else {
			arr := []int{96, 128, 160, 192, 256, 320}
			lidarr_quality = fmt.Sprintf("mp3 %d", genericRoundTo(sample_file.BitRate, arr))
		}
	case "m4a", "m4b", "m4p", "mp4", "aac":
		if sample_file.BitDepth >= 24 {
			lidarr_quality = "alac 24bit"
		} else if sample_file.BitRate > 500 {
			lidarr_quality = "alac"
		} else if sample_file.BitRate >= 176 {
			arr := []int{192, 256, 320}
			lidarr_quality = fmt.Sprintf("aac %d", genericRoundTo(sample_file.BitRate, arr))
		} else {
			lidarr_quality = "aac"
		}
	case "ogg", "oga", "opus":
		arr := []int{160, 192, 224, 256, 320, 500}
		lidarr_quality = fmt.Sprintf("%s %d", file_ext, genericRoundTo(sample_file.BitRate, arr))
	case "ape":
		lidarr_quality = "ape"
	case "wv":
		lidarr_quality = "wavpack"
	case "wav":
		lidarr_quality = "wav"
	case "wma":
		lidarr_quality = "wma"
	}

	return lidarr_quality
}

func genericRoundTo(x int, valid_values []int) int {
	sorted_valid_values := sort.IntSlice(valid_values)

	for i, value := range sorted_valid_values {
		if x == value {
			return value
		}

		if x < value {
			continue
		}

		if value == sorted_valid_values[len(sorted_valid_values)-1] {
			return value
		}

		dist_to_lesser := x - value
		dist_to_greater := sorted_valid_values[i+1] - x

		if dist_to_lesser < dist_to_greater {
			return value
		}
		return sorted_valid_values[i+1]
	}

	return sorted_valid_values[0]
}
