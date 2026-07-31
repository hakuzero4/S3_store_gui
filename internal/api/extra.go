package api

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"io"
	"net/http"
	"path"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/hakup/s3store/internal/config"
)

func (s *Server) routesExtra() {
	s.mux.HandleFunc("POST /api/objects/move", s.handleMoveObject)
	s.mux.HandleFunc("POST /api/objects/batch-copy", s.handleBatchCopy)
	s.mux.HandleFunc("POST /api/objects/batch-move", s.handleBatchMove)
	s.mux.HandleFunc("POST /api/objects/zip", s.handleZipDownload)
	s.mux.HandleFunc("GET /api/objects/content", s.handleObjectContent)
	s.mux.HandleFunc("GET /api/profiles/export", s.handleExportProfiles)
	s.mux.HandleFunc("POST /api/profiles/import", s.handleImportProfiles)
}

// ---- batch copy / move ----

type batchBody struct {
	SrcBucket string   `json:"srcBucket"`
	DstBucket string   `json:"dstBucket"`
	SrcPrefix string   `json:"srcPrefix"`
	DstPrefix string   `json:"dstPrefix"`
	Keys      []string `json:"keys"`
	Prefixes  []string `json:"prefixes"`
}

func (s *Server) handleMoveObject(w http.ResponseWriter, r *http.Request) {
	c := s.requireClient(w)
	if c == nil {
		return
	}
	var body struct {
		SrcBucket string `json:"srcBucket"`
		DstBucket string `json:"dstBucket"`
		Src       string `json:"src"`
		Dst       string `json:"dst"`
	}
	if err := readJSON(r, &body); err != nil || body.Src == "" || body.Dst == "" {
		writeErr(w, http.StatusBadRequest, errors.New("src and dst are required"))
		return
	}
	srcB := body.SrcBucket
	dstB := body.DstBucket
	if srcB == "" {
		srcB = body.DstBucket
	}
	if dstB == "" {
		dstB = srcB
	}
	if srcB == "" {
		writeErr(w, http.StatusBadRequest, errors.New("bucket is required"))
		return
	}
	if err := c.MoveObjectTo(r.Context(), srcB, body.Src, dstB, body.Dst); err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleBatchCopy(w http.ResponseWriter, r *http.Request) {
	s.handleBatchTransfer(w, r, false)
}

func (s *Server) handleBatchMove(w http.ResponseWriter, r *http.Request) {
	s.handleBatchTransfer(w, r, true)
}

func (s *Server) handleBatchTransfer(w http.ResponseWriter, r *http.Request, move bool) {
	c := s.requireClient(w)
	if c == nil {
		return
	}
	var body batchBody
	if err := readJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	if body.SrcBucket == "" {
		writeErr(w, http.StatusBadRequest, errors.New("srcBucket is required"))
		return
	}
	if body.DstBucket == "" {
		body.DstBucket = body.SrcBucket
	}
	keys, err := c.ExpandSelection(r.Context(), body.SrcBucket, body.Keys, body.Prefixes)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	if len(keys) == 0 {
		writeErr(w, http.StatusBadRequest, errors.New("no objects selected"))
		return
	}
	// safety cap
	if len(keys) > 5000 {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("too many objects (%d), max 5000", len(keys)))
		return
	}
	var n int
	if move {
		n, err = c.MoveKeys(r.Context(), body.SrcBucket, keys, body.SrcPrefix, body.DstBucket, body.DstPrefix)
	} else {
		n, err = c.CopyKeys(r.Context(), body.SrcBucket, keys, body.SrcPrefix, body.DstBucket, body.DstPrefix)
	}
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "count": n})
}

// ---- zip download ----

