package utils

import (
	"crypto/sha1"
	"fmt"
)

func Sha1Hash(str string) (string, error) {
	hasher := sha1.New()
	_, err := hasher.Write([]byte(str))
	if err != nil {
		return "", fmt.Errorf("failed to write to hasher: %w", err)
	}

	hashBytes := hasher.Sum(nil)
	hashHex := fmt.Sprintf("%x", hashBytes)

	return hashHex, nil
}

func ReleaseHash(username string, directory string) string {
	hasher := sha1.New()
	// hash.Hash promises that Write never returns an error.
	hasher.Write([]byte(username + directory))

	return fmt.Sprintf("%x", hasher.Sum(nil))
}
