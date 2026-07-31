package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hakup/s3store/internal/config"
	"github.com/hakup/s3store/internal/s3client"
)

type Server struct {
	store  *config.Store
	mu     sync.RWMutex
	client *s3client.Client
	mux    *http.ServeMux
}

func New(store *config.Store) *Server {
	s := &Server{store: store, mux: http.NewServeMux()}
	s.routes()
	// warm active client if possible
	if p, ok := store.ActiveProfile(); ok {
		if c, err := s3client.NewFromProfile(p); err == nil {
			s.client = c
		}
	}
	return s
}

func (s *Server) Handler() http.Handler {
	return cors(s.mux)
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /api/health", s.handleHealth)
	s.mux.HandleFunc("GET /api/app", s.handleAppInfo)

	s.mux.HandleFunc("GET /api/profiles", s.handleListProfiles)
	s.mux.HandleFunc("POST /api/profiles", s.handleUpsertProfile)
	s.mux.HandleFunc("PUT /api/profiles/{id}", s.handleUpdateProfile)
	s.mux.HandleFunc("DELETE /api/profiles/{id}", s.handleDeleteProfile)
	s.mux.HandleFunc("POST /api/profiles/{id}/activate", s.handleActivateProfile)
	s.mux.HandleFunc("POST /api/profiles/test", s.handleTestProfile)

	s.mux.HandleFunc("GET /api/buckets", s.handleListBuckets)
	s.mux.HandleFunc("POST /api/buckets", s.handleCreateBucket)
	s.mux.HandleFunc("DELETE /api/buckets/{name}", s.handleDeleteBucket)

	s.mux.HandleFunc("GET /api/objects", s.handleListObjects)
	s.mux.HandleFunc("GET /api/objects/detail", s.handleObjectDetail)
	s.mux.HandleFunc("POST /api/objects/folder", s.handleCreateFolder)
	s.mux.HandleFunc("POST /api/objects/upload", s.handleUpload)
	s.mux.HandleFunc("GET /api/objects/download", s.handleDownload)
	s.mux.HandleFunc("POST /api/objects/delete", s.handleDeleteObjects)
	s.mux.HandleFunc("POST /api/objects/copy", s.handleCopyObject)
	s.mux.HandleFunc("POST /api/objects/rename", s.handleRenameObject)
	s.mux.HandleFunc("POST /api/objects/presign", s.handlePresign)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "time": time.Now().UTC()})
}

func (s *Server) handleAppInfo(w http.ResponseWriter, r *http.Request) {
	snap := s.store.Snapshot()
	active, _ := s.store.ActiveProfile()
	writeJSON(w, http.StatusOK, map[string]any{
		"theme":      snap.Theme,
		"activeId":   snap.ActiveID,
		"activeName": active.Name,
		"provider":   active.Provider,
		"hasClient":  s.getClient() != nil,
		"configPath": s.store.Path(),
	})
}

func (s *Server) handleListProfiles(w http.ResponseWriter, r *http.Request) {
	profiles := s.store.ListProfiles()
	// mask secrets lightly for list? keep full for local app
	writeJSON(w, http.StatusOK, map[string]any{
		"profiles": profiles,
		"activeId": s.store.Snapshot().ActiveID,
	})
}

type profilePayload struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Endpoint       string `json:"endpoint"`
	Region         string `json:"region"`
	AccessKey      string `json:"accessKey"`
	SecretKey      string `json:"secretKey"`
	ForcePathStyle bool   `json:"forcePathStyle"`
	Provider       string `json:"provider"`
	DefaultBucket  string `json:"defaultBucket"`
	Activate       bool   `json:"activate"`
}

