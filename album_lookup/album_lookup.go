package album_lookup

import (
	"regexp"
	"strings"

	"github.com/EastArctica/qbitslskd/models"
)

var lookup_functions []func(string) (*models.Album, error) = []func(string) (*models.Album, error){
	LookupDeezer, LookupDiscogs,
}

// Thank god for the garbage collector
// This takes in the directory slskd provides (ie @@pmzzx\Music\Electronic\VA-For_DJs_Only_Extended_Club_Mixes)
func LookupAlbum(directory string) *models.Album {
	var album_scores map[models.Album]int = make(map[models.Album]int)

	for _, lookup := range lookup_functions {
		album, err := lookup(directory)
		if err != nil {
			continue
		}

		score, exists := album_scores[*album]
		if !exists {
			album_scores[*album] = 1
		} else {
			album_scores[*album] = score + 1
		}
	}

	var best_match *models.Album = nil

	for key, value := range album_scores {
		if best_match == nil {
			best_match = &key
			continue
		}

		if album_scores[*best_match] < value {
			best_match = &key
		}
	}

	return best_match
}

var parenRe = regexp.MustCompile(`[\[(][^\])]*[\])]`)
var yearRe = regexp.MustCompile(`\b\((19|20)\d{2}\)\b`)
var wsRe = regexp.MustCompile(`\s+`)

func Clean(s string) string {
	s = parenRe.ReplaceAllString(s, " ")
	s = yearRe.ReplaceAllString(s, " ")
	return strings.TrimSpace(wsRe.ReplaceAllString(s, " "))
}

// Take the last 2 directory names
func SlskdBasename(path string) string {
	split_path := strings.Split(path, "\\")

	if len(split_path) <= 2 {
		return strings.Join(split_path, "\\")
	}

	last_two := split_path[len(split_path)-2:]

	return strings.Join(last_two, "\\")
}
