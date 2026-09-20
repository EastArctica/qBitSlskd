package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/EastArctica/qbitslskd/config"
	"github.com/EastArctica/qbitslskd/internal/core"
	"github.com/EastArctica/qbitslskd/internal/translate"
	"github.com/EastArctica/qbitslskd/slskd"
	"github.com/jackpal/bencode-go"
)

var AUDIO_EXTENSIONS = []string{
	".mp3",
	".aac",
	".ogg",
	".wma",
	".opus",
	".m4a",
	".mp2",
	".ac3",
	".eac3",
	".dts",
	".amr",
	".awb",
	".ra",
	".ram",
	".flac",
	".alac",
	".ape",
	".wv",
	".tta",
	".shn",
	".wav",
	".aiff",
	".aif",
	".pcm",
	".au",
	".bwf",
	".rf64",
	".mid",
	".midi",
	".kar",
	".rmi",
	".mod",
	".s3m",
	".xm",
	".it",
	".mtm",
	".umx",
	".cda",
	".dss",
	".ds2",
	".dvf",
	".msv",
	".gsm",
	".vox",
	".sln",
	".voc",
	".iff",
	".svx",
}

func isAudioFile(path string) bool {
	for _, extension := range AUDIO_EXTENSIONS {
		if bytes.HasSuffix([]byte(path), []byte(extension)) {
			return true
		}
	}

	return false
}

func findAudioFile(files []slskd.SlskdDownloadsFiles) (*slskd.SlskdDownloadsFiles, error) {
	for _, file := range files {
		if isAudioFile(file.Filename) {
			return &file, nil
		}
	}

	return nil, fmt.Errorf("no audio files found")
}

func sha1Hash(str string) (string, error) {
	hasher := sha1.New()
	_, err := hasher.Write([]byte(str))
	if err != nil {
		return "", fmt.Errorf("failed to write to hasher: %w", err)
	}

	hashBytes := hasher.Sum(nil)
	hashHex := fmt.Sprintf("%x", hashBytes)

	return hashHex, nil
}

func version(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "2.11.4")
}

func login(w http.ResponseWriter, req *http.Request) {
	// TODO: Support username + password (it's a little jank and not recommended afaik)
	// Read api key from basic auth password (ignore username)
	// TODO: This isn't supposed to be basic actually, it's supposed to be I think some type of form but lidarr is fine with it
	_, apiKey, ok := req.BasicAuth()
	if !ok {
		fmt.Fprint(w, "Fails.")
		return
	}

	// Validate auth
	client := &http.Client{}
	appReq, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v0/application", config.SLSKD_ROOT), nil)
	if err != nil {
		fmt.Fprint(w, "Fails.")
		return
	}

	appReq.Header.Add("X-API-Key", apiKey)
	appRes, err := client.Do(appReq)
	if err != nil {
		fmt.Fprint(w, "Fails.")
		return
	}

	defer appRes.Body.Close()
	// TODO: Read use `username=_&password=...`
	_, err = io.ReadAll(req.Body)
	if err != nil {
		fmt.Fprint(w, "Fails.")
		return
	}

	// Give lidarr a "SID" cookie (it's what qbit does)
	// TODO: Maybe b64 encode this
	w.Header().Add("Set-Cookie", fmt.Sprintf("SID=%s; HttpOnly; SameSite=Strict; Path=/", apiKey))
	w.Header().Set("response-type", "application/json")
	fmt.Fprintf(w, "Ok.")
}

