package album_lookup

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/EastArctica/qbitslskd/models"
)

type discogsAlbumSearchWrapper struct {
	Results []discogsAlbumSearchResult `json:"results"`
}

type discogsAlbumSearchResult struct {
	ResourceUrl string `json:"resource_url"`
}

type discogsAlbumLookup struct {
	Artist   string `json:"artists_sort"`
	Title    string `json:"title"`
	Released string `json:"released"`
}

func LookupDiscogs(directory string) (*models.Album, error) {
	var httpClient = &http.Client{Timeout: 60 * time.Second}

	name := Clean(SlskdBasename(directory))

	q := url.Values{}
	q.Set("q", name)
	q.Set("type", "album")

	url := "https://api.discogs.com/database/search?" + q.Encode()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "qBitSlskd/1.0 (github.com/EastArctica/qBitSlskd)")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	var albumLookup discogsAlbumSearchWrapper
	_ = json.Unmarshal(body, &albumLookup)

	if len(albumLookup.Results) == 0 {
		return nil, fmt.Errorf("No album found")
	}

	year_req, err := http.NewRequest("GET", albumLookup.Results[0].ResourceUrl, nil)
	if err != nil {
		return nil, err
	}

	year_req.Header.Set("User-Agent", "qBitSlskd/1.0 (github.com/EastArctica/qBitSlskd)")
	year_req.Header.Set("Accept", "application/json")
	year_req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	year_res, err := httpClient.Do(year_req)
	if err != nil {
		return nil, err
	}

	defer year_res.Body.Close()
	year_body, _ := io.ReadAll(year_res.Body)

	var album discogsAlbumLookup
	_ = json.Unmarshal(year_body, &album)

	return &models.Album{
		Title:  album.Title,
		Artist: album.Artist,
		Year:   strings.Split(album.Released, "-")[0],
	}, nil
}
