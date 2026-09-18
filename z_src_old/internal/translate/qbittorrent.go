package translate

import (
	"time"

	"github.com/EastArctica/qbitslskd/internal/core"
)

// Safely calculate progress as a float32 between 0 and 1
func Progress(completed, total int) float32 {
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

// Convert core.Torrent State to qBittorrent state string
func State(s core.Status, completed bool) string {
	switch s {
	case core.StatusError:
		return "error"
	case core.StatusPaused:
		if completed {
			return "pausedUP"
		}
		return "pausedDL"
	case core.StatusChecking:
		return "checkingDL"
	case core.StatusQueued:
		if completed {
			return "queuedUP"
		}
		return "queuedDL"
	case core.StatusSeeding:
		return "uploading"
	case core.StatusDownloading:
		return "downloading"
	case core.StatusCompleted:
		return "completed"
	default:
		if completed {
			return "uploading"
		}
		return "stalledDL"
	}
}

func QBTorrentInfoFromCore(t core.Torrent) QBTorrentInfo {
	// epoch seconds, -1 if incomplete
	var completionOn int = -1
	var eta int = -1

	if t.Status == core.StatusCompleted {
		completionOn = int(t.CompletedAt.Unix())
	}

	// Basic eta calculation of time started vs download speed
	remainingBytes := t.Size - t.CompletedBytes
	if t.DownloadSpeed != 0 {
		eta = int(remainingBytes / t.DownloadSpeed)
	}

	return QBTorrentInfo{
		AddedOn:                  int(t.AddedAt.Unix()),
		AmountLeft:               int(t.Size - t.CompletedBytes),
		AutoTmm:                  false,
		Availability:             1,
		Category:                 "", // TODO: Lidarr only imports from it's own categories, we need to generate this value
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
		Progress:                 Progress(int(t.CompletedBytes), int(t.Size)),
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
		State:                    State(t.Status, remainingBytes == 0),
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
