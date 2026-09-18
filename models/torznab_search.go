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
