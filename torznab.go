package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/EastArctica/qbitslskd/config"
	"github.com/google/uuid"
	"github.com/jackpal/bencode-go"
)

func findAudioFileFromResults(files []SearchResultFile) (*SearchResultFile, error) {
	for _, file := range files {
		if isAudioFile(file.Filename) {
			return &file, nil
		}
	}

	return nil, fmt.Errorf("no audio files found")
}

type CapabilitiesServer struct {
	Text      string `xml:",chardata"`
	Version   string `xml:"version,attr"`
	Title     string `xml:"title,attr"`
	Strapline string `xml:"strapline,attr"`
	Email     string `xml:"email,attr"`
	URL       string `xml:"url,attr"`
	Image     string `xml:"image,attr"`
}

type CapabilitiesLimits struct {
	Text    string `xml:",chardata"`
	Max     string `xml:"max,attr"`
	Default string `xml:"default,attr"`
}

type CapabilitiesRetention struct {
	Text string `xml:",chardata"`
	Days string `xml:"days,attr,omitempty"`
}

type CapabilitiesRegistration struct {
	Text      string `xml:",chardata"`
	Available string `xml:"available,attr"`
	Open      string `xml:"open,attr"`
}

type SupportedSearchCapability struct {
	Text            string `xml:",chardata"`
	Available       string `xml:"available,attr"`
	SupportedParams string `xml:"supportedParams,attr"`
}

type CapabilitiesSearching struct {
	Text        string                    `xml:",chardata"`
	Search      SupportedSearchCapability `xml:"search"`
	TvSearch    SupportedSearchCapability `xml:"tv-search"`
	MovieSearch SupportedSearchCapability `xml:"movie-search"`
	AudioSearch SupportedSearchCapability `xml:"audio-search"`
	BookSearch  SupportedSearchCapability `xml:"book-search"`
}

type TorznabCategory struct {
	Text string `xml:",chardata"`
	ID   string `xml:"id,attr"`
	Name string `xml:"name,attr"`
}

type CategoryWithSubcat struct {
	Text   string            `xml:",chardata"`
	ID     string            `xml:"id,attr"`
	Name   string            `xml:"name,attr"`
	Subcat []TorznabCategory `xml:"subcat"`
}

type CapabilitiesCategories struct {
	Text     string               `xml:",chardata"`
	Category []CategoryWithSubcat `xml:"category"`
}

type CapabilitiesGroup struct {
	Text        string `xml:",chardata"`
	ID          string `xml:"id,attr"`
	Name        string `xml:"name,attr"`
	Description string `xml:"description,attr"`
	Lastupdate  string `xml:"lastupdate,attr"`
}

type CapabilitiesGroups struct {
	Text  string              `xml:",chardata"`
	Group []CapabilitiesGroup `xml:"group"`
}

type CapabilitiesGenre struct {
	Text       string `xml:",chardata"`
	ID         string `xml:"id,attr"`
	Categoryid string `xml:"categoryid,attr"`
	Name       string `xml:"name,attr"`
}

type CapabilitiesGenres struct {
	Text  string              `xml:",chardata"`
	Genre []CapabilitiesGenre `xml:"genre"`
}

type CapabilitiesTag struct {
	Text        string `xml:",chardata"`
	Name        string `xml:"name,attr"`
	Description string `xml:"description,attr"`
}

type CapabilitiesTags struct {
	Text string            `xml:",chardata"`
	Tag  []CapabilitiesTag `xml:"tag"`
}

type Capabilities struct {
	XMLName xml.Name           `xml:"caps"`
	Text    string             `xml:",chardata"`
	Server  CapabilitiesServer `xml:"server"`
	Limits  CapabilitiesLimits `xml:"limits"`
	// Omitted in torznab, newznab only
	Retention    *CapabilitiesRetention   `xml:"retention,omitempty"`
	Registration CapabilitiesRegistration `xml:"registration"`
	Searching    CapabilitiesSearching    `xml:"searching"`
	Categories   CapabilitiesCategories   `xml:"categories"`
	// No current application specified in torznab.
	Groups CapabilitiesGroups `xml:"groups"`
	Genres CapabilitiesGenres `xml:"genres"`
	Tags   CapabilitiesTags   `xml:"tags"`
}

