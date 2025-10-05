package translate

type QBTorrentInfo struct {
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