// TODO: Maybe actually do something about this instead of just tossing my current preferences in
func preferences(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("response-type", "application/json")
	fmt.Fprintf(w, "{\"add_stopped_enabled\":false,\"add_to_top_of_queue\":false,\"add_trackers\":\"\",\"add_trackers_enabled\":false,\"add_trackers_from_url_enabled\":false,\"add_trackers_url\":\"\",\"add_trackers_url_list\":\"\",\"alt_dl_limit\":10240,\"alt_up_limit\":10240,\"alternative_webui_enabled\":false,\"alternative_webui_path\":\"\",\"announce_ip\":\"\",\"announce_port\":0,\"announce_to_all_tiers\":true,\"announce_to_all_trackers\":false,\"anonymous_mode\":false,\"app_instance_name\":\"\",\"async_io_threads\":10,\"auto_delete_mode\":0,\"auto_tmm_enabled\":false,\"autorun_enabled\":false,\"autorun_on_torrent_added_enabled\":false,\"autorun_on_torrent_added_program\":\"\",\"autorun_program\":\"\",\"banned_IPs\":\"\",\"bdecode_depth_limit\":100,\"bdecode_token_limit\":10000000,\"bittorrent_protocol\":0,\"block_peers_on_privileged_ports\":false,\"bypass_auth_subnet_whitelist\":\"\",\"bypass_auth_subnet_whitelist_enabled\":false,\"bypass_local_auth\":false,\"category_changed_tmm_enabled\":false,\"checking_memory_use\":512,\"confirm_torrent_deletion\":true,\"confirm_torrent_recheck\":true,\"connection_speed\":75,\"current_interface_address\":\"\",\"current_interface_name\":\"\",\"current_network_interface\":\"\",\"delete_torrent_content_files\":false,\"dht\":true,\"dht_bootstrap_nodes\":\"dht.libtorrent.org:25401, dht.transmissionbt.com:6881, router.bittorrent.com:6881\",\"disk_cache\":-1,\"disk_cache_ttl\":60,\"disk_io_read_mode\":1,\"disk_io_type\":1,\"disk_io_write_mode\":1,\"disk_queue_size\":268435456,\"dl_limit\":0,\"dont_count_slow_torrents\":true,\"dyndns_domain\":\"changeme.dyndns.org\",\"dyndns_enabled\":false,\"dyndns_password\":\"\",\"dyndns_service\":0,\"dyndns_username\":\"\",\"embedded_tracker_port\":9000,\"embedded_tracker_port_forwarding\":false,\"enable_coalesce_read_write\":false,\"enable_embedded_tracker\":false,\"enable_multi_connections_from_same_ip\":false,\"enable_piece_extent_affinity\":true,\"enable_upload_suggestions\":false,\"encryption\":0,\"excluded_file_names\":\"\",\"excluded_file_names_enabled\":false,\"export_dir\":\"\",\"export_dir_fin\":\"\",\"file_log_age\":1,\"file_log_age_type\":1,\"file_log_backup_enabled\":true,\"file_log_delete_old\":true,\"file_log_enabled\":true,\"file_log_max_size\":65,\"file_log_path\":\"/config/qBittorrent/logs\",\"file_pool_size\":100,\"hashing_threads\":2,\"i2p_address\":\"127.0.0.1\",\"i2p_enabled\":false,\"i2p_inbound_length\":3,\"i2p_inbound_quantity\":3,\"i2p_mixed_mode\":false,\"i2p_outbound_length\":3,\"i2p_outbound_quantity\":3,\"i2p_port\":7656,\"idn_support_enabled\":false,\"ignore_ssl_errors\":false,\"incomplete_files_ext\":true,\"ip_filter_enabled\":false,\"ip_filter_path\":\"\",\"ip_filter_trackers\":false,\"limit_lan_peers\":true,\"limit_tcp_overhead\":false,\"limit_utp_rate\":true,\"listen_port\":23977,\"locale\":\"en\",\"lsd\":true,\"mail_notification_auth_enabled\":true,\"mail_notification_email\":\"\",\"mail_notification_enabled\":false,\"mail_notification_password\":\"\",\"mail_notification_sender\":\"qBittorrent_notification@example.com\",\"mail_notification_smtp\":\"smtp.changeme.com\",\"mail_notification_ssl_enabled\":false,\"mail_notification_username\":\"\",\"mark_of_the_web\":true,\"max_active_checking_torrents\":1,\"max_active_downloads\":50,\"max_active_torrents\":100,\"max_active_uploads\":50,\"max_concurrent_http_announces\":50,\"max_connec\":1000,\"max_connec_per_torrent\":100,\"max_inactive_seeding_time\":-1,\"max_inactive_seeding_time_enabled\":false,\"max_ratio\":5,\"max_ratio_act\":0,\"max_ratio_enabled\":true,\"max_seeding_time\":-1,\"max_seeding_time_enabled\":false,\"max_uploads\":350,\"max_uploads_per_torrent\":20,\"memory_working_set_limit\":8192,\"merge_trackers\":false,\"outgoing_ports_max\":0,\"outgoing_ports_min\":0,\"peer_tos\":4,\"peer_turnover\":4,\"peer_turnover_cutoff\":90,\"peer_turnover_interval\":300,\"performance_warning\":true,\"pex\":true,\"preallocate_all\":true,\"proxy_auth_enabled\":false,\"proxy_bittorrent\":true,\"proxy_hostname_lookup\":false,\"proxy_ip\":\"\",\"proxy_misc\":true,\"proxy_password\":\"\",\"proxy_peer_connections\":false,\"proxy_port\":8080,\"proxy_rss\":true,\"proxy_type\":\"None\",\"proxy_username\":\"\",\"python_executable_path\":\"\",\"queueing_enabled\":false,\"random_port\":false,\"reannounce_when_address_changed\":false,\"recheck_completed_torrents\":false,\"refresh_interval\":1500,\"request_queue_size\":2000,\"resolve_peer_countries\":true,\"resume_data_storage_type\":\"Legacy\",\"rss_auto_downloading_enabled\":false,\"rss_download_repack_proper_episodes\":true,\"rss_fetch_delay\":2,\"rss_max_articles_per_feed\":50,\"rss_processing_enabled\":false,\"rss_refresh_interval\":30,\"rss_smart_episode_filters\":\"s(\\\\d+)e(\\\\d+)\\n(\\\\d+)x(\\\\d+)\\n(\\\\d{4}[.\\\\-]\\\\d{1,2}[.\\\\-]\\\\d{1,2})\\n(\\\\d{1,2}[.\\\\-]\\\\d{1,2}[.\\\\-]\\\\d{4})\",\"save_path\":\"/data/downloads\",\"save_path_changed_tmm_enabled\":false,\"save_resume_data_interval\":60,\"save_statistics_interval\":15,\"scan_dirs\":{},\"schedule_from_hour\":8,\"schedule_from_min\":0,\"schedule_to_hour\":20,\"schedule_to_min\":0,\"scheduler_days\":0,\"scheduler_enabled\":false,\"send_buffer_low_watermark\":50000,\"send_buffer_watermark\":100000,\"send_buffer_watermark_factor\":200,\"slow_torrent_dl_rate_threshold\":2,\"slow_torrent_inactive_timer\":60,\"slow_torrent_ul_rate_threshold\":2,\"socket_backlog_size\":30,\"socket_receive_buffer_size\":0,\"socket_send_buffer_size\":0,\"ssl_enabled\":false,\"ssl_listen_port\":2054,\"ssrf_mitigation\":true,\"status_bar_external_ip\":false,\"stop_tracker_timeout\":2,\"temp_path\":\"/data/downloads/incomplete\",\"temp_path_enabled\":true,\"torrent_changed_tmm_enabled\":true,\"torrent_content_layout\":\"Subfolder\",\"torrent_content_remove_option\":\"Delete\",\"torrent_file_size_limit\":104857600,\"torrent_stop_condition\":\"None\",\"up_limit\":0,\"upload_choking_algorithm\":1,\"upload_slots_behavior\":0,\"upnp\":false,\"upnp_lease_duration\":0,\"use_category_paths_in_manual_mode\":false,\"use_https\":false,\"use_subcategories\":false,\"use_unwanted_folder\":true,\"utp_tcp_mixed_mode\":0,\"validate_https_tracker_certificate\":true,\"web_ui_address\":\"*\",\"web_ui_ban_duration\":3600,\"web_ui_clickjacking_protection_enabled\":true,\"web_ui_csrf_protection_enabled\":true,\"web_ui_custom_http_headers\":\"\",\"web_ui_domain_list\":\"*\",\"web_ui_host_header_validation_enabled\":true,\"web_ui_https_cert_path\":\"\",\"web_ui_https_key_path\":\"\",\"web_ui_max_auth_fail_count\":5,\"web_ui_port\":8112,\"web_ui_reverse_proxies_list\":\"\",\"web_ui_reverse_proxy_enabled\":false,\"web_ui_secure_cookie_enabled\":true,\"web_ui_session_timeout\":3600,\"web_ui_upnp\":false,\"web_ui_use_custom_http_headers_enabled\":false,\"web_ui_username\":\"_\"}")
}