func CapabilitiesHandler(w http.ResponseWriter, req *http.Request) {
	var capabilities Capabilities = Capabilities{
		Server: CapabilitiesServer{
			Version:   "1.0",
			Title:     "qBitSlskd",
			Strapline: "qBitSlskd",
			Email:     "example@example.com",
			// I think this is where the actual server is but we don't have much of a choice for that
			URL: "https://github.com/EastArctica/qBitSlskd",
			// slskd icon
			Image: "https://avatars.githubusercontent.com/u/76762370",
		},
		Limits: CapabilitiesLimits{
			Max:     "10000",
			Default: "250",
		},
		Registration: CapabilitiesRegistration{
			Available: "yes",
			Open:      "yes",
		},
		Searching: CapabilitiesSearching{
			Search: SupportedSearchCapability{
				Available:       "yes",
				SupportedParams: "q",
			},
			TvSearch: SupportedSearchCapability{
				Available:       "no",
				SupportedParams: "",
			},
			MovieSearch: SupportedSearchCapability{
				Available:       "no",
				SupportedParams: "",
			},
			AudioSearch: SupportedSearchCapability{
				Available:       "no",
				SupportedParams: "",
			},
			BookSearch: SupportedSearchCapability{
				Available:       "no",
				SupportedParams: "",
			},
		},
	}

	data, err := xml.Marshal(capabilities)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("response-type", "application/xml")
	fmt.Fprint(w, string(data))
}

type ChannelLink struct {
	Text string `xml:",chardata"`
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
}

type ChannelImage struct {
	Text        string `xml:",chardata"`
	URL         string `xml:"url"`
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
}

type ChannelResponse struct {
	Text    string `xml:",chardata"`
	Newznab string `xml:"newznab,attr"`
	Offset  string `xml:"offset,attr"`
	Total   string `xml:"total,attr"`
}

type Guid struct {
	Text        string `xml:",chardata"`
	IsPermaLink string `xml:"isPermaLink,attr"`
}

type ChannelItemEnclosure struct {
	Text   string `xml:",chardata"`
	URL    string `xml:"url,attr"`
	Length string `xml:"length,attr"`
	Type   string `xml:"type,attr"`
}

