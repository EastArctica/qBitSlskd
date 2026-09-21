package models

import (
	"sync"
)

type Cache struct {
	Mutex      sync.Mutex
	Search     map[string]SearchCacheEntry
	Categories map[string]Category
}
