package models

import "time"

type Transate_QBTorrentInfo struct {
	AddedOn                  int     `json:"added_on"`
	AmountLeft               int     `json:"amount_left"`
	AutoTmm                  bool    `json:"auto_tmm"`     // Auto Torrent Management
	Availability             int     `json:"availability"` // number of seeders with the torrent, -1 unknown
	Category                 string  `json:"category"`
	Comment                  string  `json:"comment"`
	Completed                int     `json:"completed"`
	CompletionOn             int     `json:"completion_on"`
	ContentPath              string  `json:"content_path"`
	DlLimit                  int     `json:"dl_limit"`
	Dlspeed                  int     `json:"dlspeed"`
	DownloadPath             string  `json:"download_path"`
	Downloaded               int     `json:"downloaded"`
	DownloadedSession        int     `json:"downloaded_session"`
	Eta                      int     `json:"eta"`
	FLPiecePrio              bool    `json:"f_l_piece_prio"`
	ForceStart               bool    `json:"force_start"`
	HasMetadata              bool    `json:"has_metadata"`
	Hash                     string  `json:"hash"`
	InactiveSeedingTimeLimit int     `json:"inactive_seeding_time_limit"`
	InfohashV1               string  `json:"infohash_v1"`
	InfohashV2               string  `json:"infohash_v2"`
	LastActivity             int     `json:"last_activity"`
	MagnetURI                string  `json:"magnet_uri"`
	MaxInactiveSeedingTime   int     `json:"max_inactive_seeding_time"`
	MaxRatio                 int     `json:"max_ratio"`
	MaxSeedingTime           int     `json:"max_seeding_time"`
	Name                     string  `json:"name"`
	NumComplete              int     `json:"num_complete"`
	NumIncomplete            int     `json:"num_incomplete"`
	NumLeechs                int     `json:"num_leechs"`
	NumSeeds                 int     `json:"num_seeds"`
	Popularity               float64 `json:"popularity"`
	Priority                 int     `json:"priority"`
	Private                  bool    `json:"private"`
	Progress                 float32 `json:"progress"`
	Ratio                    float64 `json:"ratio"`
	RatioLimit               int     `json:"ratio_limit"`
	Reannounce               int     `json:"reannounce"`
	RootPath                 string  `json:"root_path"`
	SavePath                 string  `json:"save_path"`
	SeedingTime              int     `json:"seeding_time"`
	SeedingTimeLimit         int     `json:"seeding_time_limit"`
	SeenComplete             int     `json:"seen_complete"`
	SeqDl                    bool    `json:"seq_dl"`
	Size                     int     `json:"size"`
	State                    string  `json:"state"`
	SuperSeeding             bool    `json:"super_seeding"`
	Tags                     string  `json:"tags"`
	TimeActive               int     `json:"time_active"`
	TotalSize                int     `json:"total_size"`
	Tracker                  string  `json:"tracker"`
	TrackersCount            int     `json:"trackers_count"`
	UpLimit                  int     `json:"up_limit"`
	Uploaded                 int     `json:"uploaded"`
	UploadedSession          int     `json:"uploaded_session"`
	Upspeed                  int     `json:"upspeed"`
}

// Safely calculate progress as a float32 between 0 and 1
func Translate_Progress(completed, total int) float32 {
	if total <= 0 {
		return 0
	}
	p := float64(completed) / float64(total)
	if p < 0 {
		return 0
	}
	if p > 1 {
		return 1
	}
	return float32(p)
}

// Convert Torrent State to qBittorrent state string
func Translate_State(s Status, completed bool) string {
	switch s {
	case StatusError:
		return "error"
	case StatusPaused:
		if completed {
			return "pausedUP"
		}
		return "pausedDL"
	case StatusChecking:
		return "checkingDL"
	case StatusQueued:
		if completed {
			return "queuedUP"
		}
		return "queuedDL"
	case StatusSeeding:
		return "uploading"
	case StatusDownloading:
		return "downloading"
	case StatusCompleted:
		return "completed"
	default:
		if completed {
			return "uploading"
		}
		return "stalledDL"
	}
}

func QBTorrentInfoFromCore(t QbtTorrent) Transate_QBTorrentInfo {
	// epoch seconds, -1 if incomplete
	var completionOn int = -1
	var eta int = -1

	if t.Status == StatusCompleted {
		completionOn = int(t.CompletedAt.Unix())
	}

	// Basic eta calculation of time started vs download speed
	remainingBytes := t.Size - t.CompletedBytes
	if t.DownloadSpeed != 0 {
		eta = int(remainingBytes / t.DownloadSpeed)
	}

	return Transate_QBTorrentInfo{
		AddedOn:                  int(t.AddedAt.Unix()),
		AmountLeft:               int(t.Size - t.CompletedBytes),
		AutoTmm:                  false,
		Availability:             1,
		Category:                 t.Category,
		Comment:                  "",
		Completed:                int(t.CompletedBytes),
		CompletionOn:             completionOn,
		ContentPath:              t.SavePath, // TODO: Validate this
		DlLimit:                  -1,         // Download speed unlimited, TODO: Get download speed from slskd config.
		Dlspeed:                  int(t.DownloadSpeed),
		DownloadPath:             t.SavePath, // TODO: Validate this
		Downloaded:               int(t.CompletedBytes),
		DownloadedSession:        int(t.CompletedBytes),
		Eta:                      eta,
		FLPiecePrio:              false,
		ForceStart:               false,
		HasMetadata:              false,
		Hash:                     t.Hash,
		InactiveSeedingTimeLimit: -1, // TODO: I have no idea what this is
		InfohashV1:               t.Hash,
		InfohashV2:               "",
		LastActivity:             0, // slskd does not provide this to us so we have to pick something
		MagnetURI:                "",
		MaxInactiveSeedingTime:   -1,
		MaxRatio:                 -1,
		MaxSeedingTime:           -1,
		Name:                     t.Name,
		NumComplete:              1, // This shouldn't matter
		NumIncomplete:            0, // This shouldn't matter
		NumLeechs:                0, // This shouldn't matter
		NumSeeds:                 1, // This shouldn't matter
		Popularity:               1, // TODO: I have no idea what this is
		Priority:                 1, // I'm not sure what the values of this could be, other than -1 if queuing is disabled or the torrent is seeding
		Private:                  false,
		Progress:                 Translate_Progress(int(t.CompletedBytes), int(t.Size)),
		Ratio:                    0,
		RatioLimit:               -1,
		Reannounce:               0,
		RootPath:                 t.SavePath,
		SavePath:                 t.SavePath,
		SeedingTime:              0,
		SeedingTimeLimit:         -1,
		SeenComplete:             int(time.Now().Unix()),
		SeqDl:                    false,
		Size:                     int(t.Size),
		State:                    Translate_State(t.Status, remainingBytes == 0),
		SuperSeeding:             false,
		Tags:                     "",
		TimeActive:               int(time.Now().Unix() - t.AddedAt.Unix()),
		TotalSize:                int(t.Size), // This should technically include all other files in the directory (including unselected) but I don't care
		Tracker:                  "_",         // Can't be an empty string
		TrackersCount:            1,
		UpLimit:                  -1,
		Uploaded:                 0,
		UploadedSession:          0,
		Upspeed:                  0,
	}
}
