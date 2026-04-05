package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/srivickynesh/light-simulator/internal/imgproc"
)

// Upload handles user photo uploads for live lighting preview.
type Upload struct {
	uploadDir  string
	maxBytes   int64
	logger     *slog.Logger
	allowedExt map[string]bool
}

// NewUpload creates an upload handler.
func NewUpload(uploadDir string, maxMB int64, logger *slog.Logger) *Upload {
	return &Upload{
		uploadDir: uploadDir,
		maxBytes:  maxMB * 1024 * 1024,
		logger:    logger,
		allowedExt: map[string]bool{
			".jpg": true, ".jpeg": true, ".png": true, ".webp": true,
		},
	}
}

// RegisterRoutes mounts upload endpoints.
func (u *Upload) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/upload", u.HandleUpload)
	mux.Handle("GET /uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(u.uploadDir))))
}

// HandleUpload processes multipart file uploads, removes the background,
// and returns a transparent PNG ready for lighting simulation.
func (u *Upload) HandleUpload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, u.maxBytes)

	if err := r.ParseMultipartForm(u.maxBytes); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "file too large or invalid form"})
		return
	}

	file, header, err := r.FormFile("photo")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing 'photo' field"})
		return
	}
	defer func() { _ = file.Close() }()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !u.allowedExt[ext] {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf("unsupported file type: %s (allowed: jpg, png, webp)", ext),
		})
		return
	}

	if err := os.MkdirAll(u.uploadDir, 0o755); err != nil {
		u.logger.Error("failed to create upload dir", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "server error"})
		return
	}

	origFilename := generateFilename(ext)
	origPath := filepath.Join(u.uploadDir, origFilename)

	dest, err := os.Create(origPath)
	if err != nil {
		u.logger.Error("failed to create file", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "server error"})
		return
	}

	if _, err := io.Copy(dest, file); err != nil {
		_ = dest.Close()
		u.logger.Error("failed to write file", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "server error"})
		return
	}
	_ = dest.Close()

	processedFilename := generateFilename(".png")
	processedPath := filepath.Join(u.uploadDir, processedFilename)

	if err := imgproc.RemoveBackground(origPath, processedPath); err != nil {
		u.logger.Error("background removal failed, serving original", "error", err)
		u.logger.Info("file uploaded (original)", "filename", origFilename, "size", header.Size)
		writeJSON(w, http.StatusOK, map[string]string{
			"url":       "/uploads/" + origFilename,
			"filename":  origFilename,
			"processed": "false",
		})
		return
	}

	u.logger.Info("file uploaded and processed", "original", origFilename, "processed", processedFilename, "size", header.Size)
	writeJSON(w, http.StatusOK, map[string]string{
		"url":       "/uploads/" + processedFilename,
		"original":  "/uploads/" + origFilename,
		"filename":  processedFilename,
		"processed": "true",
	})
}

func generateFilename(ext string) string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "upload" + ext
	}
	return hex.EncodeToString(b) + ext
}
