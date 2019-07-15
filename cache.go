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

	for slice != nil {
		slice.mutex.RLock()
		element, found := slice.elements[k]
		next := slice.next
		slice.mutex.RUnlock()

		if found {
			if cache.expiration > 0 {
				if cache.now().Sub(element.Added) < cache.expiration {
					return element.Value, true
				}
				// older slices could not contain non expired entry for the key anyway
				return nil, false
			} else {
				return element.Value, true
			}
		}
		slice = next
	}

	return nil, false
}

func (cache *Cache) Set(k string, v interface{}) {
	if cache == nil {
		return
	}

	now := cache.now()

	cache.m.RLock()
	slice := cache.slice
	cache.m.RUnlock()

	if cache.expiration > 0 {
		if slice.len() >= sliceSize {
			newSlice := newCacheSlice(cache.now())

			cache.m.Lock()
			if slice == cache.slice && cache.slice.len() >= sliceSize {
				newSlice.next = cache.slice
				cache.slice = newSlice
				go slice.cleanup(now, cache.expiration)
				slice = cache.slice
			}
			cache.m.Unlock()
		}
	}

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
	nextSlice := slice.next
	isExpired := now.Sub(slice.lastUpdate) > expiration
	slice.mutex.RUnlock()

	nextIsEmpty := nextSlice == nil || nextSlice.cleanup(now, expiration)

	if (nextIsEmpty && nextSlice != nil) || isExpired {
		slice.mutex.Lock()
		if nextIsEmpty {
			slice.next = nil
		}
		if isExpired {
			for k, _ := range slice.elements {
				delete(slice.elements, k)
			}
		}
		slice.mutex.Unlock()
	}

	return isExpired && nextIsEmpty
}