func (s *Server) handleUpsertProfile(w http.ResponseWriter, r *http.Request) {
	var p profilePayload
	if err := readJSON(r, &p); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	if strings.TrimSpace(p.Name) == "" {
		writeErr(w, http.StatusBadRequest, errors.New("name is required"))
		return
	}
	if p.ID == "" {
		p.ID = uuid.NewString()
	} else if existing, ok := s.store.GetProfile(p.ID); ok {
		if p.SecretKey == "" {
			p.SecretKey = existing.SecretKey
		}
		if p.AccessKey == "" {
			p.AccessKey = existing.AccessKey
		}
	}
	if p.Region == "" {
		p.Region = "auto"
	}
	if p.Provider == "" {
		p.Provider = detectProvider(p.Endpoint)
	}
	// R2 defaults
	if p.Provider == "r2" && !p.ForcePathStyle {
		// virtual-host style works with account subdomain endpoint
		p.ForcePathStyle = false
	}
	prof := config.Profile{
		ID:             p.ID,
		Name:           p.Name,
		Endpoint:       strings.TrimSpace(p.Endpoint),
		Region:         p.Region,
		AccessKey:      p.AccessKey,
		SecretKey:      p.SecretKey,
		ForcePathStyle: p.ForcePathStyle,
		Provider:       p.Provider,
		DefaultBucket:  p.DefaultBucket,
	}
	s.store.UpsertProfile(prof)
	if p.Activate || s.store.Snapshot().ActiveID == "" {
		s.store.SetActive(prof.ID)
		if c, err := s3client.NewFromProfile(prof); err == nil {
			s.setClient(c)
		}
	}
	_ = s.store.Save()
	writeJSON(w, http.StatusOK, prof)
}

func (s *Server) handleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	existing, ok := s.store.GetProfile(id)
	if !ok {
		writeErr(w, http.StatusNotFound, errors.New("profile not found"))
		return
	}
	var p profilePayload
	if err := readJSON(r, &p); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	if p.Name != "" {
		existing.Name = p.Name
	}
	existing.Endpoint = strings.TrimSpace(p.Endpoint)
	if p.Region != "" {
		existing.Region = p.Region
	}
	if p.AccessKey != "" {
		existing.AccessKey = p.AccessKey
	}
	if p.SecretKey != "" {
		existing.SecretKey = p.SecretKey
	}
	existing.ForcePathStyle = p.ForcePathStyle
	if p.Provider != "" {
		existing.Provider = p.Provider
	}
	existing.DefaultBucket = p.DefaultBucket
	s.store.UpsertProfile(existing)
	if s.store.Snapshot().ActiveID == id {
		if c, err := s3client.NewFromProfile(existing); err == nil {
			s.setClient(c)
		}
	}
	_ = s.store.Save()
	writeJSON(w, http.StatusOK, existing)
}

func (s *Server) handleDeleteProfile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s.store.DeleteProfile(id)
	if p, ok := s.store.ActiveProfile(); ok {
		if c, err := s3client.NewFromProfile(p); err == nil {
			s.setClient(c)
		} else {
			s.setClient(nil)
		}
	} else {
		s.setClient(nil)
	}
	_ = s.store.Save()
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleActivateProfile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	p, ok := s.store.GetProfile(id)
	if !ok {
		writeErr(w, http.StatusNotFound, errors.New("profile not found"))
		return
	}
	c, err := s3client.NewFromProfile(p)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	if err := c.Test(r.Context()); err != nil {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("connection failed: %w", err))
		return
	}
	s.store.SetActive(id)
	s.setClient(c)
	_ = s.store.Save()
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "profile": p})
}

func (s *Server) handleTestProfile(w http.ResponseWriter, r *http.Request) {
	var p profilePayload
	if err := readJSON(r, &p); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	// allow testing existing profile by id with blank secret (reuse stored)
	if p.ID != "" {
		if existing, ok := s.store.GetProfile(p.ID); ok {
			if p.SecretKey == "" {
				p.SecretKey = existing.SecretKey
			}
			if p.AccessKey == "" {
				p.AccessKey = existing.AccessKey
			}
			if p.Endpoint == "" {
				p.Endpoint = existing.Endpoint
			}
			if p.Region == "" {
				p.Region = existing.Region
			}
		}
	}
	if p.Region == "" {
		p.Region = "auto"
	}
	c, err := s3client.NewFromProfile(config.Profile{
		Endpoint:       p.Endpoint,
		Region:         p.Region,
		AccessKey:      p.AccessKey,
		SecretKey:      p.SecretKey,
		ForcePathStyle: p.ForcePathStyle,
		Provider:       p.Provider,
	})
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	if err := c.Test(r.Context()); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	buckets, err := c.ListBuckets(r.Context())
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "buckets": []any{}, "warning": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "buckets": buckets})
}

