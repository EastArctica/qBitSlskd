package album_lookup

import (
	"fmt"
	"net/url"
	"strings"

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
	name := Clean(SlskdBasename(directory))

	q := url.Values{}
	q.Set("q", name)
	q.Set("limit", "1")

	url := "https://api.deezer.com/search/album?" + q.Encode()
	var albumLookup albumLookupWrapper

	httpReqAndUnmarshal("GET", url, "", &albumLookup)

	if len(albumLookup.Data) == 0 {
		return nil, fmt.Errorf("No album found")
	}

	var year yearLookupResult
	httpReqAndUnmarshal("GET", fmt.Sprintf("https://api.deezer.com/album/%d", albumLookup.Data[0].Id), "", &year)

	return &models.Album{
		Title:  albumLookup.Data[0].Title,
		Artist: albumLookup.Data[0].Artist.Name,
		Year:   strings.Split(year.ReleaseDate, "-")[0],
	}, nil
}
