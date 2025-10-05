package core

import "time"

type Status int

const (
	StatusUnknown Status = iota
	StatusQueued
	StatusDownloading
	StatusSeeding
	StatusPaused
	StatusCompleted
	StatusChecking
	StatusError
)

type Torrent struct {
	Hash           string
	Name           string
	Size           int64 // total bytes
	CompletedBytes int64 // bytes done
	DownloadSpeed  int64 // bytes/sec

	Status Status

	AddedAt     time.Time
	CompletedAt *time.Time
	Category    string
	SavePath    string
	SourceUser  string // slsk username
}
