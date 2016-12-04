package main

import (
	"sync"
	"time"
)

type Store struct {
	mu              sync.RWMutex
	store           map[string]*Item
	defaultDuration time.Duration
}

type Item struct {
	value      interface{}
	duration   time.Duration
	expiration int64
}

const (
	DefaultDuration time.Duration = -1
	NoExpiration    time.Duration = 0
)

func NewStore(defaultDuration time.Duration, cleanupInterval time.Duration) *Store {
	if defaultDuration == DefaultDuration {
		defaultDuration = NoExpiration
	}
	s := &Store{
		store:           make(map[string]*Item),
		defaultDuration: defaultDuration,
	}
	if cleanupInterval > 0 {
		go cleaner(s, cleanupInterval)
	}
	return s
}

func cleaner(s *Store, interval time.Duration) {
	tick := time.Tick(interval)
	for now := range tick {
		unixNanoNow := now.UnixNano()
		s.mu.Lock()
		for key, item := range s.store {
			if (item.expiration != 0) && (unixNanoNow >= item.expiration) {
				delete(s.store, key)
			}
		}
		s.mu.Unlock()
	}
}

func (s *Store) Lock() {
	s.mu.Lock()
}

func (s *Store) Unlock() {
	s.mu.Unlock()
}

func (s *Store) Set(key string, value interface{}, duration time.Duration) {
	if duration == DefaultDuration {
		duration = s.defaultDuration
	}
	var expiration int64 = 0
	if duration > 0 {
		expiration = time.Now().Add(duration).UnixNano()
	}
	s.store[key] = &Item{
		value:      value,
		duration:   duration,
		expiration: expiration,
	}
}

func (s *Store) SetLock(key string, value interface{}, duration time.Duration) {
	s.mu.Lock()
	s.Set(key, value, duration)
	s.mu.Unlock()
}

func (s *Store) Get(key string) (interface{}, bool) {
	item, found := s.store[key]
	if found {
		return item.value, true
	} else {
		return nil, false
	}
}

func (s *Store) GetLock(key string) (interface{}, bool) {
	s.mu.RLock()
	value, found := s.Get(key)
	s.mu.RUnlock()
	return value, found
}

func (s *Store) RefreshExpiration(key string) {
	item, found := s.store[key]
	if found && (item.duration > 0) {
		item.expiration = time.Now().Add(item.duration).UnixNano()
	}
}

func (s *Store) RefreshExpirationLock(key string) {
	s.mu.Lock()
	s.RefreshExpiration(key)
	s.mu.Unlock()
}

func (s *Store) Delete(key string) {
	delete(s.store, key)
}

func (s *Store) DeleteLock(key string) {
	s.mu.Lock()
	s.Delete(key)
	s.mu.Unlock()
}
