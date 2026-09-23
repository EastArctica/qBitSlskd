package album_lookup

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/EastArctica/qbitslskd/models"
)

type discogsAlbumSearchWrapper struct {
	Results []discogsAlbumSearchResult `json:"results"`
}

type discogsAlbumSearchResult struct {
	ResourceUrl string `json:"resource_url"`
}

type discogsReleaseLookup struct {
	Artist   string `json:"artists_sort"`
	Title    string `json:"title"`
	Released string `json:"released"`
}

type discogsMasterLookup struct {
	MainReleaseUrl string `json:"main_release_url"`
}

func LookupDiscogs(directory string) (*models.Album, error) {
	name := Clean(SlskdBasename(directory))

	q := url.Values{}
	q.Set("q", name)
	q.Set("type", "master")

	url := "https://api.discogs.com/database/search?" + q.Encode()

	var search_results discogsAlbumSearchWrapper
	httpReqAndUnmarshal("GET", url, "", &search_results)

	if len(search_results.Results) == 0 {
		return nil, fmt.Errorf("No results found")
	}

	var master_result discogsMasterLookup
	httpReqAndUnmarshal("GET", search_results.Results[0].ResourceUrl, "", &master_result)

	var release_info discogsReleaseLookup
	httpReqAndUnmarshal("GET", master_result.MainReleaseUrl, "", &release_info)

	return &models.Album{
		Title:  release_info.Title,
		Artist: release_info.Artist,
		Year:   strings.Split(release_info.Released, "-")[0],
	}, nil
}
