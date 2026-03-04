package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Cache provides simple file-based caching
type Cache struct {
	dir string
	ttl time.Duration
}

// New creates a new cache instance
func New(dir string, ttl time.Duration) *Cache {
	return &Cache{
		dir: dir,
		ttl: ttl,
	}
}

// Set stores an item in the cache with a TTL
func (c *Cache) Set(key string, value interface{}, ttl time.Duration) error {
	path := c.filePath(key)

	// Ensure cache directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to encode cache value: %w", err)
	}

	entry := cacheEntry{
		Data:      data,
		CreatedAt: time.Now(),
		TTL:       int(ttl.Seconds()),
	}

	encoded, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to encode cache entry: %w", err)
	}

	if err := os.WriteFile(path, encoded, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// Get retrieves an item from the cache
func (c *Cache) Get(key string) (interface{}, bool) {
	path := c.filePath(key)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}

	var entry cacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, false
	}

	// Check if expired
	ttl := time.Duration(entry.TTL) * time.Second
	if time.Since(entry.CreatedAt) > ttl {
		_ = os.Remove(path)
		return nil, false
	}

	var value interface{}
	if err := json.Unmarshal(entry.Data, &value); err != nil {
		return entry.Data, true // Return raw data if unmarshal fails
	}

	return value, true
}

// Delete removes an item from the cache
func (c *Cache) Delete(key string) error {
	path := c.filePath(key)
	return os.Remove(path)
}

// filePath returns the file path for a cache key
func (c *Cache) filePath(key string) string {
	// Hash the key to create a safe filename
	hash := sha256.Sum256([]byte(key))
	hashStr := hex.EncodeToString(hash[:])
	// Use first 2 chars of hash as subdirectory to avoid too many files in one dir
	return filepath.Join(c.dir, hashStr[:2], hashStr)
}

// cacheEntry represents a cached item
type cacheEntry struct {
	Data      []byte    `json:"data"`
	CreatedAt time.Time `json:"created_at"`
	TTL       int       `json:"ttl"`
}
