// Copyright 2014 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by a BSD-style
// license found in the LICENSE file.

// Package memkv implements an in-memory key/value store.
package memkv

import (
	"errors"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

var ErrNotExist = errors.New("key does not exist")
var ErrNoMatch = errors.New("no keys match")

// A Store represents an in-memory key-value store safe for
// concurrent access.
type Store struct {
	sync.RWMutex
	m map[string]KVPair
}

// New creates and initializes a new Store.
func New() Store {
	return Store{m: make(map[string]KVPair)}
}

// Delete deletes the KVPair associated with key.
func (s Store) Del(key string) {
	s.Lock()
	delete(s.m, key)
	s.Unlock()
}

// Get gets the KVPair associated with key. If there is no KVPair
// associated with key, Get returns KVPair{}, ErrNotExist.
func (s Store) Get(key string) (KVPair, error) {
	s.RLock()
	kv, ok := s.m[key]
	s.RUnlock()
	if !ok {
		return kv, ErrNotExist
	}
	return kv, nil
}

// GetValue gets the value associated with key. If there are no values
// associated with key, GetValue returns "", ErrNotExist.
func (s Store) GetValue(key string) (string, error) {
	kv, err := s.Get(key)
	if err != nil {
		return "", err
	}
	return kv.Value, nil
}

// GetAll returns a KVPair for all nodes with keys matching pattern.
// The syntax of patterns is the same as in filepath.Match.
func (s Store) GetAll(pattern string) (KVPairs, error) {
	ks := make(KVPairs, 0)
	s.RLock()
	defer s.RUnlock()
	for _, kv := range s.m {
		m, err := filepath.Match(pattern, kv.Key)
		if err != nil {
			return nil, err
		}
		if m {
			ks = append(ks, kv)
		}
	}
	if len(ks) == 0 {
		return nil, ErrNoMatch
	}
	sort.Sort(ks)
	return ks, nil
}

func (s Store) GetAllValues(pattern string) ([]string, error) {
	vs := make([]string, 0)
	ks, err := s.GetAll(pattern)
	if err != nil {
		return vs, err
	}
	for _, kv := range ks {
		vs = append(vs, kv.Value)
	}
	return vs, nil
}

func (s Store) List(path string) []string {
	vs := make([]string, 0)
	m := make(map[string]bool)
	s.RLock()
	defer s.RUnlock()
	for _, kv := range s.m {
		if strings.HasPrefix(kv.Key, path) {
			strippedKey := strings.TrimPrefix(kv.Key, path)
			m[strings.SplitN(strippedKey[1:], "/", 2)[0]] = true
		}
	}
	for k := range m {
		vs = append(vs, k)
	}
	sort.Strings(vs)
	return vs
}

func (s Store) ListDir(path string) []string {
	vs := make([]string, 0)
	m := make(map[string]bool)
	s.RLock()
	defer s.RUnlock()
	for _, kv := range s.m {
		if strings.HasPrefix(kv.Key, path) {
			strippedKey := strings.TrimPrefix(kv.Key, path)
			items := strings.SplitN(strippedKey[1:], "/", 2)
			if len(items) < 2 {
				continue
			}
			m[items[0]] = true
		}
	}
	for k := range m {
		vs = append(vs, k)
	}
	sort.Strings(vs)
	return vs
}

// Set sets the KVPair entry associated with key to value.
func (s Store) Set(key string, value string) {
	s.Lock()
	s.m[key] = KVPair{key, value}
	s.Unlock()
}
