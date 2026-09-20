package models

type TorrentFileInfoFile struct {
	Length int64    "length"
	Path   []string "path"
}

type TorrentFileInfo struct {
	Files       []TorrentFileInfoFile "files"
	Name        string                "name"
	PieceLength int                   "piece length"
	Pieces      []byte                "pieces"
	// 0 or 1
	Private int "private"
	// Custom
	Username string "username"
	CacheId  string "cache-id"
}

type TorrentFile struct {
	Announce     string          "announce"
	AnnounceList [][]string      "announce-list"
	Comment      string          "comment"
	CreatedBy    string          "created by"
	CreationDate int             "creation date"
	Info         TorrentFileInfo "info"
	UrlList      []string        "url-list"
}