type ChannelItemAttr struct {
	Text  string `xml:",chardata"`
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

type ChannelItem struct {
	Text        string               `xml:",chardata"`
	Title       string               `xml:"title"`
	Guid        Guid                 `xml:"guid"`
	Link        string               `xml:"link"`
	Comments    string               `xml:"comments"`
	PubDate     string               `xml:"pubDate"`
	Category    string               `xml:"category"`
	Description string               `xml:"description"`
	Enclosure   ChannelItemEnclosure `xml:"enclosure"`
	Attr        []ChannelItemAttr    `xml:"attr"`
}

type SearchChannel struct {
	Text        string          `xml:",chardata"`
	Link        ChannelLink     `xml:"link"`
	Title       string          `xml:"title"`
	Description string          `xml:"description"`
	Language    string          `xml:"language"`
	WebMaster   string          `xml:"webMaster"`
	Category    string          `xml:"category"`
	Image       ChannelImage    `xml:"image"`
	Ttl         string          `xml:"ttl"`
	Response    ChannelResponse `xml:"response"`
	Item        []ChannelItem   `xml:"item"`
}

type SearchResults struct {
	XMLName xml.Name      `xml:"rss"`
	Text    string        `xml:",chardata"`
	Version string        `xml:"version,attr"`
	Atom    string        `xml:"atom,attr"`
	Torznab string        `xml:"torznab,attr"`
	Channel SearchChannel `xml:"channel"`
}

type SearchRequest struct {
	Id                       string `json:"id"`
	FileLimit                int    `json:"fileLimit"`
	FilterResponses          bool   `json:"filterResponses"`
	MaximumPeerQueueLength   int    `json:"maximumPeerQueueLength"`
	MinimumPeerUploadSpeed   int    `json:"minimumPeerUploadSpeed"`
	MinimumResponseFileCount int    `json:"minimumResponseFileCount"`
	ResponseLimit            int    `json:"responseLimit"`
	SearchText               string `json:"searchText"`
	SearchTimeout            int    `json:"searchTimeout"`
}

type SearchStateResponse struct {
	FileCount       int        `json:"fileCount"`
	ID              string     `json:"id"`
	IsComplete      bool       `json:"isComplete"`
	LockedFileCount int        `json:"lockedFileCount"`
	ResponseCount   int        `json:"responseCount"`
	Responses       []any      `json:"responses"`
	SearchText      string     `json:"searchText"`
	StartedAt       time.Time  `json:"startedAt"`
	State           string     `json:"state"`
	Token           int        `json:"token"`
	EndedAt         *time.Time `json:"endedAt"`
}

type SearchResultFile struct {
	Code      int    `json:"code"`
	Extension string `json:"extension"`
	Filename  string `json:"filename"`
	Size      int64  `json:"size"`
	IsLocked  bool   `json:"isLocked"`
}

type SearchResult struct {
	FileCount         int                `json:"fileCount"`
	Files             []SearchResultFile `json:"files"`
	HasFreeUploadSlot bool               `json:"hasFreeUploadSlot"`
	LockedFileCount   int                `json:"lockedFileCount"`
	LockedFiles       []SearchResultFile `json:"lockedFiles"`
	QueueLength       int                `json:"queueLength"`
	Token             int                `json:"token"`
	UploadSpeed       int                `json:"uploadSpeed"`
	Username          string             `json:"username"`
}

type SearchCacheEntry struct {
	SearchedAt time.Time
	Files      []SearchResultFile
	Username   string
	Name       string
}

var searchCache = make(map[string]SearchCacheEntry)

func SearchHandler(w http.ResponseWriter, req *http.Request) {
	// TODO: !!!! FILTER OUT OLD SEARCH CACHE ENTRIES

	// Should take cat, attrs, extended, offset, and limit as parameters
	// We can't really do cat, we can fake offset and limit (or just return all of it always 🙃)
	var items []ChannelItem

	searchId, err := uuid.NewUUID()
	if err != nil {
		fmt.Printf("Failed to generate new UUID: %s\n", err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
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
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Search text, it can't be empty and prowlarr tests it with an empty string
	searchText := req.URL.Query().Get("q")
	if searchText == "" {
		searchText = "mu"
	}

	fmt.Printf("Searching for %s\n", searchText)

	// Actually search :tear:
	searchReq := SearchRequest{
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
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Search with query
	client := &http.Client{}
	startSearchReq, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v0/searches", config.SLSKD_ROOT), bytes.NewBuffer(searchReqData))
	if err != nil {
		fmt.Printf("Failed to create search request: %s\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	startSearchReq.Header.Add("X-API-Key", apiKey)
	startSearchReq.Header.Add("Content-Type", "application/json")

	startSearchRes, err := client.Do(startSearchReq)
	if err != nil {
		// TODO: Actually check the error
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if startSearchRes.StatusCode == 401 || startSearchRes.StatusCode == 403 {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	startSearchResData, err := io.ReadAll(startSearchRes.Body)
	if err != nil {
		fmt.Printf("Failed to read search response %s\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var startSearch SearchStateResponse
	err = json.Unmarshal(startSearchResData, &startSearch)
	if err != nil {
		fmt.Printf("Failed to unmarshal search response code %d\n%s\n\n%s\n%s\n", startSearchRes.StatusCode, err, string(startSearchResData), string(searchReqData))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	for {
		// build status endpoint URL
		statusURL := fmt.Sprintf("%s/api/v0/searches/%s", config.SLSKD_ROOT, startSearch.ID)
		statusReq, err := http.NewRequest("GET", statusURL, nil)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		// include API key header
		statusReq.Header.Add("X-API-Key", apiKey)

		statusRes, err := client.Do(statusReq)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if statusRes.StatusCode == 403 {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		statusBody, err := io.ReadAll(statusRes.Body)
		statusRes.Body.Close()
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// parse into SearchStateResponse
		var state SearchStateResponse
		if err := json.Unmarshal(statusBody, &state); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// stop polling when complete
		if state.IsComplete {
			// fetch the actual search results
			respReq2, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v0/searches/%s/responses", config.SLSKD_ROOT, startSearch.ID), nil)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			respReq2.Header.Add("X-API-Key", apiKey)

			respRes2, err := client.Do(respReq2)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			defer respRes2.Body.Close()

			respBody2, err := io.ReadAll(respRes2.Body)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			// parse into a slice of SearchResult
			var apiResults []SearchResult
			if err := json.Unmarshal(respBody2, &apiResults); err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			// Delete old search results
			if DELETE_SEARCHES {
				respReq3, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/v0/searches/%s", config.SLSKD_ROOT, startSearch.ID), nil)
				if err != nil {
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}
				respReq3.Header.Add("X-API-Key", apiKey)

				respRes3, err := client.Do(respReq3)
				if err != nil {
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}
				respRes3.Body.Close()
			}

			// By default, apiResults is an array of users, each result is all the files that matched for that given user.
			// We want this to turn this into an array of directories instead as we're not really downloading more than 1 directory from any given user
			var resultDirs []SearchResult

			for _, result := range apiResults {
				// map of path to result
				var userDirs map[string]SearchResult = make(map[string]SearchResult)

				for _, file := range result.Files {
					if config.DOWNLOAD_AUDIO_ONLY && !isAudioFile(file.Filename) {
						continue
					}

					splitName := strings.Split(file.Filename, "\\")
					// Remove file name so we have the dir it's in
					splitName = splitName[:len(splitName)-1]

					existingResults, ok := userDirs[strings.Join(splitName, "\\")]
					if !ok {
						existingResults = SearchResult{
							FileCount:         0,
							Files:             make([]SearchResultFile, 0),
							HasFreeUploadSlot: result.HasFreeUploadSlot,
							LockedFileCount:   result.LockedFileCount,
							LockedFiles:       make([]SearchResultFile, 0),
							QueueLength:       result.QueueLength,
							Token:             result.Token,
							UploadSpeed:       result.UploadSpeed,
							Username:          result.Username,
						}
					}

					existingResults.FileCount++
					existingResults.Files = append(existingResults.Files, file)

					userDirs[strings.Join(splitName, "\\")] = existingResults
				}

				for _, userDir := range userDirs {
					resultDirs = append(resultDirs, userDir)
				}
			}

			// TODO: check if we're in an empty query and therefore have no
			var pathsToConvert []string
			for _, result := range resultDirs {
				if result.FileCount == 0 || result.LockedFileCount != 0 {
					continue
				}

				// If there are *any* locked files, immediately ignore it
				if len(result.LockedFiles) > 0 {
					continue
				}

				dirNameSplit := strings.Split(result.Files[0].Filename, "\\")
				if len(dirNameSplit) > 0 {
					dirNameSplit = dirNameSplit[:len(dirNameSplit)-1]
				}

				audioPath := strings.Join(dirNameSplit, "\\")
				audioFile, err := findAudioFileFromResults(result.Files)
				if err == nil {
					audioPath = audioFile.Filename
				}
				pathsToConvert = append(pathsToConvert, audioPath)
			}

			var waitGroup sync.WaitGroup
			semaphore := make(chan struct{}, 20)
			for _, path := range pathsToConvert {
				waitGroup.Add(1)
				go func(path string) {
					defer waitGroup.Done()

					// Aquire semaphore to only send 20 requests at a time
					semaphore <- struct{}{}
					defer func() { <-semaphore }() // Release at end

					name, err := GetAlbumName(path)
					if err != nil {
						fmt.Printf("Failed to get song title %s\n", err.Error())
						name = path
					}

					albumNameMutex.Lock()
					albumNameCache[path] = name
					albumNameMutex.Unlock()
				}(path)

			}
			waitGroup.Wait()

			// convert each SearchResult → ChannelItem
			for _, result := range resultDirs {
				if result.FileCount == 0 || result.LockedFileCount != 0 {
					continue
				}

				// If there are *any* locked files, immediately ignore it
				if len(result.LockedFiles) > 0 {
					continue
				}

				dirNameSplit := strings.Split(result.Files[0].Filename, "\\")
				if len(dirNameSplit) > 0 {
					dirNameSplit = dirNameSplit[:len(dirNameSplit)-1]
				}

				audioPath := strings.Join(dirNameSplit, "\\")
				audioFile, err := findAudioFileFromResults(result.Files)
				if err == nil {
					audioPath = audioFile.Filename
				}

				albumNameMutex.RLock()
				albumName := albumNameCache[audioPath]
				albumNameMutex.RUnlock()

				// Add this to the search cache
				resultHash, err := sha1Hash(result.Username + strings.Join(dirNameSplit, "\\"))
				searchCache[resultHash] = SearchCacheEntry{
					SearchedAt: time.Now(),
					Files:      result.Files,
					Username:   result.Username,
					Name:       albumName,
				}

				var fileSize int64 = 0
				for _, file := range result.Files {
					fileSize += file.Size
				}

				items = append(items, ChannelItem{
					Title: albumName,
					Guid:  Guid{Text: result.Files[0].Filename, IsPermaLink: "false"},
					// TODO: This should be the human link someone can "look" at the result from. We can't really do this... Cache it?
					Link:        fmt.Sprintf("%s/api?t=custom_download&id=%s&apikey=%s", config.QBITSLSKD_ROOT, resultHash, apiKey),
					PubDate:     time.Now().Format(time.RFC1123Z),
					Description: albumName,
					Enclosure: ChannelItemEnclosure{
						URL:    fmt.Sprintf("%s/api?t=custom_download&id=%s&apikey=%s", config.QBITSLSKD_ROOT, resultHash, apiKey),
						Length: fmt.Sprint(fileSize),
						Type:   "application/x-bittorrent",
					},
				})
			}
			break
		}

		// wait 1 second before next poll
		time.Sleep(1 * time.Second)
	}

	// TODO: !!! Delete search

	var searchResults = SearchResults{
		Version: "2.0",
		Atom:    "http://www.w3.org/2005/Atom",
		Torznab: "http://torznab.com/schemas/2015/feed",
		Channel: SearchChannel{
			Link: ChannelLink{
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
			Image: ChannelImage{
				URL:         "https://avatars.githubusercontent.com/u/76762370",
				Title:       "qBitSlskd",
				Link:        "https://github.com/EastArctica/qBitSlskd",
				Description: "slskd Logo",
			},
			// Cache time (minutes)
			Ttl: "30",
			Response: ChannelResponse{
				Newznab: "http://www.newznab.com/DTD/2010/feeds/attributes/",
				// TODO: Offset (Is this possible with slskd?)
				Offset: "0",
				Total:  string(len(items)),
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

	fmt.Printf("Finished search for %s\n", searchText)

	w.Header().Set("response-type", "application/xml")
	fmt.Fprint(w, string(data))
}

type TorrentFileInfoFile struct {
	Length int64    "length"
	Path   []string "path"
}

type TorrentFileInfo struct {
	Files       []TorrentFileInfoFile "files"
	Name        string                "name"
	PieceLength int                   "piece length"
	Pieces      []byte                "pieces"
	// 0 or 1
	Private int "private"
	// Custom
	Username string "username"
	CacheId  string "cache-id"
}

type TorrentFile struct {
	Announce     string          "announce"
	AnnounceList [][]string      "announce-list"
	Comment      string          "comment"
	CreatedBy    string          "created by"
	CreationDate int             "creation date"
	Info         TorrentFileInfo "info"
	UrlList      []string        "url-list"
}

func CustomDownloadHandler(w http.ResponseWriter, req *http.Request) {
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
	cacheEntry, ok := searchCache[cacheId]
	if !ok {
		http.Error(w, "id param does not map to a valid cache entry", http.StatusBadRequest)
		return
	}

	// Generate a fake sha1 hash
	hasher := sha1.New()
	_, err := hasher.Write([]byte(cacheEntry.Username + cacheEntry.Name + cacheEntry.SearchedAt.String()))
	if err != nil {
		fmt.Printf("CustomDownloadHandler: Failed to generate fake sha1 hash: %s\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	hashBytes := hasher.Sum(nil)

	var torrentFiles []TorrentFileInfoFile

	for _, file := range cacheEntry.Files {
		torrentFiles = append(torrentFiles, TorrentFileInfoFile{
			Length: file.Size,
			Path:   []string{file.Filename},
		})
	}

	torrentFile := TorrentFile{
		Announce:     "udp://example.com:1337/announce",
		AnnounceList: [][]string{{"udp://example.com:1337/announce"}},
		Comment:      "qBitSlskd fake torrent file",
		CreatedBy:    "qBitSlskd",
		CreationDate: int(time.Now().Unix()),
		Info: TorrentFileInfo{
			Files:       torrentFiles,
			Name:        cacheEntry.Name,
			PieceLength: 262144,
			Pieces:      hashBytes,
			Private:     0,
			Username:    cacheEntry.Username,
			CacheId:     cacheId,
		},
		UrlList: []string{},
	}
	err = bencode.Marshal(w, torrentFile)
	if err != nil {
		fmt.Printf("CustomDownloadHandler: Failed to marshal torrent file: %s\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func TorznabHandler(w http.ResponseWriter, req *http.Request) {
	if !req.URL.Query().Has("t") {
		http.Error(w, "Missing function parameter \"t\"", http.StatusBadRequest)
		return
	}

	functionType := req.URL.Query().Get("t")
	switch functionType {
	case "caps":
		CapabilitiesHandler(w, req)
		return
	case "search":
		SearchHandler(w, req)
		return
	case "custom_download":
		// Note: This is NOT a real torznab function and is custom to qBitSlskd
		CustomDownloadHandler(w, req)
		return
	}
}
