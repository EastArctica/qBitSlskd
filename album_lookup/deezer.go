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

type albumLookupWrapper struct {
	Data []albumLookupResult `json:"data"`
}

type albumLookupResult struct {
	Id     int    `json:"id"`
	Title  string `json:"title"`
	Artist struct {
		Name string `json:"name"`
	} `json:"artist"`
}

type yearLookupResult struct {
	ReleaseDate string `json:"release_date"`
}

func LookupDeezer(directory string) (*models.Album, error) {
	var httpClient = &http.Client{Timeout: 60 * time.Second}

	name := Clean(SlskdBasename(directory))

	q := url.Values{}
	q.Set("q", name)
	q.Set("limit", "1")

	url := "https://api.deezer.com/search/album?" + q.Encode()

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

	var albumLookup albumLookupWrapper
	_ = json.Unmarshal(body, &albumLookup)

	if len(albumLookup.Data) == 0 {
		return nil, fmt.Errorf("No album found")
	}

	year_req, err := http.NewRequest("GET", fmt.Sprintf("https://api.deezer.com/album/%d", albumLookup.Data[0].Id), nil)
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

	var year yearLookupResult
	_ = json.Unmarshal(year_body, &year)

	return &models.Album{
		Title:  albumLookup.Data[0].Title,
		Artist: albumLookup.Data[0].Artist.Name,
		Year:   strings.Split(year.ReleaseDate, "-")[0],
	}, nil
}
