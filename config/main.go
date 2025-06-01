package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Defaults from slskd
// https://github.com/slskd/slskd/blob/master/config/slskd.example.yml
var INCOMPLETE_DIR string = "~"
var COMPLETE_DIR string = "~"

var DELETE_SEARCHES bool = true
var DOWNLOAD_AUDIO_ONLY bool = false
var PORT string = "3000"
var SLSKD_ROOT string
var QBITSLSKD_ROOT string
var GEMINI_API_KEY string

func Init() {
	godotenv.Load()

	incompleteDir, ok := os.LookupEnv("SLSKD_INCOMPLETE_DIR")
	if ok {
		INCOMPLETE_DIR = incompleteDir
	}

	completeDir, ok := os.LookupEnv("SLSKD_COMPLETE_DIR")
	if ok {
		COMPLETE_DIR = completeDir
	}

	port, ok := os.LookupEnv("PORT")
	if ok {
		PORT = port
	}

	deleteSearches, ok := os.LookupEnv("DELETE_SEARCHES")
	if ok {
		if deleteSearches == "true" {
			DELETE_SEARCHES = true
		} else if deleteSearches == "false" {
			DELETE_SEARCHES = false
		} else {
			log.Fatalf("DELETE_SEARCHES env var is set to '%s', not 'true' or 'false'\n", deleteSearches)
		}
	}

	audioOnly, ok := os.LookupEnv("DOWNLOAD_AUDIO_ONLY")
	if ok {
		if audioOnly == "true" {
			DOWNLOAD_AUDIO_ONLY = true
		} else if audioOnly == "false" {
			DOWNLOAD_AUDIO_ONLY = false
		} else {
			log.Fatalf("DOWNLOAD_AUDIO_ONLY env var is set to '%s', not 'true' or 'false'\n", audioOnly)
		}
	}

	// We specifically don't verify this exists because of containered environments
	slskdRoot, ok := os.LookupEnv("SLSKD_ROOT")
	if !ok {
		log.Fatal("SLSKD_ROOT env var must be set to the slskd domain. Ex. https://slskd.example.com or http://slskd.local:5030\n")
	}
	SLSKD_ROOT = slskdRoot

	// We specifically don't verify this exists because of containered environments
	qbitSlskdRoot, ok := os.LookupEnv("QBITSLSKD_ROOT")
	if !ok {
		log.Fatal("QBITSLSKD_ROOT env var must be set to the qbitslskd domain. Ex. https://qbitslskd.example.com or http://qbitslskd.local:3000 (this needs to be accessible by lidarr)\n")
	}
	QBITSLSKD_ROOT = qbitSlskdRoot

	geminiApiKey, ok := os.LookupEnv("GEMINI_API_KEY")
	if !ok {
		log.Fatal("GEMINI_API_KEY env var must be set to a valid google api key with access to the Generative Language API. This is used to convert the soulseek paths to album names.\n")
	}
	GEMINI_API_KEY = geminiApiKey
}
