package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"io"
	"io/fs"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/srivickynesh/light-simulator/internal/handlers"
	"github.com/srivickynesh/light-simulator/internal/lighting"
	"github.com/srivickynesh/light-simulator/internal/middleware"
	"github.com/srivickynesh/light-simulator/internal/models"
	"github.com/srivickynesh/light-simulator/web"
)

func setupServer(t *testing.T) *httptest.Server {
	t.Helper()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mux := http.NewServeMux()

	staticSub, err := fs.Sub(web.Content, "static")
	if err != nil {
		t.Fatalf("embedded static fs: %v", err)
	}
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticSub))))

	templateSub, err := fs.Sub(web.Content, "templates")
	if err != nil {
		t.Fatalf("embedded template fs: %v", err)
	}
	pages := handlers.NewPagesFS(templateSub, true, logger)
	pages.RegisterRoutes(mux)

	api := handlers.NewAPI()
	api.RegisterRoutes(mux)

	uploadDir := t.TempDir()
	upload := handlers.NewUpload(uploadDir, 10, logger)
	upload.RegisterRoutes(mux)

	handler := middleware.Chain(
		mux,
		middleware.Recover(logger),
		middleware.Logger(logger),
		middleware.CORS,
		middleware.SecurityHeaders,
	)

	return httptest.NewServer(handler)
}

func doGet(t *testing.T, url string) *http.Response {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	return resp
}

func doPost(t *testing.T, url string, body []byte) *http.Response {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST %s: %v", url, err)
	}
	return resp
}

func TestE2E_HealthEndpoint(t *testing.T) {
	srv := setupServer(t)
	defer srv.Close()

	resp := doGet(t, srv.URL+"/api/health")
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("health: expected 200, got %d", resp.StatusCode)
	}

	var data map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if data["status"] != "ok" {
		t.Errorf("expected status 'ok', got %q", data["status"])
	}
}

func TestE2E_HTMLPages(t *testing.T) {
	srv := setupServer(t)
	defer srv.Close()

	pages := []string{"/", "/simulator", "/cheatsheet"}
	for _, path := range pages {
		resp := doGet(t, srv.URL+path)
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("GET %s: expected 200, got %d", path, resp.StatusCode)
			continue
		}

		ct := resp.Header.Get("Content-Type")
		if !strings.Contains(ct, "text/html") {
			t.Errorf("GET %s: expected text/html, got %q", path, ct)
		}

		html := string(body)
		if !strings.Contains(html, "Light Simulator") && !strings.Contains(html, "simulator") {
			t.Errorf("GET %s: body does not contain expected content", path)
		}
	}
}

func TestE2E_404Page(t *testing.T) {
	srv := setupServer(t)
	defer srv.Close()

	resp := doGet(t, srv.URL+"/nonexistent")
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestE2E_ListPresetsAndGetEach(t *testing.T) {
	srv := setupServer(t)
	defer srv.Close()

	resp := doGet(t, srv.URL+"/api/presets")
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("presets: expected 200, got %d", resp.StatusCode)
	}

	var categories map[string][]models.Preset
	if err := json.NewDecoder(resp.Body).Decode(&categories); err != nil {
		t.Fatalf("decode: %v", err)
	}

	totalPresets := 0
	for cat, presets := range categories {
		if len(presets) == 0 {
			t.Errorf("category %q has no presets", cat)
		}
		for _, p := range presets {
			totalPresets++
			resp2 := doGet(t, srv.URL+"/api/presets/"+p.ID)
			if resp2.StatusCode != http.StatusOK {
				t.Errorf("GET preset %q: expected 200, got %d", p.ID, resp2.StatusCode)
			}

			var fetched models.Preset
			if err := json.NewDecoder(resp2.Body).Decode(&fetched); err != nil {
				t.Errorf("preset %q: decode error: %v", p.ID, err)
			}
			_ = resp2.Body.Close()

			if fetched.ID != p.ID {
				t.Errorf("expected ID %q, got %q", p.ID, fetched.ID)
			}
			if len(fetched.Scene.Lights) == 0 {
				t.Errorf("preset %q has no lights", p.ID)
			}
		}
	}

	if totalPresets < 24 {
		t.Errorf("expected at least 24 presets total, got %d", totalPresets)
	}
}

