package objcache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Meta is stored beside each cached blob.
type Meta struct {
	ProfileID    string    `json:"profileId"`
	Bucket       string    `json:"bucket"`
	Key          string    `json:"key"`
	ETag         string    `json:"etag"`
	LastModified string    `json:"lastModified"`
	ContentType  string    `json:"contentType"`
	Size         int64     `json:"size"`
	CachedAt     time.Time `json:"cachedAt"`
}

type Store struct {
	dir        string
	maxBytes   int64
	maxFile    int64
	mu         sync.Mutex
}

func DefaultDir() (string, error) {
	base := os.TempDir()
	dir := filepath.Join(base, "s3store-cache")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

// New creates a cache under dir.
// maxFile is the largest single object to cache (0 = 64MiB default).
// maxBytes is approximate total cache budget (0 = 2GiB default).
func New(dir string, maxFile, maxBytes int64) (*Store, error) {
	if dir == "" {
		var err error
		dir, err = DefaultDir()
		if err != nil {
			return nil, err
		}
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	if maxFile <= 0 {
		maxFile = 64 << 20
	}
	if maxBytes <= 0 {
		maxBytes = 2 << 30
	}
	s := &Store{dir: dir, maxFile: maxFile, maxBytes: maxBytes}
	go s.pruneOnce()
	return s, nil
}

func (s *Store) Dir() string { return s.dir }

func (s *Store) MaxFile() int64 { return s.maxFile }

func (s *Store) hash(profileID, bucket, key string) string {
	h := sha256.Sum256([]byte(profileID + "\n" + bucket + "\n" + key))
	return hex.EncodeToString(h[:])
}

func (s *Store) paths(id string) (dataPath, metaPath string) {
	return filepath.Join(s.dir, id+".bin"), filepath.Join(s.dir, id+".json")
}

// Lookup returns a readable file if cache is fresh for the given validators.
func (s *Store) Lookup(profileID, bucket, key, etag, lastModified string) (path string, meta Meta, ok bool) {
	id := s.hash(profileID, bucket, key)
	dataPath, metaPath := s.paths(id)
	b, err := os.ReadFile(metaPath)
	if err != nil {
		return "", Meta{}, false
	}
	if err := json.Unmarshal(b, &meta); err != nil {
		return "", Meta{}, false
	}
	st, err := os.Stat(dataPath)
	if err != nil || st.Size() != meta.Size {
		return "", Meta{}, false
	}
	// Prefer ETag match; fall back to LastModified string equality.
	etag = strings.Trim(etag, `"`)
	metaETag := strings.Trim(meta.ETag, `"`)
	if etag != "" && metaETag != "" {
		if etag != metaETag {
			return "", Meta{}, false
		}
	} else if lastModified != "" && meta.LastModified != "" {
		if lastModified != meta.LastModified {
			return "", Meta{}, false
		}
	} else {
		// no validators — do not trust cache
		return "", Meta{}, false
	}
	return dataPath, meta, true
}

// Put writes reader to cache. size should be known when possible (-1 unknown).
func (s *Store) Put(profileID, bucket, key, etag, lastModified, contentType string, size int64, r io.Reader) (path string, err error) {
	if size >= 0 && size > s.maxFile {
		// still stream through without caching
		return "", fmt.Errorf("object too large to cache")
	}
	id := s.hash(profileID, bucket, key)
	dataPath, metaPath := s.paths(id)
	tmp := dataPath + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return "", err
	}
	// cap write if size unknown
	var written int64
	if size >= 0 {
		written, err = io.Copy(f, io.LimitReader(r, size+1))
		if err == nil && written > size {
			f.Close()
			os.Remove(tmp)
			return "", fmt.Errorf("size mismatch")
		}
	} else {
		lr := io.LimitReader(r, s.maxFile+1)
		written, err = io.Copy(f, lr)
		if err == nil && written > s.maxFile {
			f.Close()
			os.Remove(tmp)
			return "", fmt.Errorf("object too large to cache")
		}
	}
	if err != nil {
		f.Close()
		os.Remove(tmp)
		return "", err
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return "", err
	}
	meta := Meta{
		ProfileID:    profileID,
		Bucket:       bucket,
		Key:          key,
		ETag:         strings.Trim(etag, `"`),
		LastModified: lastModified,
		ContentType:  contentType,
		Size:         written,
		CachedAt:     time.Now().UTC(),
	}
	mb, _ := json.MarshalIndent(meta, "", "  ")
	if err := os.WriteFile(metaPath+".tmp", mb, 0o600); err != nil {
		os.Remove(tmp)
		return "", err
	}
	if err := os.Rename(tmp, dataPath); err != nil {
		os.Remove(tmp)
		os.Remove(metaPath + ".tmp")
		return "", err
	}
	_ = os.Rename(metaPath+".tmp", metaPath)
	go s.pruneOnce()
	return dataPath, nil
}

// Open opens a cached file for reading.
func (s *Store) Open(path string) (*os.File, error) {
	return os.Open(path)
}

func (s *Store) pruneOnce() {
	s.mu.Lock()
	defer s.mu.Unlock()
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return
	}
	type item struct {
		data string
		meta string
		at   time.Time
		size int64
	}
	var items []item
	var total int64
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".bin") {
			continue
		}
		id := strings.TrimSuffix(name, ".bin")
		dataPath, metaPath := s.paths(id)
		st, err := os.Stat(dataPath)
		if err != nil {
			continue
		}
		at := st.ModTime()
		var m Meta
		if b, err := os.ReadFile(metaPath); err == nil {
			_ = json.Unmarshal(b, &m)
			if !m.CachedAt.IsZero() {
				at = m.CachedAt
			}
		}
		// drop entries older than 14 days
		if time.Since(at) > 14*24*time.Hour {
			os.Remove(dataPath)
			os.Remove(metaPath)
			continue
		}
		items = append(items, item{data: dataPath, meta: metaPath, at: at, size: st.Size()})
		total += st.Size()
	}
	if total <= s.maxBytes {
		return
	}
	// delete oldest until under budget
	for total > s.maxBytes && len(items) > 0 {
		// find oldest
		oi := 0
		for i := 1; i < len(items); i++ {
			if items[i].at.Before(items[oi].at) {
				oi = i
			}
		}
		os.Remove(items[oi].data)
		os.Remove(items[oi].meta)
		total -= items[oi].size
		items = append(items[:oi], items[oi+1:]...)
	}
}


// Invalidate drops a single cached object if present.
func (s *Store) Invalidate(profileID, bucket, key string) {
	id := s.hash(profileID, bucket, key)
	dataPath, metaPath := s.paths(id)
	_ = os.Remove(dataPath)
	_ = os.Remove(metaPath)
}

// Clear removes all cache files.
func (s *Store) Clear() error {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		_ = os.Remove(filepath.Join(s.dir, e.Name()))
	}
	return nil
}

// Stats returns basic usage.
func (s *Store) Stats() (files int, bytes int64, dir string) {
	dir = s.dir
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return 0, 0, dir
	}
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".bin") {
			continue
		}
		st, err := e.Info()
		if err != nil {
			continue
		}
		files++
		bytes += st.Size()
	}
	return files, bytes, dir
}
