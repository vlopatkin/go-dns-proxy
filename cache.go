package main

import (
	"sync"
	"time"
)

type Element struct {
	Value interface{}
	Added time.Time
}

type CacheSlice struct {
	next *CacheSlice

	mutex      sync.RWMutex
	elements   map[string]Element
	lastUpdate time.Time
}

type Cache struct {
	m     sync.RWMutex
	slice *CacheSlice

	now        func() time.Time
	expiration time.Duration
}

const sliceSize = 1024

func NewCache(expiration time.Duration) Cache {
	now := time.Now
	return Cache{
		now:        now,
		slice:      newCacheSlice(now()),
		expiration: expiration,
	}
}

func (cache *Cache) Get(k string) (interface{}, bool) {
	if cache == nil {
		return nil, false
	}

	cache.m.RLock()
	slice := cache.slice
	cache.m.RUnlock()

	for ; slice != nil; slice = slice.next {
		element, found := slice.get(k)
		if found {
			if cache.expiration > 0 {
				if cache.now().Sub(element.Added) < cache.expiration {
					return element.Value, true
				}
			} else {
				return element.Value, true
			}
		}
	}

	return nil, false
}

func (cache *Cache) Set(k string, v interface{}) {
	if cache == nil {
		return
	}

	now := cache.now()

	cache.m.RLock()
	if cache.expiration > 0 {
		if cache.slice.len() >= sliceSize {
			cache.m.RUnlock()

			slice := newCacheSlice(cache.now())
			cache.m.Lock()
			var sliceToClean *CacheSlice = nil
			if cache.slice.len() >= sliceSize {
				slice.next = cache.slice
				cache.slice = slice
				sliceToClean = slice.next
			}

			cache.m.Unlock()

			if sliceToClean != nil {
				go sliceToClean.cleanup(now, cache.expiration)
			}

			cache.m.RLock()
		}
	}
	slice := cache.slice
	cache.m.RUnlock()

	slice.mutex.Lock()
	slice.elements[k] = Element{
		Value: v,
		Added: now,
	}
	slice.lastUpdate = now
	slice.mutex.Unlock()
}

func newCacheSlice(now time.Time) *CacheSlice {
	return &CacheSlice{elements: make(map[string]Element, sliceSize+5), lastUpdate: now}
}

func (slice *CacheSlice) get(key string) (Element, bool) {
	slice.mutex.RLock()
	element, found := slice.elements[key]
	slice.mutex.RUnlock()

	return element, found
}

func (slice *CacheSlice) len() int {
	if slice == nil {
		return 0
	}
	slice.mutex.RLock()
	result := len(slice.elements)
	slice.mutex.RUnlock()

	return result
}

// returns true if and only if the slice could be totally removed
func (slice *CacheSlice) cleanup(now time.Time, expiration time.Duration) bool {
	if slice == nil {
		return true
	}

	slice.mutex.RLock()
	nextIsEmpty := slice.next == nil || slice.next.cleanup(now, expiration)
	freeToClean := now.Sub(slice.lastUpdate) > expiration
	slice.mutex.RUnlock()

	slice.mutex.Lock()
	if nextIsEmpty && slice.next != nil {
		m := slice.next.mutex
		m.Lock()
		slice.next = slice.next.next
		m.Unlock()
	}
	if freeToClean {
		for k, _ := range slice.elements {
			delete(slice.elements, k)
		}
	}
	slice.mutex.Unlock()

	return freeToClean && nextIsEmpty
}