func TestE2E_GetPresetNotFound(t *testing.T) {
	srv := setupServer(t)
	defer srv.Close()

	resp := doGet(t, srv.URL+"/api/presets/does_not_exist")
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestE2E_AnalyzeScene(t *testing.T) {
	srv := setupServer(t)
	defer srv.Close()

	scene := models.Scene{
		Lights: []models.Light{
			{ID: "key", Role: models.RoleKey, Enabled: true, Modifier: models.ModifierSoftbox,
				Position: models.Position3D{X: -1.5, Y: 0.5, Z: 1.5, Distance: 2.0, Angle: 45},
				Power:    75, ColorTemp: 5500, CRI: 95},
			{ID: "fill", Role: models.RoleFill, Enabled: true, Modifier: models.ModifierUmbrella,
				Position: models.Position3D{X: 1.0, Y: 0, Z: 2.0, Distance: 2.2, Angle: -25},
				Power:    30, ColorTemp: 5500, CRI: 90},
		},
		Camera: models.CameraSettings{
			FocalLength: 85, Aperture: 2.8, ShutterSpeed: "1/200",
			ISO: 100, WhiteBalance: 5500, SensorSize: "full_frame",
			Distance: 2.5,
		},
		Backdrop: "#1a1a1a",
		Ambient:  0.1,
	}

	body, _ := json.Marshal(scene)
	resp := doPost(t, srv.URL+"/api/analyze", body)
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("analyze: expected 200, got %d", resp.StatusCode)
	}

	var analysis lighting.SceneAnalysis
	if err := json.NewDecoder(resp.Body).Decode(&analysis); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if len(analysis.Contributions) != 2 {
		t.Errorf("expected 2 contributions, got %d", len(analysis.Contributions))
	}
	if analysis.KeyToFillRatio <= 0 {
		t.Error("expected positive key-to-fill ratio")
	}
	if analysis.OverallEV == 0 {
		t.Error("expected non-zero EV")
	}
	if analysis.ShadowQuality == "" {
		t.Error("expected non-empty shadow quality")
	}
	if analysis.CatchlightType == "" {
		t.Error("expected non-empty catchlight type")
	}
	if analysis.CSSFilters.Brightness == 0 {
		t.Error("expected non-zero brightness filter")
	}
	if analysis.CSSFilters.ShadowGradient == "" {
		t.Error("expected shadow gradient CSS")
	}
}

func TestE2E_AnalyzeInvalidJSON(t *testing.T) {
	srv := setupServer(t)
	defer srv.Close()

	resp := doPost(t, srv.URL+"/api/analyze", []byte("not-json"))
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestE2E_FlashGuides(t *testing.T) {
	srv := setupServer(t)
	defer srv.Close()

	resp := doGet(t, srv.URL+"/api/guides/flash")
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("flash guides: expected 200, got %d", resp.StatusCode)
	}

	var guides []json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&guides); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(guides) < 3 {
		t.Errorf("expected at least 3 flash guides, got %d", len(guides))
	}
}

func TestE2E_ModifierGuides(t *testing.T) {
	srv := setupServer(t)
	defer srv.Close()

	resp := doGet(t, srv.URL+"/api/guides/modifiers")
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("modifier guides: expected 200, got %d", resp.StatusCode)
	}

	var guides []json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&guides); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(guides) < 6 {
		t.Errorf("expected at least 6 modifier guides, got %d", len(guides))
	}
}

func TestE2E_LensGuides(t *testing.T) {
	srv := setupServer(t)
	defer srv.Close()

	resp := doGet(t, srv.URL+"/api/guides/lenses")
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("lens guides: expected 200, got %d", resp.StatusCode)
	}

	var guides []json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&guides); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(guides) < 4 {
		t.Errorf("expected at least 4 lens guides, got %d", len(guides))
	}
}

func TestE2E_SecurityHeaders(t *testing.T) {
	srv := setupServer(t)
	defer srv.Close()

	resp := doGet(t, srv.URL+"/api/health")
	_ = resp.Body.Close()

	if resp.Header.Get("X-Content-Type-Options") != "nosniff" {
		t.Error("missing X-Content-Type-Options")
	}
	if resp.Header.Get("X-Frame-Options") != "DENY" {
		t.Error("missing X-Frame-Options")
	}
}

func TestE2E_CORSHeaders(t *testing.T) {
	srv := setupServer(t)
	defer srv.Close()

	resp := doGet(t, srv.URL+"/api/health")
	_ = resp.Body.Close()

	if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Error("missing CORS Allow-Origin")
	}
}

func TestE2E_CORSPreflight(t *testing.T) {
	srv := setupServer(t)
	defer srv.Close()

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodOptions, srv.URL+"/api/analyze", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("OPTIONS: %v", err)
	}
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected 204, got %d", resp.StatusCode)
	}
}

func makeE2EPNG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			if x > w/4 && x < 3*w/4 && y > h/4 && y < 3*h/4 {
				img.SetNRGBA(x, y, color.NRGBA{R: 30, G: 30, B: 30, A: 255})
			} else {
				img.SetNRGBA(x, y, color.NRGBA{R: 200, G: 180, B: 160, A: 255})
			}
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode PNG: %v", err)
	}
	return buf.Bytes()
}