type Category struct {
	Name     string `json:"name"`
	SavePath string `json:"savePath"`
}

var categories map[string]Category = map[string]Category{
	"lidarr": {
		Name:     "lidarr",
		SavePath: "",
	},
}

func categoriesHandler(w http.ResponseWriter, req *http.Request) {
	data, err := json.Marshal(categories)
	if err != nil {
		// TODO: idk qbit says it returns "200 in all scenarios" but this is most definitely a failure
		fmt.Fprintf(w, "{}")
		return
	}

	w.Header().Set("response-type", "application/json")
	fmt.Fprint(w, string(data))
}

func createCategoryHandler(w http.ResponseWriter, req *http.Request) {
	req.ParseMultipartForm(10 << 10) // 10MB
	category := req.PostForm.Get("category")
	if category == "" {
		http.Error(w, "Category cannot be empty", 400)
		return
	}

	// Check if category already exists
	if _, ok := categories[category]; ok {
		http.Error(w, "Unable to create category", http.StatusConflict)
		return
	}

	// TODO: In qBit, the category has other restrictions. In this we'll ignore them.
	categories[category] = Category{
		Name:     category,
		SavePath: req.Form.Get("savePath"),
	}

	categoriesHandler(w, req)
}

var albumNameMutex sync.RWMutex
var albumNameCache = make(map[string]string)

func torrentsInfoHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Printf("Request to path %s\n", req.URL.Path)

	// There's a lot of scawwy parameters to this :3
	// TODO: filter, category, tag, sort, reverse, limit, offset, hashes

	sidCookie, err := req.Cookie("SID")
	if err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	apiKey := sidCookie.Value

	users, err := slskd.GetDownloads(apiKey)
	if err != nil {
		fmt.Printf("torrentsInfoHandler: Failed to get slskd downloads: %s\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var pathsToConvert []string
	for _, user := range users {
		for _, dir := range user.Directories {
			audioPath := dir.Directory

			audioFile, err := findAudioFile(dir.Files)
			if err == nil {
				audioPath = audioFile.Filename
			}

			albumNameMutex.RLock()
			_, ok := albumNameCache[audioPath]
			albumNameMutex.RUnlock()
			if !ok {
				pathsToConvert = append(pathsToConvert, audioPath)
			}
		}
	}

	var waitGroup sync.WaitGroup
	for _, path := range pathsToConvert {
		waitGroup.Add(1)
		go func(path string) {
			defer waitGroup.Done()

			name, err := GetAlbumName(path)
			if err != nil {
				fmt.Printf("Failed to get song title %s\n", err.Error())
				name = path
			}

			albumNameMutex.Lock()
			albumNameCache[path] = name
			albumNameMutex.Unlock()
		}(path)

	}
	waitGroup.Wait()

	// Convert downloads to torrents info
	// TODO: This doesn't need to be a slice, we can determine the size by summing the dirs
	var torrents = []translate.QBTorrentInfo{}

	for _, user := range users {
		for _, dir := range user.Directories {
			var firstAddedAt int64
			var totalBytes int
			var bytesRemaining int
			// epoch seconds, -1 if incomplete
			var completionTime *time.Time = nil
			var downloadSpeed int64 = 0
			var status core.Status = core.StatusDownloading

			audioPath := dir.Directory
			audioFile, err := findAudioFile(dir.Files)
			if err == nil {
				audioPath = audioFile.Filename
			}

			hash, err := sha1Hash(user.Username + dir.Directory)
			if err != nil {
				fmt.Printf("Failed to calculate sha1 hash for: %s\n", user.Username+dir.Directory)
				continue
			}

			for _, file := range dir.Files {
				totalBytes += file.Size
				bytesRemaining += file.BytesRemaining

				// This should get the latest downloading file's speed
				if file.AverageSpeed != 0 {
					downloadSpeed = int64(file.AverageSpeed)
				}

				// Find earliest enqueued file
				enqueuedTime, err := time.Parse(time.RFC3339, file.EnqueuedAt)
				if err == nil {
					enq := enqueuedTime.Unix()
					if firstAddedAt == 0 || enq < firstAddedAt {
						firstAddedAt = enq
					}
				}

				// Track the latest completion time for when all files are done
				if !file.EndedAt.IsZero() {
					if completionTime == nil || file.EndedAt.After(*completionTime) {
						completionTime = &file.EndedAt
					}
				}
			}

			// TODO: Implement more types of state
			// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#get-torrent-list
			if bytesRemaining == 0 {
				status = core.StatusCompleted
			}

			if Some(dir.Files, func(file slskd.SlskdDownloadsFiles) bool { return includes(file.State, "Errored") }) {
				status = core.StatusError
			}

			// Get the full content path
			// Note: This is not a bug, but more of a note. If slskd ever switches off always using backslashes for directories, this will break
			// TODO: Get the actual download directory from slskd instead of forcing the user to specify it
			splitPath := strings.Split(dir.Directory, "\\")
			path := splitPath[len(splitPath)-1]
			if bytesRemaining == 0 {
				path = config.COMPLETE_DIR + path
			} else {
				path = config.INCOMPLETE_DIR + path
			}

			// Lidarr will only import from it's own categories, this is what we need to do to make it work
			allCategories := ""
			for category := range categories {
				if allCategories == "" {
					allCategories += category
				} else {
					allCategories += "," + category
				}
			}

			albumNameMutex.RLock()
			albumName := albumNameCache[audioPath]
			albumNameMutex.RUnlock()

			t := core.Torrent{
				Hash:           hash,
				Name:           albumName,
				Size:           int64(totalBytes),
				CompletedBytes: int64(totalBytes - bytesRemaining),
				DownloadSpeed:  int64(downloadSpeed),
				Status:         status,
				AddedAt:        time.Unix(firstAddedAt, 0),
				CompletedAt:    completionTime,
				Category:       "",
				SavePath:       path,
				SourceUser:     user.Username,
			}

			torrent := translate.QBTorrentInfoFromCore(t)

			torrents = append(torrents, torrent)
		}
	}

	// Marshal torrents and send it back
	torrentsData, err := json.Marshal(torrents)
	if err != nil {
		fmt.Printf("torrentsInfoHandler: Failed to marshal torrents\n")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.Write(torrentsData)
}

// TODO: All failures should 200 with "Fails." as the body
func addTorrentHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	sidCookie, err := req.Cookie("SID")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	apiKey := sidCookie.Value

	req.ParseMultipartForm(100 << 10) // 100MB

	// We only support file upload directly, no url or any of that
	file, _, err := req.FormFile("torrents")
	if err != nil {
		http.Error(w, "torrents missing or invalid", http.StatusBadRequest)
		return
	}

	fileData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusUnsupportedMediaType)
		return
	}

	var torrentFile TorrentFile
	err = bencode.Unmarshal(bytes.NewReader(fileData), &torrentFile)
	if err != nil {
		fmt.Printf("Invalid file: %s\n", err)
		http.Error(w, "Invalid/non-torrent file supplied", http.StatusUnsupportedMediaType)
		return
	}

	cacheEntry, ok := searchCache[torrentFile.Info.CacheId]
	if !ok {
		http.Error(w, "id param does not map to a valid cache entry", http.StatusBadRequest)
		return
	}

	filesRaw, err := json.Marshal(cacheEntry.Files)
	if err != nil {
		fmt.Printf("addTorrentHandler: Failed to marshal cacheEntry.files: %s\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Download files
	client := &http.Client{}
	downloadReq, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v0/transfers/downloads/%s", config.SLSKD_ROOT, cacheEntry.Username), bytes.NewBuffer(filesRaw))
	if err != nil {
		fmt.Printf("addTorrentHandler: Failed to create download request: %s\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	downloadReq.Header.Add("X-API-Key", apiKey)
	downloadReq.Header.Add("Content-Type", "application/json")

	downloadRes, err := client.Do(downloadReq)
	if err != nil {
		fmt.Printf("addTorrentHandler: Failed to send download request: %s\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if downloadRes.StatusCode == http.StatusUnauthorized || downloadRes.StatusCode == http.StatusForbidden {
		http.Error(w, downloadRes.Status, downloadRes.StatusCode)
		return
	}

	if downloadRes.StatusCode == http.StatusCreated {
		fmt.Fprintf(w, "Ok.")
		return
	}

	resBody, err := io.ReadAll(downloadRes.Body)
	if err != nil {
		fmt.Printf("addTorrentHandler: Unable to read body from slskd download: %d\n", downloadRes.StatusCode)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	fmt.Printf("addTorrentHandler: Bad response from slskd download: %d - %s\n", downloadRes.StatusCode, string(resBody))
	http.Error(w, "Internal server error", http.StatusInternalServerError)
}

type DeleteDownloadRequest struct {
	Remove bool `json:"remove"`
}

func deleteTorrentHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != "DELETE" && req.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := req.ParseForm(); err != nil {
		fmt.Printf("deleteTorrentHandler: Failed to read form: %s\n", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if !req.Form.Has("hashes") {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// TODO: Check this, delete files from disk if true
	if !req.Form.Has("deleteFiles") {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	sidCookie, err := req.Cookie("SID")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	apiKey := sidCookie.Value

	hashesStr := req.Form.Get("hashes")
	hashes := strings.Split(hashesStr, "|")

	// TODO: multithread slskd deletions
	downloads, err := slskd.GetDownloads(apiKey)
	if err != nil {
		fmt.Printf("deleteTorrentHandler: Failed to get slskd downloads: %s\n", err)
		http.Error(w, "Internal Server Error", http.StatusUnauthorized)
		return
	}

	for _, hash := range hashes {
		var foundIds []string
		var username string

	DownloadsLoop:
		for _, user := range downloads {
			for _, dir := range user.Directories {
				newHash, err := sha1Hash(user.Username + dir.Directory)
				if err != nil {
					continue
				}

				if hash == newHash {
					username = user.Username
					for _, file := range dir.Files {
						foundIds = append(foundIds, file.ID)
					}

					break DownloadsLoop
				}
			}
		}

		for _, id := range foundIds {
			err := slskd.DeleteFile(apiKey, username, id, true)
			if err != nil {
				fmt.Printf("deleteTorrentHandler: Failed to create delete files: %s\n", err)
				continue
			}
		}

		// TODO: Delete from cache
	}

	// qBittorrent returns nothing so we don't either
}

func logAll() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Print request line
		fmt.Printf("%s %s %s\n", req.Method, req.URL.RequestURI(), req.Proto)
		// Print headers
		for name, values := range req.Header {
			for _, v := range values {
				fmt.Printf("%s: %s\n", name, v)
			}
		}
		fmt.Println()

		// Print body
		if req.Body != nil {
			bodyBytes, err := io.ReadAll(req.Body)
			if err != nil {
				fmt.Printf("Error reading body: %v\n", err)
			} else if len(bodyBytes) > 0 {
				fmt.Printf("%s\n", string(bodyBytes))
			}
			// Restore Body so handlers down‐stream (if any) can still read it
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}

		// Always return 404
		http.NotFoundHandler().ServeHTTP(w, req)
	})
}

func main() {
	config.Init()
	initPrompt()

	mux := http.NewServeMux()

	// TODO: validate SID somehow?
	mux.HandleFunc("/api/v2/app/webapiVersion", version)
	mux.HandleFunc("/api/v2/auth/login", login)
	mux.HandleFunc("/api/v2/app/preferences", preferences)

	// Categories are entirely fake as there is no similar thing in slskd, also we don't need them
	mux.HandleFunc("/api/v2/torrents/categories", categoriesHandler)
	mux.HandleFunc("/api/v2/torrents/createCategory", createCategoryHandler)

	mux.HandleFunc("/api/v2/torrents/info", torrentsInfoHandler)
	mux.HandleFunc("/api/v2/torrents/add", addTorrentHandler)
	mux.HandleFunc("/api/v2/torrents/delete", deleteTorrentHandler)

	// TODO: /api/v2/torrents/delete

	// Torznab
	mux.HandleFunc("/api", TorznabHandler)

	mux.Handle("/", logAll())

	fmt.Printf("qBitSlskd started on port %s!\n", config.PORT)
	http.ListenAndServe(fmt.Sprintf(":%s", config.PORT), mux)
}
