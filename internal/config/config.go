package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Profile is an S3-compatible connection profile (AWS S3, Cloudflare R2, MinIO, etc.).
type Profile struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Endpoint       string `json:"endpoint"`
	Region         string `json:"region"`
	AccessKey      string `json:"accessKey"`
	SecretKey      string `json:"secretKey"`
	ForcePathStyle bool   `json:"forcePathStyle"`
	// Provider hint: aws | r2 | minio | other
	Provider string `json:"provider"`
	// DefaultBucket optional preferred bucket
	DefaultBucket string `json:"defaultBucket,omitempty"`
}

type Store struct {
	Profiles    []Profile `json:"profiles"`
	ActiveID    string    `json:"activeId"`
	ListenAddr  string    `json:"listenAddr"`
	OpenBrowser bool      `json:"openBrowser"`
	Theme       string    `json:"theme"`
	mu          sync.RWMutex
	path        string
}

// DefaultPath returns config.json next to the executable.
// When launched via `go run` (temp build dir), falls back to the working directory.
func DefaultPath() (string, error) {
	dir, err := appDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func appDataDir() (string, error) {
	exe, err := os.Executable()
	if err == nil {
		if resolved, rerr := filepath.EvalSymlinks(exe); rerr == nil {
			exe = resolved
		}
		dir := filepath.Dir(exe)
		if !isEphemeralExeDir(dir) {
			return dir, nil
		}
	}
	// go run / tests: keep data with the project / launch cwd
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return cwd, nil
}

func isEphemeralExeDir(dir string) bool {
	lower := strings.ToLower(filepath.Clean(dir))
	// Go build cache / temp binaries
	markers := []string{
		string(filepath.Separator) + "go-build",
		string(filepath.Separator) + "tmp",
		string(filepath.Separator) + "temp",
		`\appdata\local\temp`,
		`/appdata/local/temp`,
	}
	for _, m := range markers {
		if strings.Contains(lower, m) {
			return true
		}
	}
	return false
}

func Load(path string) (*Store, error) {
	s := &Store{
		Profiles:    []Profile{},
		ListenAddr:  "127.0.0.1:17890",
		OpenBrowser: true,
		Theme:       "light",
		path:        path,
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(data, s); err != nil {
		return nil, err
	}
	s.path = path
	if s.ListenAddr == "" {
		s.ListenAddr = "127.0.0.1:17890"
	}
	if s.Theme == "" {
		s.Theme = "light"
	}
	return s, nil
}

func (s *Store) Path() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.path
}

func (s *Store) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	dir := filepath.Dir(s.path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		// fallback direct write (e.g. some FS without rename)
		return os.WriteFile(s.path, data, 0o600)
	}
	if err := os.Rename(tmp, s.path); err != nil {
		_ = os.Remove(tmp)
		return os.WriteFile(s.path, data, 0o600)
	}
	return nil
}

func (s *Store) Snapshot() Store {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cp := Store{
		Profiles:    append([]Profile(nil), s.Profiles...),
		ActiveID:    s.ActiveID,
		ListenAddr:  s.ListenAddr,
		OpenBrowser: s.OpenBrowser,
		Theme:       s.Theme,
	}
	return cp
}

func (s *Store) ListProfiles() []Profile {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Profile, len(s.Profiles))
	copy(out, s.Profiles)
	return out
}

func (s *Store) GetProfile(id string) (Profile, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, p := range s.Profiles {
		if p.ID == id {
			return p, true
		}
	}
	return Profile{}, false
}

func (s *Store) ActiveProfile() (Profile, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.ActiveID == "" {
		return Profile{}, false
	}
	for _, p := range s.Profiles {
		if p.ID == s.ActiveID {
			return p, true
		}
	}
	return Profile{}, false
}

func (s *Store) UpsertProfile(p Profile) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, existing := range s.Profiles {
		if existing.ID == p.ID {
			s.Profiles[i] = p
			return
		}
	}
	s.Profiles = append(s.Profiles, p)
}

func (s *Store) DeleteProfile(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := s.Profiles[:0]
	for _, p := range s.Profiles {
		if p.ID != id {
			out = append(out, p)
		}
	}
	s.Profiles = out
	if s.ActiveID == id {
		s.ActiveID = ""
		if len(s.Profiles) > 0 {
			s.ActiveID = s.Profiles[0].ID
		}
	}
}

func (s *Store) SetActive(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ActiveID = id
}

func (s *Store) SetTheme(theme string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Theme = theme
}