func TestE2E_FileUpload(t *testing.T) {
	srv := setupServer(t)
	defer srv.Close()

	pngData := makeE2EPNG(t, 100, 100)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("photo", "test.png")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(pngData); err != nil {
		t.Fatalf("write file data: %v", err)
	}
	_ = writer.Close()

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, srv.URL+"/api/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("upload: expected 200, got %d: %s", resp.StatusCode, body)
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result["url"] == "" {
		t.Error("expected non-empty upload URL")
	}
	if result["filename"] == "" {
		t.Error("expected non-empty filename")
	}
	if result["processed"] != "true" {
		t.Errorf("expected processed=true, got %q", result["processed"])
	}
}

func TestE2E_FileUploadAndServe(t *testing.T) {
	srv := setupServer(t)
	defer srv.Close()

	pngData := makeE2EPNG(t, 80, 80)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("photo", "subject.png")
	_, _ = part.Write(pngData)
	_ = writer.Close()

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, srv.URL+"/api/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}

	fetchResp := doGet(t, srv.URL+result["url"])
	fetchBody, _ := io.ReadAll(fetchResp.Body)
	_ = fetchResp.Body.Close()

	if fetchResp.StatusCode != http.StatusOK {
		t.Errorf("serve uploaded file: expected 200, got %d", fetchResp.StatusCode)
	}
	if len(fetchBody) == 0 {
		t.Error("served file is empty")
	}

	decoded, err := png.Decode(bytes.NewReader(fetchBody))
	if err != nil {
		t.Fatalf("result is not valid PNG: %v", err)
	}
	if decoded.Bounds().Dx() != 80 || decoded.Bounds().Dy() != 80 {
		t.Errorf("expected 80x80 PNG, got %dx%d", decoded.Bounds().Dx(), decoded.Bounds().Dy())
	}
}

func TestE2E_FileUploadBadExtension(t *testing.T) {
	srv := setupServer(t)
	defer srv.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("photo", "test.exe")
	_, _ = part.Write([]byte("data"))
	_ = writer.Close()

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, srv.URL+"/api/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for bad extension, got %d", resp.StatusCode)
	}
}

func TestE2E_FileUploadMissingField(t *testing.T) {
	srv := setupServer(t)
	defer srv.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("wrong_field", "test.png")
	_, _ = part.Write([]byte("data"))
	_ = writer.Close()

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, srv.URL+"/api/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for missing field, got %d", resp.StatusCode)
	}
}

func TestE2E_StaticFiles(t *testing.T) {
	srv := setupServer(t)
	defer srv.Close()

	staticFiles := []string{
		"/static/css/main.css",
		"/static/js/simulator.js",
		"/static/js/cheatsheet.js",
	}

	for _, path := range staticFiles {
		resp := doGet(t, srv.URL+path)
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("GET %s: expected 200, got %d", path, resp.StatusCode)
			continue
		}

		if len(body) == 0 {
			t.Errorf("GET %s: empty response body", path)
		}
	}
}

func TestE2E_FullWorkflow(t *testing.T) {
	srv := setupServer(t)
	defer srv.Close()

	// 1. Load presets
	presetsResp := doGet(t, srv.URL+"/api/presets")
	var categories map[string][]models.Preset
	if err := json.NewDecoder(presetsResp.Body).Decode(&categories); err != nil {
		t.Fatalf("decode presets: %v", err)
	}
	_ = presetsResp.Body.Close()

	// 2. Pick first preset
	var firstPreset models.Preset
	for _, presets := range categories {
		if len(presets) > 0 {
			firstPreset = presets[0]
			break
		}
	}
	if firstPreset.ID == "" {
		t.Fatal("no presets found")
	}

	// 3. Load specific preset
	presetResp := doGet(t, srv.URL+"/api/presets/"+firstPreset.ID)
	var loadedPreset models.Preset
	if err := json.NewDecoder(presetResp.Body).Decode(&loadedPreset); err != nil {
		t.Fatalf("decode preset: %v", err)
	}
	_ = presetResp.Body.Close()

	if loadedPreset.ID != firstPreset.ID {
		t.Errorf("expected ID %q, got %q", firstPreset.ID, loadedPreset.ID)
	}

	// 4. Analyze the preset's scene
	sceneBody, _ := json.Marshal(loadedPreset.Scene)
	analyzeResp := doPost(t, srv.URL+"/api/analyze", sceneBody)
	var analysis lighting.SceneAnalysis
	if err := json.NewDecoder(analyzeResp.Body).Decode(&analysis); err != nil {
		t.Fatalf("decode analysis: %v", err)
	}
	_ = analyzeResp.Body.Close()

	if len(analysis.Contributions) != len(loadedPreset.Scene.Lights) {
		t.Errorf("contributions count %d != lights count %d",
			len(analysis.Contributions), len(loadedPreset.Scene.Lights))
	}

	// 5. Load guides
	for _, endpoint := range []string{"/api/guides/flash", "/api/guides/modifiers", "/api/guides/lenses"} {
		resp := doGet(t, srv.URL+endpoint)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("GET %s: expected 200, got %d", endpoint, resp.StatusCode)
		}
		_ = resp.Body.Close()
	}
}

