package models

import "time"

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

// The response from slskd after initiating the search
type SearchResponse struct {
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
