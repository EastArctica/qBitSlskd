package models

import "time"

type SearchCacheEntry struct {
	SearchedAt time.Time
	Files      []SearchResultFile
	Username   string
	Name       string
}

type SearchResultFile struct {
	Code      int    `json:"code"`
	Extension string `json:"extension"`
	Filename  string `json:"filename"`
	Size      int64  `json:"size"`
	IsLocked  bool   `json:"isLocked"`
}
