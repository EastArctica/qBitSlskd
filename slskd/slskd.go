package slskd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/EastArctica/qbitslskd/config"
)

func DeleteFile(apiKey string, username string, id string, remove bool) error {
	client := &http.Client{}

	deleteReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/v0/transfers/downloads/%s/%s?remove=%t", config.SLSKD_ROOT, username, id, remove), nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %s", err)
	}

	deleteReq.Header.Add("X-API-Key", apiKey)
	deleteReq.Header.Add("Content-Type", "application/json")

	downloadRes, err := client.Do(deleteReq)
	if err != nil {
		return fmt.Errorf("failed to send delete request: %s", err)
	}

	if downloadRes.StatusCode < 200 || downloadRes.StatusCode > 299 {
		return fmt.Errorf("delete request failed: %s", downloadRes.Status)
	}

	return nil
}

type SlskdDownloadsFiles struct {
	ID               string    `json:"id"`
	Username         string    `json:"username"`
	Direction        string    `json:"direction"`
	Filename         string    `json:"filename"`
	Size             int       `json:"size"`
	StartOffset      int       `json:"startOffset"`
	State            string    `json:"state"`
	RequestedAt      string    `json:"requestedAt"`
	EnqueuedAt       string    `json:"enqueuedAt"`
	StartedAt        time.Time `json:"startedAt"`
	EndedAt          time.Time `json:"endedAt"`
	BytesTransferred int       `json:"bytesTransferred"`
	AverageSpeed     float32   `json:"averageSpeed"`
	BytesRemaining   int       `json:"bytesRemaining"`
	ElapsedTime      string    `json:"elapsedTime"`
	PercentComplete  float32   `json:"percentComplete"`
	RemainingTime    *string   `json:"remainingTime"`
}

type SlskdDownloadUser struct {
	Username    string `json:"username"`
	Directories []struct {
		Directory string                `json:"directory"`
		FileCount int                   `json:"fileCount"`
		Files     []SlskdDownloadsFiles `json:"files"`
	} `json:"directories"`
}

func GetDownloads(apiKey string) ([]SlskdDownloadUser, error) {
	// Get downloads
	client := &http.Client{}
	slskdReq, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v0/transfers/downloads", config.SLSKD_ROOT), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create slskd downloads request: %s", err)
	}

	slskdReq.Header.Add("X-API-Key", apiKey)
	downloadsRes, err := client.Do(slskdReq)
	if err != nil {
		// TODO: Actually check the error
		return nil, fmt.Errorf("failed to create slskd downloads request: %s", err)
	}

	downloadsData, err := io.ReadAll(downloadsRes.Body)
	if err != nil {
		fmt.Printf("torrentsInfoHandler: Failed to read downloads: %s\n", err)
		return nil, fmt.Errorf("failed to create slskd downloads request: %s", err)
	}

	var users []SlskdDownloadUser
	err = json.Unmarshal(downloadsData, &users)
	if err != nil {
		fmt.Printf("torrentsInfoHandler: Failed to unmarshal downloads: %d - %s - %s\n", downloadsRes.StatusCode, err, downloadsData)
		return nil, fmt.Errorf("failed to create slskd downloads request: %s", err)
	}

	return users, nil
}