func (s *Server) requireClient(w http.ResponseWriter) *s3client.Client {
	c := s.getClient()
	if c == nil {
		writeErr(w, http.StatusPreconditionFailed, errors.New("no active connection profile"))
		return nil
	}
	return c
}

func (s *Server) handleListBuckets(w http.ResponseWriter, r *http.Request) {
	c := s.requireClient(w)
	if c == nil {
		return
	}
	items, err := c.ListBuckets(r.Context())
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"buckets": items})
}

func (s *Server) handleCreateBucket(w http.ResponseWriter, r *http.Request) {
	c := s.requireClient(w)
	if c == nil {
		return
	}
	var body struct {
		Name string `json:"name"`
	}
	if err := readJSON(r, &body); err != nil || strings.TrimSpace(body.Name) == "" {
		writeErr(w, http.StatusBadRequest, errors.New("name is required"))
		return
	}
	if err := c.CreateBucket(r.Context(), strings.TrimSpace(body.Name)); err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleDeleteBucket(w http.ResponseWriter, r *http.Request) {
	c := s.requireClient(w)
	if c == nil {
		return
	}
	name := r.PathValue("name")
	if err := c.DeleteBucket(r.Context(), name); err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleListObjects(w http.ResponseWriter, r *http.Request) {
	c := s.requireClient(w)
	if c == nil {
		return
	}
	q := r.URL.Query()
	bucket := q.Get("bucket")
	if bucket == "" {
		writeErr(w, http.StatusBadRequest, errors.New("bucket is required"))
		return
	}
	prefix := q.Get("prefix")
	token := q.Get("token")
	maxKeys := int32(200)
	if v := q.Get("maxKeys"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			maxKeys = int32(n)
		}
	}
	res, err := c.ListObjects(r.Context(), bucket, prefix, token, maxKeys)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (s *Server) handleObjectDetail(w http.ResponseWriter, r *http.Request) {
	c := s.requireClient(w)
	if c == nil {
		return
	}
	bucket := r.URL.Query().Get("bucket")
	key := r.URL.Query().Get("key")
	if bucket == "" || key == "" {
		writeErr(w, http.StatusBadRequest, errors.New("bucket and key are required"))
		return
	}
	d, err := c.HeadObject(r.Context(), bucket, key)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, d)
}

func (s *Server) handleCreateFolder(w http.ResponseWriter, r *http.Request) {
	c := s.requireClient(w)
	if c == nil {
		return
	}
	var body struct {
		Bucket string `json:"bucket"`
		Key    string `json:"key"`
	}
	if err := readJSON(r, &body); err != nil || body.Bucket == "" || body.Key == "" {
		writeErr(w, http.StatusBadRequest, errors.New("bucket and key are required"))
		return
	}
	if err := c.CreateFolder(r.Context(), body.Bucket, body.Key); err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	c := s.requireClient(w)
	if c == nil {
		return
	}
	// Limit memory; stream file from multipart
	const maxMem = 32 << 20
	if err := r.ParseMultipartForm(maxMem); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	bucket := r.FormValue("bucket")
	key := r.FormValue("key")
	if bucket == "" {
		writeErr(w, http.StatusBadRequest, errors.New("bucket is required"))
		return
	}
	file, hdr, err := r.FormFile("file")
	if err != nil {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("file is required: %w", err))
		return
	}
	defer file.Close()
	if key == "" {
		key = hdr.Filename
	}
	ct := hdr.Header.Get("Content-Type")
	if ct == "" {
		ct = "application/octet-stream"
	}
	if err := c.PutObject(r.Context(), bucket, key, ct, file, hdr.Size); err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "key": key, "size": hdr.Size})
}

