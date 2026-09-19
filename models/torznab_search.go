package models

import (
	"encoding/xml"
	"time"
)

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

type TorznabSearchResults struct {
	XMLName xml.Name             `xml:"rss"`
	Text    string               `xml:",chardata"`
	Version string               `xml:"version,attr"`
	Atom    string               `xml:"atom,attr"`
	Torznab string               `xml:"torznab,attr"`
	Channel TorznabSearchChannel `xml:"channel"`
}

type TorznabSearchChannel struct {
	Text        string                 `xml:",chardata"`
	Link        TorznabChannelLink     `xml:"link"`
	Title       string                 `xml:"title"`
	Description string                 `xml:"description"`
	Language    string                 `xml:"language"`
	WebMaster   string                 `xml:"webMaster"`
	Category    string                 `xml:"category"`
	Image       TorznabChannelImage    `xml:"image"`
	Ttl         string                 `xml:"ttl"`
	Response    TorznabChannelResponse `xml:"response"`
	Item        []TorznabChannelItem   `xml:"item"`
}

type TorznabChannelLink struct {
	Text string `xml:",chardata"`
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
}

type TorznabChannelImage struct {
	Text        string `xml:",chardata"`
	URL         string `xml:"url"`
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
}

type TorznabChannelResponse struct {
	Text    string `xml:",chardata"`
	Newznab string `xml:"newznab,attr"`
	Offset  string `xml:"offset,attr"`
	Total   string `xml:"total,attr"`
}

type TorznabChannelItem struct {
	Text        string                      `xml:",chardata"`
	Title       string                      `xml:"title"`
	Guid        TorznabGuid                 `xml:"guid"`
	Link        string                      `xml:"link"`
	Comments    string                      `xml:"comments"`
	PubDate     string                      `xml:"pubDate"`
	Category    string                      `xml:"category"`
	Description string                      `xml:"description"`
	Enclosure   TorznabChannelItemEnclosure `xml:"enclosure"`
	Attr        []TorznabChannelItemAttr    `xml:"attr"`
}

type TorznabGuid struct {
	Text        string `xml:",chardata"`
	IsPermaLink string `xml:"isPermaLink,attr"`
}

type TorznabChannelItemEnclosure struct {
	Text   string `xml:",chardata"`
	URL    string `xml:"url,attr"`
	Length string `xml:"length,attr"`
	Type   string `xml:"type,attr"`
}

type TorznabChannelItemAttr struct {
	Text  string `xml:",chardata"`
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}
