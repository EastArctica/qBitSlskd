package models

import "time"

type SearchCacheEntry struct {
	SearchedAt time.Time
	Files      []SearchResultFile
	Username   string
	Name       string
}

type Album struct {
	Title  string `json:"title"`
	Artist string `json:"artist"`
	Year   string `json:"year"`
}

type AlbumLookupWrapper struct {
	Data []AlbumLookupResult `json:"data"`
}

type AlbumLookupResult struct {
	Id     int    `json:"id"`
	Title  string `json:"title"`
	Artist struct {
		Name string `json:"name"`
	} `json:"artist"`
}

type YearLookupResult struct {
	ReleaseDate string `json:"release_date"`
}