func TestE2E_PresetFlashSettings(t *testing.T) {
	srv := setupServer(t)
	defer srv.Close()

	resp := doGet(t, srv.URL+"/api/presets/rembrandt")
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var preset models.Preset
	if err := json.NewDecoder(resp.Body).Decode(&preset); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if len(preset.Scene.Lights) == 0 {
		t.Fatal("preset has no lights")
	}

	for i, l := range preset.Scene.Lights {
		if l.Type == "" {
			t.Errorf("light %d: missing type", i)
		}
		if l.Modifier == "" {
			t.Errorf("light %d: missing modifier", i)
		}
		if l.Role == "" {
			t.Errorf("light %d: missing role", i)
		}
		if l.ColorTemp == 0 {
			t.Errorf("light %d: missing color_temp", i)
		}
		if l.CRI == 0 {
			t.Errorf("light %d: missing CRI", i)
		}
	}

	if preset.Scene.Camera.FocalLength == 0 {
		t.Error("camera missing focal_length")
	}
	if preset.Scene.Camera.Aperture == 0 {
		t.Error("camera missing aperture")
	}
	if preset.Scene.Camera.ShutterSpeed == "" {
		t.Error("camera missing shutter_speed")
	}
	if preset.Scene.Camera.ISO == 0 {
		t.Error("camera missing iso")
	}
	if preset.Scene.Camera.WhiteBalance == 0 {
		t.Error("camera missing white_balance")
	}
}

func TestE2E_AllPresetsHaveFlashDetails(t *testing.T) {
	srv := setupServer(t)
	defer srv.Close()

	presetsResp := doGet(t, srv.URL+"/api/presets")
	var categories map[string][]models.Preset
	if err := json.NewDecoder(presetsResp.Body).Decode(&categories); err != nil {
		t.Fatalf("decode: %v", err)
	}
	_ = presetsResp.Body.Close()

	for cat, presets := range categories {
		for _, p := range presets {
			resp := doGet(t, srv.URL+"/api/presets/"+p.ID)
			var full models.Preset
			if err := json.NewDecoder(resp.Body).Decode(&full); err != nil {
				t.Errorf("%s/%s: decode: %v", cat, p.ID, err)
				_ = resp.Body.Close()
				continue
			}
			_ = resp.Body.Close()

			for i, l := range full.Scene.Lights {
				if l.Power == 0 {
					t.Errorf("%s/%s: light %d (%s) has zero power", cat, p.ID, i, l.Name)
				}
				if l.Position.Distance == 0 {
					t.Errorf("%s/%s: light %d (%s) has zero distance", cat, p.ID, i, l.Name)
				}
			}
		}
	}
}

func TestE2E_SimulatorPageHasCustomPresetUI(t *testing.T) {
	srv := setupServer(t)
	defer srv.Close()

	resp := doGet(t, srv.URL+"/simulator")
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	html := string(body)
	requiredElements := []string{
		"savePresetBtn",
		"saveDialogOverlay",
		"savePresetName",
		"savePresetCategory",
		"confirmSaveBtn",
		"cancelSaveBtn",
		"customPresetsPanel",
		"customPresetsList",
	}

	for _, el := range requiredElements {
		if !strings.Contains(html, el) {
			t.Errorf("simulator page missing element: %s", el)
		}
	}
}

func TestE2E_AnalyzeAllPresets(t *testing.T) {
	srv := setupServer(t)
	defer srv.Close()

	presetsResp := doGet(t, srv.URL+"/api/presets")
	var categories map[string][]models.Preset
	if err := json.NewDecoder(presetsResp.Body).Decode(&categories); err != nil {
		t.Fatalf("decode: %v", err)
	}
	_ = presetsResp.Body.Close()

	for cat, presets := range categories {
		for _, p := range presets {
			body, _ := json.Marshal(p.Scene)
			resp := doPost(t, srv.URL+"/api/analyze", body)

			if resp.StatusCode != http.StatusOK {
				t.Errorf("analyze %s/%s: expected 200, got %d", cat, p.ID, resp.StatusCode)
				_ = resp.Body.Close()
				continue
			}

			var analysis lighting.SceneAnalysis
			if err := json.NewDecoder(resp.Body).Decode(&analysis); err != nil {
				t.Errorf("analyze %s/%s: decode error: %v", cat, p.ID, err)
			}
			_ = resp.Body.Close()

			if analysis.OverallEV == 0 {
				t.Errorf("analyze %s/%s: expected non-zero EV", cat, p.ID)
			}
		}
	}
}
