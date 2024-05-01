package cache

import (
	"dist-concurrency/pkg/event"
	"sync"
	"time"
)

type CacheItem struct {
	Event      *event.Event
	Mu         sync.Mutex
	lastAccess time.Time
}
