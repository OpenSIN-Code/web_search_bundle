// Purpose: SQLite-based cache for search queries and video analysis results.
// Docs: internal/cache/cache.doc.md
package cache

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// Cache provides persistent query caching.
type Cache struct {
	db *sql.DB
}

// New opens or creates a cache at the given path.
func New(path string) (*Cache, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	if err := migrate(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &Cache{db: db}, nil
}

func migrate(db *sql.DB) error {
	schema := `
CREATE TABLE IF NOT EXISTS cache (
	key TEXT PRIMARY KEY,
	sources TEXT NOT NULL,
	payload TEXT NOT NULL,
	created_at INTEGER NOT NULL,
	ttl_seconds INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS video_cache (
	key TEXT PRIMARY KEY,
	url TEXT NOT NULL,
	payload TEXT NOT NULL,
	created_at INTEGER NOT NULL,
	ttl_seconds INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_cache_created ON cache(created_at);
CREATE INDEX IF NOT EXISTS idx_video_cache_created ON video_cache(created_at);
`
	_, err := db.Exec(schema)
	return err
}

// HashKey returns a deterministic cache key for a query + source list.
func HashKey(query string, sources []string) string {
	h := sha256.New()
	h.Write([]byte(query))
	for _, s := range sources {
		h.Write([]byte(s))
	}
	return hex.EncodeToString(h.Sum(nil))
}

// Set stores a result in the cache.
func (c *Cache) Set(key string, sources []string, payload interface{}, ttl time.Duration) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	sourcesJSON, err := json.Marshal(sources)
	if err != nil {
		return err
	}

	_, err = c.db.Exec(
		`INSERT INTO cache(key, sources, payload, created_at, ttl_seconds)
		VALUES(?, ?, ?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET
			sources = excluded.sources,
			payload = excluded.payload,
			created_at = excluded.created_at,
			ttl_seconds = excluded.ttl_seconds`,
		key, sourcesJSON, data, time.Now().Unix(), int(ttl.Seconds()),
	)
	return err
}

// Get retrieves a cached entry if it is still valid.
func (c *Cache) Get(key string) ([]byte, bool, error) {
	var payload string
	var createdAt int64
	var ttlSeconds int64

	row := c.db.QueryRow(
		`SELECT payload, created_at, ttl_seconds FROM cache WHERE key = ?`,
		key,
	)
	if err := row.Scan(&payload, &createdAt, &ttlSeconds); err != nil {
		if err == sql.ErrNoRows {
			return nil, false, nil
		}
		return nil, false, err
	}

	if time.Now().Unix() > createdAt+ttlSeconds {
		return nil, false, nil
	}

	return []byte(payload), true, nil
}

// SetVideo stores a video analysis result.
func (c *Cache) SetVideo(key, url string, payload interface{}, ttl time.Duration) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = c.db.Exec(
		`INSERT INTO video_cache(key, url, payload, created_at, ttl_seconds)
		VALUES(?, ?, ?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET
			url = excluded.url,
			payload = excluded.payload,
			created_at = excluded.created_at,
			ttl_seconds = excluded.ttl_seconds`,
		key, url, data, time.Now().Unix(), int(ttl.Seconds()),
	)
	return err
}

// GetVideo retrieves a cached video analysis.
func (c *Cache) GetVideo(key string) ([]byte, bool, error) {
	var payload string
	var createdAt int64
	var ttlSeconds int64

	row := c.db.QueryRow(
		`SELECT payload, created_at, ttl_seconds FROM video_cache WHERE key = ?`,
		key,
	)
	if err := row.Scan(&payload, &createdAt, &ttlSeconds); err != nil {
		if err == sql.ErrNoRows {
			return nil, false, nil
		}
		return nil, false, err
	}

	if time.Now().Unix() > createdAt+ttlSeconds {
		return nil, false, nil
	}

	return []byte(payload), true, nil
}

// Stats returns cache row counts.
func (c *Cache) Stats() (searchCount, videoCount int, err error) {
	row := c.db.QueryRow(`SELECT COUNT(*) FROM cache`)
	if err := row.Scan(&searchCount); err != nil {
		return 0, 0, err
	}
	row = c.db.QueryRow(`SELECT COUNT(*) FROM video_cache`)
	if err := row.Scan(&videoCount); err != nil {
		return 0, 0, err
	}
	return searchCount, videoCount, nil
}

// Clear removes all cached entries.
func (c *Cache) Clear() error {
	if _, err := c.db.Exec(`DELETE FROM cache`); err != nil {
		return err
	}
	_, err := c.db.Exec(`DELETE FROM video_cache`)
	return err
}

// Close closes the underlying database.
func (c *Cache) Close() error {
	return c.db.Close()
}

// Compact removes expired entries.
func (c *Cache) Compact() error {
	now := time.Now().Unix()
	if _, err := c.db.Exec(`DELETE FROM cache WHERE created_at + ttl_seconds < ?`, now); err != nil {
		return err
	}
	_, err := c.db.Exec(`DELETE FROM video_cache WHERE created_at + ttl_seconds < ?`, now)
	return err
}

// String returns a short status string for CLI display.
func (c *Cache) String() string {
	search, video, err := c.Stats()
	if err != nil {
		return fmt.Sprintf("cache: error (%v)", err)
	}
	return fmt.Sprintf("cache: %d search entries, %d video entries", search, video)
}
