package qbittorrent

import (
	"fmt"
	"io"
	"net/http"

	"github.com/EastArctica/qbitslskd/config"
)

// APIKey pulls the slskd api key out of a request. qBittorrent hands clients a
// SID cookie after /auth/login, but Lidarr skips login entirely and just sends
// HTTP basic auth on every request, so accept either.
func APIKey(req *http.Request) (string, bool) {
	if sidCookie, err := req.Cookie("SID"); err == nil && sidCookie.Value != "" {
		return sidCookie.Value, true
	}

	if _, apiKey, ok := req.BasicAuth(); ok && apiKey != "" {
		return apiKey, true
	}

	return "", false
}

func Version(w http.ResponseWriter) {
	fmt.Fprintf(w, "2.11.4")
}

func Login(w http.ResponseWriter, req *http.Request) {
	_, apiKey, ok := req.BasicAuth()
	if !ok {
		fmt.Fprint(w, "Fails.")
		return
	}

	// Validate the api key works
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
	_, err = io.ReadAll(req.Body)
	if err != nil {
		fmt.Fprint(w, "Fails.")
		return
	}

	// Give lidarr a "SID" cookie thats literally just the slskd api key (it's what qbit does, so we emulate it)
	w.Header().Add("Set-Cookie", fmt.Sprintf("SID=%s; HttpOnly; SameSite=Strict; Path=/", apiKey))
	w.Header().Set("response-type", "application/json")
	fmt.Fprintf(w, "Ok.")
}

