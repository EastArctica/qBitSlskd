package models

import (
	"sync"
)

type Cache struct {
	CacheMutex  sync.Mutex
	SearchCache map[string]SearchCacheEntry
}
