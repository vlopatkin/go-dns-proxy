package main

import (
	"sync"
	"time"
)

type Element struct {
	Value     interface{}
	TimeAdded time.Time
}

type Cache struct {
	elements   map[string]Element
	mutex      sync.RWMutex
	expiration time.Duration
}

func InitCache(expiration time.Duration) Cache {
	return Cache{
		elements:   make(map[string]Element),
		expiration: expiration,
	}
}

func (cache *Cache) Get(k string) (interface{}, bool) {
	cache.mutex.RLock()

	element, found := cache.elements[k]
	if !found {
		cache.mutex.RUnlock()
		return "", false
	}
	if cache.expiration > 0 {
		if time.Since(element.TimeAdded) > cache.expiration {
			cache.mutex.RUnlock()
			return "", false
		}
	}

	cache.mutex.RUnlock()
	return element.Value, true
}

func (cache *Cache) Set(k string, v interface{}) {
	cache.mutex.Lock()

	cache.elements[k] = Element{
		Value:     v,
		TimeAdded: time.Now(),
	}

	cache.mutex.Unlock()
}
