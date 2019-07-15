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
	if cache == nil {
		return nil, false
	}

	cache.mutex.RLock()
	defer cache.mutex.RUnlock()

	element, found := cache.elements[k]
	if !found {
		return "", false
	}
	if cache.expiration > 0 {
		if time.Since(element.TimeAdded) > cache.expiration {
			return "", false
		}
	}

	return element.Value, true
}

func (cache *Cache) Set(k string, v interface{}) {
	if cache == nil {
		return
	}

	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	cache.elements[k] = Element{
		Value:     v,
		TimeAdded: time.Now(),
	}
}
