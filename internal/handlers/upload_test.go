package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testLogger(t *testing.T) *slog.Logger {
	t.Helper()
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func makeTestPNG(t *testing.T, width, height int) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := range height {
		for x := range width {
			if x < width/2 {
				img.SetNRGBA(x, y, color.NRGBA{R: 200, G: 200, B: 200, A: 255})
			} else {
				img.SetNRGBA(x, y, color.NRGBA{R: 30, G: 30, B: 30, A: 255})
			}
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode test PNG: %v", err)
	}
	return buf.Bytes()
}

func createUploadRequest(t *testing.T, fieldName, filename string, data []byte) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile(fieldName, filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(data); err != nil {
		t.Fatalf("write data: %v", err)
	}
	_ = writer.Close()

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func TestUploadSuccess(t *testing.T) {
	dir := t.TempDir()
	logger := testLogger(t)
	upload := NewUpload(dir, 10, logger)

	pngData := makeTestPNG(t, 60, 60)
	req := createUploadRequest(t, "photo", "test.png", pngData)
	w := httptest.NewRecorder()

	upload.HandleUpload(w, req)

	if w.Code != http.StatusOK {
		body, _ := io.ReadAll(w.Body)
		t.Fatalf("expected 200, got %d: %s", w.Code, body)
	}

	var result map[string]string
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result["url"] == "" {
		t.Error("expected non-empty url")
	}
	if result["filename"] == "" {
		t.Error("expected non-empty filename")
	}
}

func TestUploadReturnsProcessedPNG(t *testing.T) {
	dir := t.TempDir()
	logger := testLogger(t)
	upload := NewUpload(dir, 10, logger)

	pngData := makeTestPNG(t, 80, 80)
	req := createUploadRequest(t, "photo", "subject.png", pngData)
	w := httptest.NewRecorder()

	upload.HandleUpload(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var result map[string]string
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if result["processed"] != "true" {
		t.Errorf("expected processed=true, got %q", result["processed"])
	}
	if result["original"] == "" {
		t.Error("expected non-empty original URL")
	}
}

func TestUploadMissingField(t *testing.T) {
	dir := t.TempDir()
	logger := testLogger(t)
	upload := NewUpload(dir, 10, logger)

	req := createUploadRequest(t, "wrong_field", "test.png", []byte("data"))
	w := httptest.NewRecorder()

	upload.HandleUpload(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestUploadBadExtension(t *testing.T) {
	dir := t.TempDir()
	logger := testLogger(t)
	upload := NewUpload(dir, 10, logger)

	req := createUploadRequest(t, "photo", "test.exe", []byte("data"))
	w := httptest.NewRecorder()

	upload.HandleUpload(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for .exe, got %d", w.Code)
	}
}

func TestUploadAllowedExtensions(t *testing.T) {
	dir := t.TempDir()
	logger := testLogger(t)
	upload := NewUpload(dir, 10, logger)

	pngData := makeTestPNG(t, 40, 40)

	for _, ext := range []string{"jpg", "jpeg", "png", "webp"} {
		req := createUploadRequest(t, "photo", "test."+ext, pngData)
		w := httptest.NewRecorder()
		upload.HandleUpload(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("extension .%s: expected 200, got %d", ext, w.Code)
		}
	}
}

func TestUploadRejectedExtensions(t *testing.T) {
	dir := t.TempDir()
	logger := testLogger(t)
	upload := NewUpload(dir, 10, logger)

	for _, ext := range []string{"gif", "bmp", "tiff", "svg", "exe", "pdf"} {
		req := createUploadRequest(t, "photo", "test."+ext, []byte("data"))
		w := httptest.NewRecorder()
		upload.HandleUpload(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("extension .%s: expected 400, got %d", ext, w.Code)
		}
	}
}

func TestUploadInvalidMultipartForm(t *testing.T) {
	dir := t.TempDir()
	logger := testLogger(t)
	upload := NewUpload(dir, 10, logger)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/upload", bytes.NewReader([]byte("not multipart")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	upload.HandleUpload(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestUploadServeFiles(t *testing.T) {
	dir := t.TempDir()
	logger := testLogger(t)
	upload := NewUpload(dir, 10, logger)

	mux := http.NewServeMux()
	upload.RegisterRoutes(mux)

	pngData := makeTestPNG(t, 40, 40)
	req := createUploadRequest(t, "photo", "test.png", pngData)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("upload: expected 200, got %d", w.Code)
	}

	var result map[string]string
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}

	fetchReq := httptest.NewRequestWithContext(context.Background(), http.MethodGet, result["url"], nil)
	fetchW := httptest.NewRecorder()
	mux.ServeHTTP(fetchW, fetchReq)

	if fetchW.Code != http.StatusOK {
		t.Errorf("serve uploaded file: expected 200, got %d", fetchW.Code)
	}
	if fetchW.Body.Len() == 0 {
		t.Error("expected non-empty file body")
	}
}

func TestNewUpload(t *testing.T) {
	logger := testLogger(t)
	u := NewUpload("/tmp/test", 5, logger)
	if u.uploadDir != "/tmp/test" {
		t.Errorf("expected /tmp/test, got %s", u.uploadDir)
	}
	if u.maxBytes != 5*1024*1024 {
		t.Errorf("expected %d, got %d", 5*1024*1024, u.maxBytes)
	}
	if !u.allowedExt[".jpg"] {
		t.Error("expected .jpg to be allowed")
	}
	if u.allowedExt[".gif"] {
		t.Error("expected .gif to NOT be allowed")
	}
}

func TestGenerateFilename(t *testing.T) {
	name1 := generateFilename(".png")
	name2 := generateFilename(".png")

	if name1 == name2 {
		t.Error("expected unique filenames")
	}
	if len(name1) < 5 {
		t.Errorf("filename too short: %q", name1)
	}
	if name1[len(name1)-4:] != ".png" {
		t.Errorf("expected .png extension, got %q", name1)
	}
}