func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	c := s.requireClient(w)
	if c == nil {
		return
	}
	bucket := r.URL.Query().Get("bucket")
	key := r.URL.Query().Get("key")
	if bucket == "" || key == "" {
		writeErr(w, http.StatusBadRequest, errors.New("bucket and key are required"))
		return
	}
	body, detail, err := c.GetObject(r.Context(), bucket, key)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	defer body.Close()
	name := path.Base(key)
	if detail.ContentType != "" {
		w.Header().Set("Content-Type", detail.ContentType)
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	disp := "attachment"
	if r.URL.Query().Get("inline") == "1" || r.URL.Query().Get("inline") == "true" {
		disp = "inline"
	}
	w.Header().Set("Content-Disposition", fmt.Sprintf("%s; filename*=UTF-8''%s", disp, url.PathEscape(name)))
	w.Header().Set("Cache-Control", "private, max-age=120")
	if detail.Size > 0 {
		w.Header().Set("Content-Length", strconv.FormatInt(detail.Size, 10))
	}
	w.WriteHeader(http.StatusOK)
	if _, err := io.Copy(w, body); err != nil {
		log.Printf("download stream error: %v", err)
	}
}

func (s *Server) handleDeleteObjects(w http.ResponseWriter, r *http.Request) {
	c := s.requireClient(w)
	if c == nil {
		return
	}
	var body struct {
		Bucket   string   `json:"bucket"`
		Keys     []string `json:"keys"`
		Prefixes []string `json:"prefixes"`
	}
	if err := readJSON(r, &body); err != nil || body.Bucket == "" {
		writeErr(w, http.StatusBadRequest, errors.New("bucket is required"))
		return
	}
	for _, pfx := range body.Prefixes {
		if err := c.DeletePrefix(r.Context(), body.Bucket, pfx); err != nil {
			writeErr(w, http.StatusBadGateway, err)
			return
		}
	}
	if err := c.DeleteObjects(r.Context(), body.Bucket, body.Keys); err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleCopyObject(w http.ResponseWriter, r *http.Request) {
	c := s.requireClient(w)
	if c == nil {
		return
	}
	var body struct {
		Bucket string `json:"bucket"`
		Src    string `json:"src"`
		Dst    string `json:"dst"`
	}
	if err := readJSON(r, &body); err != nil || body.Bucket == "" || body.Src == "" || body.Dst == "" {
		writeErr(w, http.StatusBadRequest, errors.New("bucket, src, dst are required"))
		return
	}
	if err := c.CopyObject(r.Context(), body.Bucket, body.Src, body.Dst); err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleRenameObject(w http.ResponseWriter, r *http.Request) {
	c := s.requireClient(w)
	if c == nil {
		return
	}
	var body struct {
		Bucket string `json:"bucket"`
		Src    string `json:"src"`
		Dst    string `json:"dst"`
	}
	if err := readJSON(r, &body); err != nil || body.Bucket == "" || body.Src == "" || body.Dst == "" {
		writeErr(w, http.StatusBadRequest, errors.New("bucket, src, dst are required"))
		return
	}
	if err := c.RenameObject(r.Context(), body.Bucket, body.Src, body.Dst); err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handlePresign(w http.ResponseWriter, r *http.Request) {
	c := s.requireClient(w)
	if c == nil {
		return
	}
	var body struct {
		Bucket  string `json:"bucket"`
		Key     string `json:"key"`
		Expires int    `json:"expires"` // seconds
	}
	if err := readJSON(r, &body); err != nil || body.Bucket == "" || body.Key == "" {
		writeErr(w, http.StatusBadRequest, errors.New("bucket and key are required"))
		return
	}
	exp := time.Duration(body.Expires) * time.Second
	if exp <= 0 {
		exp = time.Hour
	}
	u, err := c.PresignGet(r.Context(), body.Bucket, body.Key, exp)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"url": u, "expires": int(exp.Seconds())})
}

func (s *Server) getClient() *s3client.Client {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.client
}

func (s *Server) setClient(c *s3client.Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.client = c
}

func detectProvider(endpoint string) string {
	e := strings.ToLower(endpoint)
	switch {
	case strings.Contains(e, "r2.cloudflarestorage.com"):
		return "r2"
	case strings.Contains(e, "amazonaws.com"):
		return "aws"
	case strings.Contains(e, "aliyuncs.com"):
		return "oss"
	case endpoint == "":
		return "aws"
	default:
		return "other"
	}
}

func readJSON(r *http.Request, dst any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]any{"error": err.Error()})
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
