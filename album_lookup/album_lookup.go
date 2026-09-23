package album_lookup

import (
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/EastArctica/qbitslskd/models"
)

var lookup_functions []func(string) (*models.Album, error) = []func(string) (*models.Album, error){
	LookupDeezer, LookupDiscogs,
}

// A vote for one album. The sources disagree on spelling far more often than
// they disagree on the release, so the map is keyed on the normalized form
// while album keeps the first original spelling to hand back to the caller.
type albumVote struct {
	album models.Album
	score int
}

// Thank god for the garbage collector
// This takes in the directory slskd provides (ie @@pmzzx\Music\Electronic\VA-For_DJs_Only_Extended_Club_Mixes)
func LookupAlbum(directory string) (*models.Album, int) {
	var album_votes map[models.Album]*albumVote = make(map[models.Album]*albumVote)

	for _, lookup := range lookup_functions {
		album, err := lookup(directory)
		if err != nil {
			continue
		}

		key := normalizeAlbum(*album)

		vote, exists := album_votes[key]
		if !exists {
			album_votes[key] = &albumVote{album: *album, score: 1}
		} else {
			vote.score++
		}
	}

	var best_match *albumVote = nil

	// fmt.Printf("Lookup Album: %s\n", directory)
	// fmt.Printf("    Searched As: %s\n", Clean(SlskdBasename(directory)))
	for _, candidate := range album_votes {
		// fmt.Printf("    %s - %s (%s) [%d]\n", candidate.album.Artist, candidate.album.Title, candidate.album.Year, candidate.score)

		if best_match == nil {
			best_match = candidate
			continue
		}

		if best_match.score < candidate.score {
			best_match = candidate
		}
	}

	if best_match == nil {
		return nil, 0
	}

	return &best_match.album, best_match.score
}

var parenRe = regexp.MustCompile(`[\[(][^\])]*[\])]`)
var yearRe = regexp.MustCompile(`\b\((19|20)\d{2}\)\b`)
var wsRe = regexp.MustCompile(`\s+`)
var symbolRe = regexp.MustCompile(`(\\|\/|\-)`)
var apostropheRe = regexp.MustCompile("['\u2018\u2019`]")
var punctRe = regexp.MustCompile(`[^\p{L}\p{N} ]+`)
var editionRe = regexp.MustCompile(`(?i)\s+-\s+(ep|single|lp|deluxe edition|remastered)$`)

// Collapse the spelling differences between sources so equivalent results vote
// together. Deezer says "Various Artists" where Discogs says "Various", Discogs
// disambiguates with "Daft Punk (2)" and sorts as "Beatles, The", and the two
// disagree freely on punctuation and case.
func normalizeAlbum(album models.Album) models.Album {
	return models.Album{
		Title:  normalizeTitle(album.Title),
		Artist: normalizeArtist(album.Artist),
		Year:   strings.TrimSpace(album.Year),
	}
}

func normalizeTitle(s string) string {
	return normalizeText(editionRe.ReplaceAllString(s, ""))
}

func normalizeArtist(s string) string {
	s = normalizeText(s)

	// "Beatles, The" (Discogs artists_sort) and "The Beatles" are one artist.
	s = strings.TrimSuffix(s, " the")
	s = strings.TrimPrefix(s, "the ")

	if s == "various artists" {
		s = "various"
	}

	return strings.TrimSpace(s)
}

func normalizeText(s string) string {
	s = Clean(s)
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "&", " and ")
	// Dropped rather than spaced, so "DJ's" and "DJs" agree.
	s = apostropheRe.ReplaceAllString(s, "")
	s = punctRe.ReplaceAllString(s, " ")
	return strings.TrimSpace(wsRe.ReplaceAllString(s, " "))
}

func Clean(s string) string {
	s = parenRe.ReplaceAllString(s, " ")
	s = yearRe.ReplaceAllString(s, " ")
	s = symbolRe.ReplaceAllString(s, " ")
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

func httpReqAndUnmarshal(method string, url string, body string, object any) error {
	var httpClient = &http.Client{Timeout: 60 * time.Second}

	var body_io_reader io.Reader

	if body != "" {
		body_io_reader = strings.NewReader(body)
	} else {
		body_io_reader = nil
	}

	req, err := http.NewRequest(method, url, body_io_reader)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "qBitSlskd/1.0 (github.com/EastArctica/qBitSlskd)")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	res_body, _ := io.ReadAll(res.Body)

	err = json.Unmarshal(res_body, object)
	return err
}
