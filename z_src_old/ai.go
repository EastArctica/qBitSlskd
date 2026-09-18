package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/EastArctica/qbitslskd/config"
	"google.golang.org/genai"
)

var systemInstruction string

func initPrompt() {
	data, err := os.ReadFile("prompt.md")
	if err != nil {
		fmt.Printf("failed to read prompt.md: %v", err)
	}

	systemInstruction = string(data)
}

func GetAlbumName(albumPath string) (string, error) {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  config.GEMINI_API_KEY,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return "", err
	}

	config := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(systemInstruction, "model"),
	}

	chat, err := client.Chats.Create(ctx, "gemini-1.5-flash-8b", config, nil)
	if err != nil {
		return "", err
	}

	msg, err := chat.SendMessage(ctx, *genai.NewPartFromText(albumPath))
	if err != nil {
		return "", err
	}

	albumName := strings.TrimSpace(msg.Text())

	fmt.Printf("Converted soulseek path:\n\t%s\n\t%s\n", albumPath, albumName)

	return albumName, nil
}