// TODO: This is just my preferences, maybe improve this?
func Preferences(w http.ResponseWriter) {
	w.Header().Set("response-type", "application/json")
	fmt.Fprintf(w, "{\"add_stopped_enabled\":false,\"add_to_top_of_queue\":false,\"add_trackers\":\"\",\"add_trackers_enabled\":false,\"add_trackers_from_url_enabled\":false,\"add_trackers_url\":\"\",\"add_trackers_url_list\":\"\",\"alt_dl_limit\":10240,\"alt_up_limit\":10240,\"alternative_webui_enabled\":false,\"alternative_webui_path\":\"\",\"announce_ip\":\"\",\"announce_port\":0,\"announce_to_all_tiers\":true,\"announce_to_all_trackers\":false,\"anonymous_mode\":false,\"app_instance_name\":\"\",\"async_io_threads\":10,\"auto_delete_mode\":0,\"auto_tmm_enabled\":false,\"autorun_enabled\":false,\"autorun_on_torrent_added_enabled\":false,\"autorun_on_torrent_added_program\":\"\",\"autorun_program\":\"\",\"banned_IPs\":\"\",\"bdecode_depth_limit\":100,\"bdecode_token_limit\":10000000,\"bittorrent_protocol\":0,\"block_peers_on_privileged_ports\":false,\"bypass_auth_subnet_whitelist\":\"\",\"bypass_auth_subnet_whitelist_enabled\":false,\"bypass_local_auth\":false,\"category_changed_tmm_enabled\":false,\"checking_memory_use\":512,\"confirm_torrent_deletion\":true,\"confirm_torrent_recheck\":true,\"connection_speed\":75,\"current_interface_address\":\"\",\"current_interface_name\":\"\",\"current_network_interface\":\"\",\"delete_torrent_content_files\":false,\"dht\":true,\"dht_bootstrap_nodes\":\"dht.libtorrent.org:25401, dht.transmissionbt.com:6881, router.bittorrent.com:6881\",\"disk_cache\":-1,\"disk_cache_ttl\":60,\"disk_io_read_mode\":1,\"disk_io_type\":1,\"disk_io_write_mode\":1,\"disk_queue_size\":268435456,\"dl_limit\":0,\"dont_count_slow_torrents\":true,\"dyndns_domain\":\"changeme.dyndns.org\",\"dyndns_enabled\":false,\"dyndns_password\":\"\",\"dyndns_service\":0,\"dyndns_username\":\"\",\"embedded_tracker_port\":9000,\"embedded_tracker_port_forwarding\":false,\"enable_coalesce_read_write\":false,\"enable_embedded_tracker\":false,\"enable_multi_connections_from_same_ip\":false,\"enable_piece_extent_affinity\":true,\"enable_upload_suggestions\":false,\"encryption\":0,\"excluded_file_names\":\"\",\"excluded_file_names_enabled\":false,\"export_dir\":\"\",\"export_dir_fin\":\"\",\"file_log_age\":1,\"file_log_age_type\":1,\"file_log_backup_enabled\":true,\"file_log_delete_old\":true,\"file_log_enabled\":true,\"file_log_max_size\":65,\"file_log_path\":\"/config/qBittorrent/logs\",\"file_pool_size\":100,\"hashing_threads\":2,\"i2p_address\":\"127.0.0.1\",\"i2p_enabled\":false,\"i2p_inbound_length\":3,\"i2p_inbound_quantity\":3,\"i2p_mixed_mode\":false,\"i2p_outbound_length\":3,\"i2p_outbound_quantity\":3,\"i2p_port\":7656,\"idn_support_enabled\":false,\"ignore_ssl_errors\":false,\"incomplete_files_ext\":true,\"ip_filter_enabled\":false,\"ip_filter_path\":\"\",\"ip_filter_trackers\":false,\"limit_lan_peers\":true,\"limit_tcp_overhead\":false,\"limit_utp_rate\":true,\"listen_port\":23977,\"locale\":\"en\",\"lsd\":true,\"mail_notification_auth_enabled\":true,\"mail_notification_email\":\"\",\"mail_notification_enabled\":false,\"mail_notification_password\":\"\",\"mail_notification_sender\":\"qBittorrent_notification@example.com\",\"mail_notification_smtp\":\"smtp.changeme.com\",\"mail_notification_ssl_enabled\":false,\"mail_notification_username\":\"\",\"mark_of_the_web\":true,\"max_active_checking_torrents\":1,\"max_active_downloads\":50,\"max_active_torrents\":100,\"max_active_uploads\":50,\"max_concurrent_http_announces\":50,\"max_connec\":1000,\"max_connec_per_torrent\":100,\"max_inactive_seeding_time\":-1,\"max_inactive_seeding_time_enabled\":false,\"max_ratio\":5,\"max_ratio_act\":0,\"max_ratio_enabled\":true,\"max_seeding_time\":-1,\"max_seeding_time_enabled\":false,\"max_uploads\":350,\"max_uploads_per_torrent\":20,\"memory_working_set_limit\":8192,\"merge_trackers\":false,\"outgoing_ports_max\":0,\"outgoing_ports_min\":0,\"peer_tos\":4,\"peer_turnover\":4,\"peer_turnover_cutoff\":90,\"peer_turnover_interval\":300,\"performance_warning\":true,\"pex\":true,\"preallocate_all\":true,\"proxy_auth_enabled\":false,\"proxy_bittorrent\":true,\"proxy_hostname_lookup\":false,\"proxy_ip\":\"\",\"proxy_misc\":true,\"proxy_password\":\"\",\"proxy_peer_connections\":false,\"proxy_port\":8080,\"proxy_rss\":true,\"proxy_type\":\"None\",\"proxy_username\":\"\",\"python_executable_path\":\"\",\"queueing_enabled\":false,\"random_port\":false,\"reannounce_when_address_changed\":false,\"recheck_completed_torrents\":false,\"refresh_interval\":1500,\"request_queue_size\":2000,\"resolve_peer_countries\":true,\"resume_data_storage_type\":\"Legacy\",\"rss_auto_downloading_enabled\":false,\"rss_download_repack_proper_episodes\":true,\"rss_fetch_delay\":2,\"rss_max_articles_per_feed\":50,\"rss_processing_enabled\":false,\"rss_refresh_interval\":30,\"rss_smart_episode_filters\":\"s(\\\\d+)e(\\\\d+)\\n(\\\\d+)x(\\\\d+)\\n(\\\\d{4}[.\\\\-]\\\\d{1,2}[.\\\\-]\\\\d{1,2})\\n(\\\\d{1,2}[.\\\\-]\\\\d{1,2}[.\\\\-]\\\\d{4})\",\"save_path\":\"/data/downloads\",\"save_path_changed_tmm_enabled\":false,\"save_resume_data_interval\":60,\"save_statistics_interval\":15,\"scan_dirs\":{},\"schedule_from_hour\":8,\"schedule_from_min\":0,\"schedule_to_hour\":20,\"schedule_to_min\":0,\"scheduler_days\":0,\"scheduler_enabled\":false,\"send_buffer_low_watermark\":50000,\"send_buffer_watermark\":100000,\"send_buffer_watermark_factor\":200,\"slow_torrent_dl_rate_threshold\":2,\"slow_torrent_inactive_timer\":60,\"slow_torrent_ul_rate_threshold\":2,\"socket_backlog_size\":30,\"socket_receive_buffer_size\":0,\"socket_send_buffer_size\":0,\"ssl_enabled\":false,\"ssl_listen_port\":2054,\"ssrf_mitigation\":true,\"status_bar_external_ip\":false,\"stop_tracker_timeout\":2,\"temp_path\":\"/data/downloads/incomplete\",\"temp_path_enabled\":true,\"torrent_changed_tmm_enabled\":true,\"torrent_content_layout\":\"Subfolder\",\"torrent_content_remove_option\":\"Delete\",\"torrent_file_size_limit\":104857600,\"torrent_stop_condition\":\"None\",\"up_limit\":0,\"upload_choking_algorithm\":1,\"upload_slots_behavior\":0,\"upnp\":false,\"upnp_lease_duration\":0,\"use_category_paths_in_manual_mode\":false,\"use_https\":false,\"use_subcategories\":false,\"use_unwanted_folder\":true,\"utp_tcp_mixed_mode\":0,\"validate_https_tracker_certificate\":true,\"web_ui_address\":\"*\",\"web_ui_ban_duration\":3600,\"web_ui_clickjacking_protection_enabled\":true,\"web_ui_csrf_protection_enabled\":true,\"web_ui_custom_http_headers\":\"\",\"web_ui_domain_list\":\"*\",\"web_ui_host_header_validation_enabled\":true,\"web_ui_https_cert_path\":\"\",\"web_ui_https_key_path\":\"\",\"web_ui_max_auth_fail_count\":5,\"web_ui_port\":8112,\"web_ui_reverse_proxies_list\":\"\",\"web_ui_reverse_proxy_enabled\":false,\"web_ui_secure_cookie_enabled\":true,\"web_ui_session_timeout\":3600,\"web_ui_upnp\":false,\"web_ui_use_custom_http_headers_enabled\":false,\"web_ui_username\":\"_\"}")
}
