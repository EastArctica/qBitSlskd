package torznab

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/EastArctica/qbitslskd/config"
	"github.com/EastArctica/qbitslskd/models"
	"github.com/google/uuid"
)

func SearchHandler(w http.ResponseWriter, req *http.Request, cache *models.Cache) {
	startSearchResponseData, http_err := initSearch(req)
	if http_err != nil {
		http.Error(w, http_err.Err, http_err.Code)
		return
	}

	var startSearchResponse models.SearchResponse
	err := json.Unmarshal(startSearchResponseData, &startSearchResponse)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	apiKey := req.URL.Query().Get("apikey")
	if apiKey == "" {
		println("No api key")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var search_response *models.SearchResponse = nil

	search_response, http_err = blockUntilSearchComplete(startSearchResponse.ID, apiKey)
	if http_err != nil {
		http.Error(w, http_err.Err, http_err.Code)
		return
	}

	fmt.Println(search_response.ID)

	var search_results []models.SearchResult = make([]models.SearchResult, 0)

	search_results, http_err = getSearchResults(search_response.ID, apiKey)
	if http_err != nil {
		http.Error(w, http_err.Err, http_err.Code)
		return
	}

	for i := 0; i < len(search_results); i++ {
		// fmt.Println(search_results[i])
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Search complete"))
}

func initSearch(req *http.Request) ([]byte, *models.HttpError) {
	searchId, err := uuid.NewUUID()
	if err != nil {
		fmt.Printf("Failed to generate new UUID: %s\n", err.Error())
		return nil, models.HttpError_From("Internal server error", http.StatusInternalServerError)
	}

	responseLimit := 250
	if req.URL.Query().Has("limit") {
		queryLimit, err := strconv.Atoi(req.URL.Query().Get("limit"))
		if err == nil {
			responseLimit = queryLimit
		}
	}

	// Ensure we have an api key
	apiKey := req.URL.Query().Get("apikey")
	if apiKey == "" {
		println("No api key")
		return nil, models.HttpError_From("Unauthorized", http.StatusUnauthorized)
	}

	// Search text, it can't be empty and prowlarr tests it with an empty string
	searchText := req.URL.Query().Get("q")
	if searchText == "" {
		searchText = "mu"
	}

	// Actually search :tear:
	searchReq := models.SearchRequest{
		Id:              searchId.String(),
		FileLimit:       100000,
		FilterResponses: true,
		// TODO: Queue length env variable (actually env variable for all this)
		MaximumPeerQueueLength:   1500,
		MinimumPeerUploadSpeed:   0,
		MinimumResponseFileCount: 1,
		ResponseLimit:            responseLimit,
		SearchText:               searchText,
		// idk 30 second timeout?
		SearchTimeout: 30000,
	}

	searchReqData, err := json.Marshal(searchReq)
	if err != nil {
		fmt.Printf("Failed to marshal searchReq: %s\n", err)
		return nil, models.HttpError_From("Internal server error", http.StatusInternalServerError)
	}

	// Search with query
	client := &http.Client{}
	startSearchReq, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v0/searches", config.SLSKD_ROOT), bytes.NewBuffer(searchReqData))
	if err != nil {
		fmt.Printf("Failed to create search request: %s\n", err)
		return nil, models.HttpError_From("Internal Server Error", http.StatusInternalServerError)
	}

	startSearchReq.Header.Add("X-API-Key", apiKey)
	startSearchReq.Header.Add("Content-Type", "application/json")

	startSearchRes, err := client.Do(startSearchReq)
	if err != nil {
		// TODO: Actually check the error
		return nil, models.HttpError_From("Internal server error", http.StatusInternalServerError)
	}

	if startSearchRes.StatusCode == 401 || startSearchRes.StatusCode == 403 {
		return nil, models.HttpError_From("Forbidden", http.StatusForbidden)
	}

	startSearchResData, err := io.ReadAll(startSearchRes.Body)
	if err != nil {
		fmt.Printf("Failed to read search response %s\n", err)
		return nil, models.HttpError_From("Internal Server Error", http.StatusInternalServerError)
	}

	return startSearchResData, nil
}

func blockUntilSearchComplete(id string, apiKey string) (*models.SearchResponse, *models.HttpError) {
	client := &http.Client{}

	statusURL := fmt.Sprintf("%s/api/v0/searches/%s", config.SLSKD_ROOT, id)
	statusReq, err := http.NewRequest("GET", statusURL, nil)
	if err != nil {
		return nil, models.HttpError_From("Internal server error", http.StatusInternalServerError)
	}

	// include API key header
	statusReq.Header.Add("X-API-Key", apiKey)

	for {

		statusRes, err := client.Do(statusReq)
		if err != nil {
			return nil, models.HttpError_From("Internal server error", http.StatusInternalServerError)
		}

		if statusRes.StatusCode == 403 {
			return nil, models.HttpError_From("Forbidden", http.StatusForbidden)
		}

		statusBody, err := io.ReadAll(statusRes.Body)
		statusRes.Body.Close()
		if err != nil {
			return nil, models.HttpError_From("Internal server error", http.StatusInternalServerError)
		}

		// parse into SearchStateResponse
		var state models.SearchResponse
		if err := json.Unmarshal(statusBody, &state); err != nil {
			return nil, models.HttpError_From("Internal server error", http.StatusInternalServerError)
		}

		if state.IsComplete {
			return &state, nil
		}

		time.Sleep(1 * time.Second)
	}
}

func getSearchResults(search_id string, api_key string) ([]models.SearchResult, *models.HttpError) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v0/searches/%s/responses", config.SLSKD_ROOT, search_id), nil)
	if err != nil {
		return nil, models.HttpError_From("Internal server error", http.StatusInternalServerError)
	}
	req.Header.Add("X-API-Key", api_key)

	res, err := client.Do(req)
	if err != nil {
		return nil, models.HttpError_From("Internal server error", http.StatusInternalServerError)
	}
	defer res.Body.Close()

	respBody2, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, models.HttpError_From("Internal server error", http.StatusInternalServerError)
	}

	// parse into a slice of SearchResult
	var apiResults []models.SearchResult
	if err := json.Unmarshal(respBody2, &apiResults); err != nil {
		return nil, models.HttpError_From("Internal server error", http.StatusInternalServerError)
	}

	return apiResults, nil
}
