package models

import "time"

type SearchCacheEntry struct {
	SearchedAt time.Time
	Files      []SearchResultFile
	Username   string
	Name       string

	// Lidarr tracks a grab by the infohash it computed from the .torrent we
	// served, not by the release hash we key searches under, so
	// /api/v2/torrents/info has to report this back or the download never
	// matches the queue item and the album details stay blank.
	InfoHash string

	// The category Lidarr passed to /api/v2/torrents/add. Lidarr only imports
	// from its own category, so it has to come back out unchanged.
	Category string
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