func (s *Server) handleZipDownload(w http.ResponseWriter, r *http.Request) {
	c := s.requireClient(w)
	if c == nil {
		return
	}
	var body struct {
		Bucket   string   `json:"bucket"`
		Keys     []string `json:"keys"`
		Prefixes []string `json:"prefixes"`
		Name     string   `json:"name"`
	}
	if err := readJSON(r, &body); err != nil || body.Bucket == "" {
		writeErr(w, http.StatusBadRequest, errors.New("bucket is required"))
		return
	}
	keys, err := c.ExpandSelection(r.Context(), body.Bucket, body.Keys, body.Prefixes)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	// filter out folder placeholders ending with / and empty
	var files []string
	for _, k := range keys {
		if k == "" || strings.HasSuffix(k, "/") {
			continue
		}
		files = append(files, k)
	}
	if len(files) == 0 {
		writeErr(w, http.StatusBadRequest, errors.New("no files to zip"))
		return
	}
	if len(files) > 2000 {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("too many files (%d), max 2000", len(files)))
		return
	}
	name := body.Name
	if name == "" {
		name = "download-" + time.Now().Format("20060102-150405") + ".zip"
	}
	if !strings.HasSuffix(strings.ToLower(name), ".zip") {
		name += ".zip"
	}
	// Preflight first object head to fail early before headers
	if _, err := c.HeadObject(r.Context(), body.Bucket, files[0]); err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", path.Base(name)))
	zw := zip.NewWriter(w)
	defer zw.Close()
	for _, key := range files {
		rc, _, err := c.GetObject(r.Context(), body.Bucket, key)
		if err != nil {
			log.Printf("zip get %s: %v", key, err)
			return
		}
		entryName := strings.TrimLeft(key, "/")
		fw, err := zw.CreateHeader(&zip.FileHeader{Name: entryName, Method: zip.Deflate})
		if err != nil {
			rc.Close()
			log.Printf("zip create %s: %v", key, err)
			return
		}
		_, copyErr := io.Copy(fw, rc)
		rc.Close()
		if copyErr != nil {
			log.Printf("zip copy %s: %v", key, copyErr)
			return
		}
	}
}

// ---- text content ----

func (s *Server) handleObjectContent(w http.ResponseWriter, r *http.Request) {
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
	const maxBytes = 1 << 20 // 1 MiB
	data, detail, err := c.ReadObjectLimited(r.Context(), bucket, key, maxBytes)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	truncated := detail != nil && detail.Size > int64(len(data))
	// detect binary
	binary := false
	for i := 0; i < len(data); i++ {
		if data[i] == 0 {
			binary = true
			break
		}
	}
	text := ""
	if !binary {
		if utf8.Valid(data) {
			text = string(data)
		} else {
			binary = true
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"key":       key,
		"size":      detail.Size,
		"contentType": detail.ContentType,
		"text":      text,
		"binary":    binary,
		"truncated": truncated,
		"maxBytes":  maxBytes,
	})
}

// ---- profile import/export ----

type exportPayload struct {
	Version  int              `json:"version"`
	Exported time.Time        `json:"exported"`
	Profiles []config.Profile `json:"profiles"`
}

func (s *Server) handleExportProfiles(w http.ResponseWriter, r *http.Request) {
	profiles := s.store.ListProfiles()
	// optionally strip secrets? user asked export for backup - include secrets
	payload := exportPayload{
		Version:  1,
		Exported: time.Now().UTC(),
		Profiles: profiles,
	}
	w.Header().Set("Content-Disposition", "attachment; filename=s3store-profiles.json")
	writeJSON(w, http.StatusOK, payload)
}

func (s *Server) handleImportProfiles(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Profiles []config.Profile `json:"profiles"`
		// also accept raw array
		Merge bool `json:"merge"`
	}
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	// try object first
	if err := json.Unmarshal(raw, &body); err != nil || len(body.Profiles) == 0 {
		var arr []config.Profile
		if err2 := json.Unmarshal(raw, &arr); err2 != nil || len(arr) == 0 {
			writeErr(w, http.StatusBadRequest, errors.New("invalid profiles payload"))
			return
		}
		body.Profiles = arr
		body.Merge = true
	}
	if !body.Merge {
		body.Merge = true
	}
	imported := 0
	for _, p := range body.Profiles {
		if strings.TrimSpace(p.Name) == "" || p.AccessKey == "" {
			continue
		}
		if p.ID == "" {
			p.ID = uuid.NewString()
		}
		if p.Region == "" {
			p.Region = "auto"
		}
		// if id collision with different content, keep id (upsert)
		s.store.UpsertProfile(p)
		imported++
	}
	_ = s.store.Save()
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "imported": imported})
}
