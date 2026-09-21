package utils

import (
	"bytes"
	"fmt"

	"github.com/EastArctica/qbitslskd/slskd"
)

func FindAudioFile(files []slskd.SlskdDownloadsFiles) (*slskd.SlskdDownloadsFiles, error) {
	for _, file := range files {
		if IsAudioFile(file.Filename) {
			return &file, nil
		}
	}

	return nil, fmt.Errorf("no audio files found")
}

func IsAudioFile(path string) bool {
	for _, extension := range AUDIO_EXTENSIONS {
		if bytes.HasSuffix([]byte(path), []byte(extension)) {
			return true
		}
	}

	return false
}

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
