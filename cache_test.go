package main

import (
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCache_Get(t *testing.T) {
	//arrange
	cache := NewCache(0)
	cache.Set("k0", "v")

	//act
	v0, found0 := cache.Get("k0")
	_, found1 := cache.Get("k1")

	//assert
	assert.True(t, found0)
	assert.Equal(t, "v", v0)
	assert.False(t, found1)
}

func TestCache_Get_Expire(t *testing.T) {
	//arrange
	cache := NewCache(10 * time.Millisecond)
	now := time.Now()
	cache.now = func() time.Time {
		return now
	}
	cache.Set("k0", "v")

	//act
	v0, found0 := cache.Get("k0")
	now = now.Add(time.Second)
	_, found1 := cache.Get("k0")

	//assert
	assert.True(t, found0)
	assert.Equal(t, "v", v0)
	assert.False(t, found1)
}

func TestCache_Get_SliceExpire(t *testing.T) {
	//arrange
	cache := NewCache(10 * time.Millisecond)
	now := time.Now()
	cache.now = func() time.Time {
		return now
	}
	for i := 0; i < sliceSize*5; i++ {
		cache.Set(strconv.Itoa(i), nil)
	}

	//act

	//assert
	for i := 0; i < sliceSize*5; i++ {
		_, found := cache.Get(strconv.Itoa(i))
		assert.True(t, found, "before expire key %d", i)
	}

	//act
	now = now.Add(time.Second)
	cache.Set("-1", nil)
	runtime.Gosched()

	//assert
	for i := 0; i < sliceSize*5; i++ {
		_, found := cache.Get(strconv.Itoa(i))
		assert.False(t, found, "before expire key %d", i)
	}

	sliceCount := 0
	for s := cache.slice; s != nil; s = s.next {
		sliceCount++
	}

	assert.Less(t, sliceCount, 3)
}
