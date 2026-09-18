package models

import "time"

type SearchCacheEntry struct {
	SearchedAt time.Time
	Files      []SearchResultFile
	Username   string
	Name       string
}
