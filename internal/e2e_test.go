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

	// Real-world Rembrandt setup: key 45° left at 2m, fill reflector 30° right at 1.5m
	scene := models.Scene{
		Lights: []models.Light{
			{ID: "key", Role: models.RoleKey, Enabled: true, Type: models.LightTypeStrobe,
				Modifier: models.ModifierSoftbox,
				Position: models.Position3D{X: -1.41, Y: 0.5, Z: -1.41, Distance: 2.0, Angle: -135},
				Power:    75, ColorTemp: 5500, CRI: 95},
			{ID: "fill", Role: models.RoleFill, Enabled: true, Type: models.LightTypeContinuous,
				Modifier: models.ModifierReflector,
				Position: models.Position3D{X: 0.75, Y: 0.0, Z: -1.30, Distance: 1.5, Angle: 210},
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
	if analysis.CatchlightType != "rectangular" {
		t.Errorf("softbox key should produce rectangular catchlight, got %q", analysis.CatchlightType)
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

func TestE2E_AnalyzeSceneWithPanels(t *testing.T) {
	srv := setupServer(t)
	defer srv.Close()

	// Realistic studio scene: softbox key 45° left, camera-side bounce below chin,
	// black V-flat on shadow side to deepen contrast. Mirrors a classic editorial setup.
	scene := models.Scene{
		Lights: []models.Light{
			{ID: "key", Role: models.RoleKey, Enabled: true, Type: models.LightTypeStrobe,
				Modifier: models.ModifierSoftbox,
				Position: models.Position3D{X: -1.41, Y: 0.5, Z: -1.41, Distance: 2.0, Angle: -135},
				Power:    80, ColorTemp: 5600, CRI: 95},
			{ID: "fill", Role: models.RoleFill, Enabled: true, Type: models.LightTypeContinuous,
				Modifier: models.ModifierUmbrella,
				Position: models.Position3D{X: 1.10, Y: 0, Z: -1.91, Distance: 2.2, Angle: 210},
				Power:    25, ColorTemp: 5500, CRI: 90},
		},
		Panels: []models.Panel{
			{ID: "neg", Name: "Black V-Flat", Type: models.PanelNegativeFill, Size: models.PanelSizeLarge,
				Position: models.Position3D{X: 1.0, Y: 0, Z: 0, Distance: 1.0, Angle: 90}, Enabled: true},
			{ID: "chin_bounce", Name: "White Chin Bounce", Type: models.PanelBounceWhite, Size: models.PanelSizeMedium,
				Position: models.Position3D{X: 0, Y: -0.4, Z: -0.6, Distance: 0.6, Angle: 180}, Enabled: true},
		},
		Camera: models.CameraSettings{
			FocalLength: 85, Aperture: 2.8, ShutterSpeed: "1/200",
			ISO: 100, WhiteBalance: 5500, SensorSize: "full_frame", Distance: 2.5,
		},
		Backdrop: "#1a1a1a", Ambient: 0.1,
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

	if len(analysis.PanelEffects) != 2 {
		t.Errorf("expected 2 panel effects, got %d", len(analysis.PanelEffects))
	}
	if len(analysis.Contributions) != 2 {
		t.Errorf("expected 2 contributions, got %d", len(analysis.Contributions))
	}

	// V-flat on shadow side should have negative intensity (absorbing light)
	for _, pe := range analysis.PanelEffects {
		if pe.PanelID == "neg" && pe.EffectIntensity >= 0 {
			t.Errorf("negative fill should have negative effect, got %f", pe.EffectIntensity)
		}
		if pe.PanelID == "chin_bounce" && pe.EffectIntensity <= 0 {
			t.Errorf("white bounce should have positive effect, got %f", pe.EffectIntensity)
		}
	}
}

func TestE2E_SimulatorPageHasPanelUI(t *testing.T) {
	srv := setupServer(t)
	defer srv.Close()

	resp := doGet(t, srv.URL+"/simulator")
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	html := string(body)
	for _, el := range []string{"addPanelBtn", "panelsList", "tab-panels"} {
		if !strings.Contains(html, el) {
			t.Errorf("simulator page missing panel element: %s", el)
		}
	}
}

func TestE2E_SimulatorPageHasAccessoriesDropdown(t *testing.T) {
	srv := setupServer(t)
	defer srv.Close()

	resp := doGet(t, srv.URL+"/simulator")
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	html := string(body)
	required := []string{
		"accessorySelect",
		"fill_light",
		"rim_light",
		"hair_light",
		"bg_light",
		"neg_fill",
		"bounce_white",
		"bounce_silver",
		"bounce_gold",
		"flag",
		"diffusion",
	}
	for _, el := range required {
		if !strings.Contains(html, el) {
			t.Errorf("simulator page missing accessory element: %s", el)
		}
	}
}

func TestE2E_SimulatorJSHasDeleteAndAccessorySupport(t *testing.T) {
	srv := setupServer(t)
	defer srv.Close()

	resp := doGet(t, srv.URL+"/static/js/simulator.js")
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	js := string(body)
	required := []string{
		"data-delete-light",
		"data-delete-panel",
		"createSVGDeleteButton",
		"addAccessory",
		"ACCESSORY_DEFAULTS",
		"svg-delete-btn",
		"panel-rotate",
		"normalizedIntensity",
		"distToSubject",
		"beamGrad_",
		"peakOpacity",
		"isoFractions",
		"effectiveReachPx",
		"renderSunBeam",
		"renderSunMarker",
		"isSun",
	}
	for _, s := range required {
		if !strings.Contains(js, s) {
			t.Errorf("simulator.js missing: %s", s)
		}
	}
}

func TestE2E_SimulatorUsesNewSubjectImage(t *testing.T) {
	srv := setupServer(t)
	defer srv.Close()

	resp := doGet(t, srv.URL+"/simulator")
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()

	html := string(body)
	if !strings.Contains(html, "default-subject.svg") {
		t.Error("simulator page should reference default-subject.svg")
	}

	svgResp := doGet(t, srv.URL+"/static/images/default-subject.svg")
	svgBody, _ := io.ReadAll(svgResp.Body)
	_ = svgResp.Body.Close()

	if svgResp.StatusCode != http.StatusOK {
		t.Errorf("default-subject.svg: expected 200, got %d", svgResp.StatusCode)
	}
	if len(svgBody) < 500 {
		t.Error("default-subject.svg seems too small")
	}
}

func TestE2E_PanelEffectsAffectCSSFilters(t *testing.T) {
	srv := setupServer(t)
	defer srv.Close()

	// Baseline: single key light at 45° left, 2m, softbox — common studio setup
	sceneNoPanels := models.Scene{
		Lights: []models.Light{
			{ID: "key", Role: models.RoleKey, Enabled: true, Type: models.LightTypeStrobe,
				Modifier: models.ModifierSoftbox,
				Position: models.Position3D{X: -1.41, Y: 0.5, Z: -1.41, Distance: 2.0, Angle: -135},
				Power:    75, ColorTemp: 5500},
		},
		Camera: models.CameraSettings{Aperture: 2.8, ISO: 100, ShutterSpeed: "1/200"},
	}

	// Same key, plus two V-flats on shadow side and behind subject
	sceneWithNegFill := models.Scene{
		Lights: sceneNoPanels.Lights,
		Panels: []models.Panel{
			{ID: "neg", Name: "V-Flat", Type: models.PanelNegativeFill, Size: models.PanelSizeLarge,
				Position: models.Position3D{X: 1.0, Y: 0, Z: 0, Distance: 1.0, Angle: 90}, Enabled: true},
			{ID: "neg2", Name: "V-Flat Rear", Type: models.PanelNegativeFill, Size: models.PanelSizeLarge,
				Position: models.Position3D{X: 1.41, Y: 0, Z: 1.41, Distance: 2.0, Angle: 45}, Enabled: true},
		},
		Camera: sceneNoPanels.Camera,
	}

	body1, _ := json.Marshal(sceneNoPanels)
	resp1 := doPost(t, srv.URL+"/api/analyze", body1)
	var analysis1 lighting.SceneAnalysis
	if err := json.NewDecoder(resp1.Body).Decode(&analysis1); err != nil {
		t.Fatalf("decode: %v", err)
	}
	_ = resp1.Body.Close()

	body2, _ := json.Marshal(sceneWithNegFill)
	resp2 := doPost(t, srv.URL+"/api/analyze", body2)
	var analysis2 lighting.SceneAnalysis
	if err := json.NewDecoder(resp2.Body).Decode(&analysis2); err != nil {
		t.Fatalf("decode: %v", err)
	}
	_ = resp2.Body.Close()

	if len(analysis2.PanelEffects) == 0 {
		t.Error("expected panel effects in response")
	}

	if analysis2.CSSFilters.Brightness >= analysis1.CSSFilters.Brightness {
		t.Errorf("negative fill should reduce brightness; without=%.3f, with=%.3f",
			analysis1.CSSFilters.Brightness, analysis2.CSSFilters.Brightness)
	}
}

func TestE2E_BounceBoostsFillandReducesRatio(t *testing.T) {
	srv := setupServer(t)
	defer srv.Close()

	// Studio test: key at 45° camera-left, weak fill at 25° camera-right.
	// Adding an XL white bounce close on fill side should lower the key:fill ratio.
	sceneNoBounce := models.Scene{
		Lights: []models.Light{
			{ID: "key", Role: models.RoleKey, Enabled: true, Type: models.LightTypeStrobe,
				Modifier: models.ModifierSoftbox,
				Position: models.Position3D{X: -1.41, Y: 0.5, Z: -1.41, Distance: 2.0, Angle: -135},
				Power:    80, ColorTemp: 5500},
			{ID: "fill", Role: models.RoleFill, Enabled: true, Type: models.LightTypeContinuous,
				Modifier: models.ModifierUmbrella,
				Position: models.Position3D{X: 1.10, Y: 0, Z: -1.91, Distance: 2.2, Angle: 210},
				Power:    10, ColorTemp: 5500},
		},
		Camera: models.CameraSettings{Aperture: 2.8, ISO: 100, ShutterSpeed: "1/200"},
	}

	sceneWithBounce := models.Scene{
		Lights: sceneNoBounce.Lights,
		Panels: []models.Panel{
			{ID: "wb", Name: "White V-Flat Bounce", Type: models.PanelBounceWhite, Size: models.PanelSizeXLarge,
				Position: models.Position3D{X: 0.49, Y: 0, Z: -0.49, Distance: 0.7, Angle: 225}, Enabled: true},
		},
		Camera: sceneNoBounce.Camera,
	}

	body1, _ := json.Marshal(sceneNoBounce)
	resp1 := doPost(t, srv.URL+"/api/analyze", body1)
	var a1 lighting.SceneAnalysis
	_ = json.NewDecoder(resp1.Body).Decode(&a1)
	_ = resp1.Body.Close()

	body2, _ := json.Marshal(sceneWithBounce)
	resp2 := doPost(t, srv.URL+"/api/analyze", body2)
	var a2 lighting.SceneAnalysis
	_ = json.NewDecoder(resp2.Body).Decode(&a2)
	_ = resp2.Body.Close()

	if a2.KeyToFillRatio >= a1.KeyToFillRatio {
		t.Errorf("bounce should reduce key:fill ratio; without=%.2f, with=%.2f",
			a1.KeyToFillRatio, a2.KeyToFillRatio)
	}
}

func TestE2E_SimulatorJSHasPanelPhysics(t *testing.T) {
	srv := setupServer(t)
	defer srv.Close()

	resp := doGet(t, srv.URL+"/static/js/simulator.js")
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()

	js := string(body)
	required := []string{
		"computeLocalPanelEffects",
		"computeLocalSinglePanelEffect",
		"getPanelSizeFactor",
		"incidentLight",
		"panelIntensityDelta",
		"renderPanelLightInteractions",
		"cosIncidence",
		"spillFraction",
		"edgeFalloff",
	}
	for _, s := range required {
		if !strings.Contains(js, s) {
			t.Errorf("simulator.js missing panel physics function/variable: %s", s)
		}
	}
}

func TestE2E_AllPresetsHavePanelsViaAPI(t *testing.T) {
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

			if len(full.Scene.Panels) == 0 {
				t.Errorf("preset %s/%s has no panels", cat, p.ID)
			}
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

func TestE2E_OutdoorPresetsExist(t *testing.T) {
	srv := setupServer(t)
	defer srv.Close()

	resp := doGet(t, srv.URL+"/api/presets")
	var categories map[string][]models.Preset
	if err := json.NewDecoder(resp.Body).Decode(&categories); err != nil {
		t.Fatalf("decode: %v", err)
	}
	_ = resp.Body.Close()

	outdoor, ok := categories["outdoor"]
	if !ok {
		t.Fatal("no 'outdoor' category in presets")
	}
	if len(outdoor) < 3 {
		t.Errorf("expected at least 3 outdoor presets, got %d", len(outdoor))
	}

	ids := make(map[string]bool)
	for _, p := range outdoor {
		ids[p.ID] = true
	}
	for _, want := range []string{"outdoor_golden_hour", "outdoor_harsh_midday", "outdoor_open_shade"} {
		if !ids[want] {
			t.Errorf("missing outdoor preset: %s", want)
		}
	}
}

func TestE2E_SunLightAnalysis(t *testing.T) {
	srv := setupServer(t)
	defer srv.Close()

	scene := models.Scene{
		ID: "sun_test", Name: "Sun Test", Mode: models.ModeOutdoor,
		Lights: []models.Light{
			{
				ID: "sun1", Name: "Sun", Type: models.LightTypeSun,
				Modifier: models.ModifierNone, Role: models.RoleKey,
				Position: models.Position3D{X: -1, Y: 2, Z: -2, Distance: 3.0, Angle: 210},
				Power:    80, ColorTemp: 5600, CRI: 100, Enabled: true,
			},
		},
		Camera: models.CameraSettings{
			FocalLength: 85, Aperture: 2.8, ShutterSpeed: "1/200",
			ISO: 100, WhiteBalance: 5500, SensorSize: "full_frame",
			Distance: 2.5,
		},
		Ambient: 0.3,
	}

	body, _ := json.Marshal(scene)
	resp := doPost(t, srv.URL+"/api/analyze", body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var analysis lighting.SceneAnalysis
	if err := json.NewDecoder(resp.Body).Decode(&analysis); err != nil {
		t.Fatalf("decode: %v", err)
	}
	_ = resp.Body.Close()

	if len(analysis.Contributions) != 1 {
		t.Fatalf("expected 1 contribution, got %d", len(analysis.Contributions))
	}

	sunContrib := analysis.Contributions[0]
	if sunContrib.Intensity < 50 {
		t.Errorf("sun intensity too low: %f (expected > 50 for power=80)", sunContrib.Intensity)
	}
	if sunContrib.SpillAngle != 180 {
		t.Errorf("sun spill should be 180°, got %f", sunContrib.SpillAngle)
	}
	if sunContrib.Softness > 0.2 {
		t.Errorf("sun softness should be hard (<0.2), got %f", sunContrib.Softness)
	}
}

func TestE2E_SunPanelInteraction(t *testing.T) {
	srv := setupServer(t)
	defer srv.Close()

	scene := models.Scene{
		ID: "sun_panel", Name: "Sun + Reflector", Mode: models.ModeOutdoor,
		Lights: []models.Light{
			{
				ID: "sun1", Name: "Sun", Type: models.LightTypeSun,
				Modifier: models.ModifierNone, Role: models.RoleKey,
				Position: models.Position3D{X: -1, Y: 2, Z: -2, Distance: 3.0, Angle: 210},
				Power:    80, ColorTemp: 5600, CRI: 100, Enabled: true,
			},
		},
		Panels: []models.Panel{
			{ID: "bounce", Name: "Silver Reflector", Type: models.PanelBounceSilver, Size: models.PanelSizeMedium,
				Position: models.Position3D{X: 1, Y: 0, Z: 1, Distance: 1.0, Angle: 45}, Rotation: 225, Enabled: true},
		},
		Camera: models.CameraSettings{
			FocalLength: 85, Aperture: 2.8, ShutterSpeed: "1/200",
			ISO: 100, WhiteBalance: 5500, SensorSize: "full_frame",
			Distance: 2.5,
		},
		Ambient: 0.3,
	}

	body, _ := json.Marshal(scene)
	resp := doPost(t, srv.URL+"/api/analyze", body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var analysis lighting.SceneAnalysis
	if err := json.NewDecoder(resp.Body).Decode(&analysis); err != nil {
		t.Fatalf("decode: %v", err)
	}
	_ = resp.Body.Close()

	if len(analysis.PanelEffects) == 0 {
		t.Fatal("expected panel effects for silver reflector with sun")
	}
	if analysis.PanelEffects[0].EffectIntensity <= 0 {
		t.Errorf("silver reflector should produce positive fill, got %f", analysis.PanelEffects[0].EffectIntensity)
	}
}

func TestE2E_OutdoorModeInHTML(t *testing.T) {
	srv := setupServer(t)
	defer srv.Close()

	resp := doGet(t, srv.URL+"/simulator")
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()

	html := string(body)
	for _, want := range []string{
		`value="outdoor"`,
		`Sun (Outdoor)`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("simulator HTML missing: %s", want)
		}
	}
}

// presetExpectation defines the expected physics output for a given preset
// based on real-world studio photography knowledge.
type presetExpectation struct {
	catchlight    string // expected catchlight type from key modifier
	shadowQuality string // "hard", "medium", "soft"
	minRatio      float64
	maxRatio      float64
	hasNegFill    bool // at least one negative_fill panel
	hasBounce     bool // at least one bounce panel
	hasFlag       bool // at least one flag panel
	hasDiffusion  bool // at least one diffusion panel
	minLights     int
	minPanels     int
	panelSign     string // "positive", "negative", "mixed", "" (don't check)
}

func TestE2E_AllPresetsPhysicsValidation(t *testing.T) {
	srv := setupServer(t)
	defer srv.Close()

	expectations := map[string]presetExpectation{
		// ── Portrait presets ──
		// Rembrandt: softbox key + reflector fill → soft shadows, rectangular catchlight
		"rembrandt": {
			catchlight: "rectangular", shadowQuality: "medium",
			minRatio: 0, maxRatio: 0, hasNegFill: true, hasBounce: true,
			minLights: 2, minPanels: 2, panelSign: "mixed",
		},
		// Butterfly: beauty dish key (softness=0.5) + reflector fill → medium shadows, circular_ring catchlight
		"butterfly": {
			catchlight: "circular_ring", shadowQuality: "medium",
			minRatio: 0, maxRatio: 0, hasBounce: true,
			minLights: 2, minPanels: 1,
		},
		// Split: single softbox key at 90° → soft shadows, no fill
		"split": {
			catchlight: "rectangular", shadowQuality: "soft",
			minRatio: 0, maxRatio: 0, hasNegFill: true, hasFlag: true,
			minLights: 1, minPanels: 2,
		},
		// Loop: octabox key + umbrella fill → soft shadows
		"loop": {
			catchlight: "octagonal", shadowQuality: "soft",
			minRatio: 0, maxRatio: 0, hasBounce: true,
			minLights: 2, minPanels: 1,
		},
		// Clamshell: octabox key + softbox fill → soft shadows
		"clamshell": {
			catchlight: "octagonal", shadowQuality: "soft",
			minRatio: 0, maxRatio: 0, hasBounce: true,
			minLights: 2, minPanels: 1,
		},
		// Broad: softbox key + umbrella fill → soft shadows
		"broad": {
			catchlight: "rectangular", shadowQuality: "soft",
			minRatio: 0, maxRatio: 0, hasBounce: true,
			minLights: 2, minPanels: 1,
		},
		// Short: single softbox key → soft shadows (only 1 light)
		"short": {
			catchlight: "rectangular", shadowQuality: "soft",
			minRatio: 0, maxRatio: 0, hasNegFill: true,
			minLights: 1, minPanels: 2,
		},
		// High key: octabox key(0.85) + umbrella fill(0.65) + 2 bare BG(0.1) → avg ~0.44 = medium
		"high_key": {
			catchlight: "octagonal", shadowQuality: "medium",
			minRatio: 0, maxRatio: 0, hasBounce: true,
			minLights: 4, minPanels: 2,
		},
		// Low key: single honeycomb grid key → hard shadows (softness=0.15)
		"low_key": {
			catchlight: "point", shadowQuality: "hard",
			minRatio: 0, maxRatio: 0, hasNegFill: true, hasFlag: true,
			minLights: 1, minPanels: 2,
		},
		// Beauty ring: ring_light + 2 stripbox accents → avg softness ~0.3 = medium
		"beauty_ring": {
			catchlight: "point", shadowQuality: "medium",
			minRatio: 0, maxRatio: 0, hasBounce: true,
			minLights: 3, minPanels: 1,
		},
		// Cinematic noir: barn_doors key (0.1) + snoot BG (0.05) → hard
		"cinematic_noir": {
			catchlight: "point", shadowQuality: "hard",
			minRatio: 0, maxRatio: 0, hasNegFill: true, hasFlag: true,
			minLights: 2, minPanels: 2,
		},
		// Cross light: 2 softbox keys + honeycomb hair → avg ~0.55 = medium
		"cross_light": {
			catchlight: "rectangular", shadowQuality: "medium",
			minRatio: 0, maxRatio: 0, hasNegFill: true,
			minLights: 3, minPanels: 2,
		},
		// Rim dramatic: 2 stripbox rims, no key → "none" catchlight, soft shadows (0.7 avg)
		"rim_dramatic": {
			catchlight: "none", shadowQuality: "soft",
			minRatio: 0, maxRatio: 0, hasBounce: true,
			minLights: 2, minPanels: 1,
		},

		// ── Product presets ──
		// Product top-down: 2 stripbox lights → soft (avg 0.7)
		"product_topdown": {
			catchlight: "rectangular", shadowQuality: "soft",
			minRatio: 0, maxRatio: 0, hasBounce: true,
			minLights: 2, minPanels: 1,
		},
		// Product hero: softbox key + stripbox rim + reflector fill → medium
		"product_hero": {
			catchlight: "rectangular", shadowQuality: "medium",
			minRatio: 0, maxRatio: 0, hasBounce: true, hasNegFill: true,
			minLights: 3, minPanels: 2, panelSign: "mixed",
		},
		// Product white BG: diffusion key (0.9) + 2 bare BG (0.1) → medium
		"product_white_bg": {
			catchlight: "point", shadowQuality: "medium",
			minRatio: 0, maxRatio: 0, hasBounce: true,
			minLights: 3, minPanels: 1,
		},
		// Product glassware: 2 stripbox + 1 snoot → medium
		"product_glass": {
			catchlight: "rectangular", shadowQuality: "medium",
			minRatio: 0, maxRatio: 0, hasNegFill: true, hasBounce: true,
			minLights: 3, minPanels: 3, panelSign: "mixed",
		},

		// ── Fashion presets ──
		// Fashion editorial: parabolic key + honeycomb hair + stripbox kicker → medium
		"fashion_editorial": {
			catchlight: "parabolic", shadowQuality: "medium",
			minRatio: 0, maxRatio: 0, hasNegFill: true,
			minLights: 3, minPanels: 1,
		},
		// Fashion catalog: octabox key(0.85) + softbox fill(0.75) + 2 bare BG(0.1) → ~0.45 medium
		"fashion_catalog": {
			catchlight: "octagonal", shadowQuality: "medium",
			minRatio: 0, maxRatio: 0, hasBounce: true,
			minLights: 4, minPanels: 2,
		},

		// ── Food presets ──
		// Food moody: diffusion key (0.9) + reflector fill (0.3) → avg 0.6 = soft
		"food_moody": {
			catchlight: "point", shadowQuality: "medium",
			minRatio: 0, maxRatio: 0, hasBounce: true, hasFlag: true,
			minLights: 2, minPanels: 4, panelSign: "mixed",
		},
		// Food bright: diffusion key (0.9) + reflector fill (0.3) → medium
		"food_bright": {
			catchlight: "point", shadowQuality: "medium",
			minRatio: 0, maxRatio: 0, hasBounce: true, hasFlag: true,
			minLights: 2, minPanels: 3,
		},

		// ── Headshot / Group / Sport ──
		// Headshot corporate: octabox key + softbox fill → soft (avg 0.8)
		"headshot_corporate": {
			catchlight: "octagonal", shadowQuality: "soft",
			minRatio: 0, maxRatio: 0, hasBounce: true,
			minLights: 2, minPanels: 1,
		},
		// Group photo: 2 umbrella (0.65) + stripbox hair (0.7) → avg ~0.67 = soft
		"group_photo": {
			catchlight: "circular", shadowQuality: "soft",
			minRatio: 0, maxRatio: 0, hasBounce: true,
			minLights: 3, minPanels: 2,
		},
		// Sport action: honeycomb key (0.15) + 2 stripbox rims (0.7) → avg ~0.52 = medium
		"sport_action": {
			catchlight: "point", shadowQuality: "medium",
			minRatio: 0, maxRatio: 0, hasBounce: true,
			minLights: 3, minPanels: 1,
		},

		// ── Outdoor presets ──
		// Outdoor golden hour: sun (0.15 hardness) → hard
		"outdoor_golden_hour": {
			catchlight: "point", shadowQuality: "hard",
			minRatio: 0, maxRatio: 0, hasBounce: true,
			minLights: 1, minPanels: 1,
		},
		// Outdoor harsh midday: sun (0.15) → hard
		"outdoor_harsh_midday": {
			catchlight: "point", shadowQuality: "hard",
			minRatio: 0, maxRatio: 0, hasDiffusion: true, hasBounce: true,
			minLights: 1, minPanels: 2,
		},
		// Outdoor open shade: sun fill (0.15) → hard, no key role → "none" catchlight
		"outdoor_open_shade": {
			catchlight: "none", shadowQuality: "hard",
			minRatio: 0, maxRatio: 0, hasBounce: true,
			minLights: 1, minPanels: 1,
		},
	}

	presetsResp := doGet(t, srv.URL+"/api/presets")
	var categories map[string][]models.Preset
	if err := json.NewDecoder(presetsResp.Body).Decode(&categories); err != nil {
		t.Fatalf("decode presets: %v", err)
	}
	_ = presetsResp.Body.Close()

	testedCount := 0
	for _, presets := range categories {
		for _, p := range presets {
			exp, ok := expectations[p.ID]
			if !ok {
				t.Errorf("preset %q has no physics expectation defined — add one", p.ID)
				continue
			}
			testedCount++

			// Analyze the preset scene
			body, _ := json.Marshal(p.Scene)
			resp := doPost(t, srv.URL+"/api/analyze", body)
			if resp.StatusCode != http.StatusOK {
				t.Errorf("preset %q: analyze returned %d", p.ID, resp.StatusCode)
				_ = resp.Body.Close()
				continue
			}
			var analysis lighting.SceneAnalysis
			if err := json.NewDecoder(resp.Body).Decode(&analysis); err != nil {
				t.Errorf("preset %q: decode analysis: %v", p.ID, err)
				_ = resp.Body.Close()
				continue
			}
			_ = resp.Body.Close()

			t.Run(p.ID, func(t *testing.T) {
				// Validate light count
				if len(p.Scene.Lights) < exp.minLights {
					t.Errorf("expected >= %d lights, got %d", exp.minLights, len(p.Scene.Lights))
				}

				// Validate panel count
				if len(p.Scene.Panels) < exp.minPanels {
					t.Errorf("expected >= %d panels, got %d", exp.minPanels, len(p.Scene.Panels))
				}

				// Validate contributions match light count
				if len(analysis.Contributions) != len(p.Scene.Lights) {
					t.Errorf("contributions (%d) != lights (%d)",
						len(analysis.Contributions), len(p.Scene.Lights))
				}

				// Validate catchlight type
				if exp.catchlight != "" && analysis.CatchlightType != exp.catchlight {
					t.Errorf("catchlight: expected %q, got %q", exp.catchlight, analysis.CatchlightType)
				}

				// Validate shadow quality
				if exp.shadowQuality != "" && analysis.ShadowQuality != exp.shadowQuality {
					t.Errorf("shadow quality: expected %q, got %q", exp.shadowQuality, analysis.ShadowQuality)
				}

				// Validate key:fill ratio (only for presets with fill light)
				if exp.minRatio > 0 || exp.maxRatio > 0 {
					if analysis.KeyToFillRatio < exp.minRatio {
						t.Errorf("key:fill ratio %.2f below min %.2f", analysis.KeyToFillRatio, exp.minRatio)
					}
					if exp.maxRatio > 0 && analysis.KeyToFillRatio > exp.maxRatio {
						t.Errorf("key:fill ratio %.2f above max %.2f", analysis.KeyToFillRatio, exp.maxRatio)
					}
				}

				// Validate EV is reasonable
				if analysis.OverallEV == 0 {
					t.Error("expected non-zero EV")
				}

				// Validate panel types present
				panelTypes := make(map[string]bool)
				for _, panel := range p.Scene.Panels {
					panelTypes[string(panel.Type)] = true
				}
				if exp.hasNegFill && !panelTypes[string(models.PanelNegativeFill)] {
					t.Error("expected negative fill panel")
				}
				if exp.hasBounce {
					hasBounce := panelTypes[string(models.PanelBounceWhite)] ||
						panelTypes[string(models.PanelBounceSilver)] ||
						panelTypes[string(models.PanelBounceGold)]
					if !hasBounce {
						t.Error("expected at least one bounce panel")
					}
				}
				if exp.hasFlag && !panelTypes[string(models.PanelFlag)] {
					t.Error("expected flag panel")
				}
				if exp.hasDiffusion && !panelTypes[string(models.PanelDiffusion)] {
					t.Error("expected diffusion panel")
				}

				// Validate panel effects exist for each enabled panel
				enabledPanelCount := 0
				for _, panel := range p.Scene.Panels {
					if panel.Enabled {
						enabledPanelCount++
					}
				}
				if len(analysis.PanelEffects) != enabledPanelCount {
					t.Errorf("panel effects (%d) != enabled panels (%d)",
						len(analysis.PanelEffects), enabledPanelCount)
				}

				// Validate panel sign (positive bounce vs negative absorption)
				if exp.panelSign != "" {
					var hasPos, hasNeg bool
					for _, pe := range analysis.PanelEffects {
						if pe.EffectIntensity > 0 {
							hasPos = true
						}
						if pe.EffectIntensity < 0 {
							hasNeg = true
						}
					}
					switch exp.panelSign {
					case "positive":
						if !hasPos {
							t.Error("expected positive panel effects (bounce)")
						}
					case "negative":
						if !hasNeg {
							t.Error("expected negative panel effects (absorption)")
						}
					case "mixed":
						if !hasPos || !hasNeg {
							t.Errorf("expected both positive and negative panel effects; pos=%v neg=%v", hasPos, hasNeg)
						}
					}
				}

				// Validate CSS filters are reasonable
				if analysis.CSSFilters.Brightness <= 0 {
					t.Error("expected positive CSS brightness")
				}
				if analysis.CSSFilters.ShadowGradient == "" {
					t.Error("expected shadow gradient CSS")
				}

				// Validate all contributions have positive intensity
				for _, c := range analysis.Contributions {
					if c.Intensity <= 0 {
						t.Errorf("light %q has non-positive intensity: %f", c.LightID, c.Intensity)
					}
				}

				// Validate light roles present
				roles := make(map[string]bool)
				for _, l := range p.Scene.Lights {
					roles[string(l.Role)] = true
				}
				if !roles["key"] && len(p.Scene.Lights) > 0 {
					// Most presets should have a key light (outdoor open shade is exception: fill only)
					if p.ID != "outdoor_open_shade" {
						hasKeyWarning := false
						for _, w := range analysis.Warnings {
							if strings.Contains(w, "key light") {
								hasKeyWarning = true
								break
							}
						}
						if !hasKeyWarning {
							t.Error("no key light and no warning about it")
						}
					}
				}
			})
		}
	}

	if testedCount < 27 {
		t.Errorf("expected to validate at least 27 presets, only tested %d", testedCount)
	}
}

// TestE2E_PresetPositionsRelativeToCamera validates that light and panel
// positions in each preset are physically plausible relative to the camera.
func TestE2E_PresetPositionsRelativeToCamera(t *testing.T) {
	srv := setupServer(t)
	defer srv.Close()

	presetsResp := doGet(t, srv.URL+"/api/presets")
	var categories map[string][]models.Preset
	if err := json.NewDecoder(presetsResp.Body).Decode(&categories); err != nil {
		t.Fatalf("decode: %v", err)
	}
	_ = presetsResp.Body.Close()

	for _, presets := range categories {
		for _, p := range presets {
			t.Run(p.ID, func(t *testing.T) {
				cam := p.Scene.Camera

				// Camera distance should be positive and reasonable (0.5m – 10m)
				if cam.Distance < 0.5 || cam.Distance > 10 {
					t.Errorf("camera distance %.1f out of range [0.5, 10]", cam.Distance)
				}

				for _, l := range p.Scene.Lights {
					// Every light should have a positive distance
					if l.Position.Distance <= 0 {
						t.Errorf("light %q has non-positive distance: %f", l.ID, l.Position.Distance)
					}

					// Validate power is within [0, 100]
					if l.Power < 0 || l.Power > 100 {
						t.Errorf("light %q power %f out of [0, 100]", l.ID, l.Power)
					}

					// Color temp should be realistic (1800K – 10000K)
					if l.ColorTemp < 1800 || l.ColorTemp > 10000 {
						t.Errorf("light %q color temp %d out of [1800, 10000]", l.ID, l.ColorTemp)
					}

					// Key and fill lights should generally be on the camera side (Z < 0)
					// unless it's a product/overhead setup or sun
					if l.Type != models.LightTypeSun && l.Role == models.RoleKey {
						switch p.Scene.Mode {
						case models.ModePortrait, models.ModeHeadshot, models.ModeFashion:
							if l.Position.Z > 0.1 {
								t.Errorf("key light %q Z=%.2f should be <= 0 (camera side) for %s mode",
									l.ID, l.Position.Z, p.Scene.Mode)
							}
						}
					}

					// Background lights should be behind subject (Z > 0)
					if l.Role == models.RoleBackground && l.Position.Z < -0.1 {
						t.Errorf("background light %q Z=%.2f should be > 0 (behind subject)",
							l.ID, l.Position.Z)
					}
				}

				for _, panel := range p.Scene.Panels {
					if panel.Position.Distance <= 0 && panel.Position.Distance != 0.01 {
						t.Errorf("panel %q has non-positive distance: %f", panel.ID, panel.Position.Distance)
					}
				}
			})
		}
	}
}

// TestE2E_EVCalculationParsesShutterSpeed ensures the EV calculation correctly
// parses shutter speed strings like "1/200", "1/500", etc.
func TestE2E_EVCalculationParsesShutterSpeed(t *testing.T) {
	srv := setupServer(t)
	defer srv.Close()

	makeScene := func(shutter string) models.Scene {
		return models.Scene{
			Lights: []models.Light{
				{ID: "key", Role: models.RoleKey, Enabled: true, Type: models.LightTypeStrobe,
					Modifier: models.ModifierSoftbox,
					Position: models.Position3D{X: -1.41, Y: 0.5, Z: -1.41, Distance: 2.0, Angle: -135},
					Power:    75, ColorTemp: 5500},
			},
			Camera: models.CameraSettings{Aperture: 2.8, ISO: 100, ShutterSpeed: shutter},
		}
	}

	// 1/200 should give a different EV than 1/60
	scene200 := makeScene("1/200")
	scene60 := makeScene("1/60")

	body200, _ := json.Marshal(scene200)
	resp200 := doPost(t, srv.URL+"/api/analyze", body200)
	var a200 lighting.SceneAnalysis
	_ = json.NewDecoder(resp200.Body).Decode(&a200)
	_ = resp200.Body.Close()

	body60, _ := json.Marshal(scene60)
	resp60 := doPost(t, srv.URL+"/api/analyze", body60)
	var a60 lighting.SceneAnalysis
	_ = json.NewDecoder(resp60.Body).Decode(&a60)
	_ = resp60.Body.Close()

	// 1/200 is faster → higher EV; 1/60 is slower → lower EV
	if a200.OverallEV <= a60.OverallEV {
		t.Errorf("1/200 EV (%.1f) should be > 1/60 EV (%.1f)", a200.OverallEV, a60.OverallEV)
	}

	// Both should be non-zero
	if a200.OverallEV == 0 {
		t.Error("1/200 EV should be non-zero")
	}
	if a60.OverallEV == 0 {
		t.Error("1/60 EV should be non-zero")
	}
}

// TestE2E_RealStudioRembrandt tests a complete Rembrandt lighting setup with
// realistic studio positions and validates all expected physics outcomes.
func TestE2E_RealStudioRembrandt(t *testing.T) {
	srv := setupServer(t)
	defer srv.Close()

	// Fetch the actual Rembrandt preset from the API
	resp := doGet(t, srv.URL+"/api/presets/rembrandt")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var preset models.Preset
	_ = json.NewDecoder(resp.Body).Decode(&preset)
	_ = resp.Body.Close()

	// Analyze
	body, _ := json.Marshal(preset.Scene)
	analyzeResp := doPost(t, srv.URL+"/api/analyze", body)
	var analysis lighting.SceneAnalysis
	_ = json.NewDecoder(analyzeResp.Body).Decode(&analysis)
	_ = analyzeResp.Body.Close()

	// Rembrandt must have:
	// 1. Rectangular catchlight (softbox key)
	if analysis.CatchlightType != "rectangular" {
		t.Errorf("Rembrandt: expected rectangular catchlight, got %q", analysis.CatchlightType)
	}

	// 2. Key light should be the brightest contributor
	var keyIntensity, maxIntensity float64
	for _, c := range analysis.Contributions {
		if c.Role == "key" {
			keyIntensity = c.Intensity
		}
		if c.Intensity > maxIntensity {
			maxIntensity = c.Intensity
		}
	}
	if keyIntensity < maxIntensity {
		t.Errorf("Rembrandt: key (%.1f) should be brightest (max=%.1f)", keyIntensity, maxIntensity)
	}

	// 3. Key-to-fill ratio should be positive (panels modify effective fill)
	if analysis.KeyToFillRatio <= 0 {
		t.Errorf("Rembrandt: key:fill ratio %.2f should be positive", analysis.KeyToFillRatio)
	}

	// 4. Negative fill panel should absorb light
	foundNegEffect := false
	for _, pe := range analysis.PanelEffects {
		if pe.Type == string(models.PanelNegativeFill) && pe.EffectIntensity < 0 {
			foundNegEffect = true
		}
	}
	if !foundNegEffect {
		t.Error("Rembrandt: black V-flat should produce negative panel effect")
	}

	// 5. Bounce panel should add light
	foundBounceEffect := false
	for _, pe := range analysis.PanelEffects {
		if pe.Type == string(models.PanelBounceWhite) && pe.EffectIntensity > 0 {
			foundBounceEffect = true
		}
	}
	if !foundBounceEffect {
		t.Error("Rembrandt: chin bounce should produce positive panel effect")
	}
}

// TestE2E_RealStudioSplitLight tests a split lighting setup matching studio practice.
func TestE2E_RealStudioSplitLight(t *testing.T) {
	srv := setupServer(t)
	defer srv.Close()

	resp := doGet(t, srv.URL+"/api/presets/split")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var preset models.Preset
	_ = json.NewDecoder(resp.Body).Decode(&preset)
	_ = resp.Body.Close()

	body, _ := json.Marshal(preset.Scene)
	analyzeResp := doPost(t, srv.URL+"/api/analyze", body)
	var analysis lighting.SceneAnalysis
	_ = json.NewDecoder(analyzeResp.Body).Decode(&analysis)
	_ = analyzeResp.Body.Close()

	// Split light: single key at 90°, no fill light
	if len(analysis.Contributions) != 1 {
		t.Errorf("split: expected 1 contribution, got %d", len(analysis.Contributions))
	}

	// Key should be the only light with high intensity
	if analysis.Contributions[0].Role != "key" {
		t.Errorf("split: expected key light, got %q", analysis.Contributions[0].Role)
	}

	// Should have both negative fill and flag effects
	var hasNeg, hasFlag bool
	for _, pe := range analysis.PanelEffects {
		if pe.Type == string(models.PanelNegativeFill) {
			hasNeg = true
			if pe.EffectIntensity >= 0 {
				t.Errorf("split: V-flat negative fill should absorb light, got %f", pe.EffectIntensity)
			}
		}
		if pe.Type == string(models.PanelFlag) {
			hasFlag = true
			if pe.EffectIntensity >= 0 {
				t.Errorf("split: flag should block light, got %f", pe.EffectIntensity)
			}
		}
	}
	if !hasNeg {
		t.Error("split: expected negative fill effect")
	}
	if !hasFlag {
		t.Error("split: expected flag effect")
	}
}

// TestE2E_RealStudioHighKey tests a high-key portrait setup with background lights.
func TestE2E_RealStudioHighKey(t *testing.T) {
	srv := setupServer(t)
	defer srv.Close()

	resp := doGet(t, srv.URL+"/api/presets/high_key")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var preset models.Preset
	_ = json.NewDecoder(resp.Body).Decode(&preset)
	_ = resp.Body.Close()

	body, _ := json.Marshal(preset.Scene)
	analyzeResp := doPost(t, srv.URL+"/api/analyze", body)
	var analysis lighting.SceneAnalysis
	_ = json.NewDecoder(analyzeResp.Body).Decode(&analysis)
	_ = analyzeResp.Body.Close()

	// High key should have: key + fill + 2 background lights = 4
	if len(analysis.Contributions) != 4 {
		t.Errorf("high key: expected 4 contributions, got %d", len(analysis.Contributions))
	}

	// Shadow quality: octabox(0.85) + umbrella(0.65) + 2 bare(0.1) avg → medium
	if analysis.ShadowQuality != "medium" && analysis.ShadowQuality != "soft" {
		t.Errorf("high key: expected medium or soft shadows, got %q", analysis.ShadowQuality)
	}

	// Both bounce panels should add positive intensity
	for _, pe := range analysis.PanelEffects {
		if pe.Type == string(models.PanelBounceWhite) && pe.EffectIntensity <= 0 {
			t.Errorf("high key: bounce panel %q should add positive fill, got %f",
				pe.PanelID, pe.EffectIntensity)
		}
	}
}

// TestE2E_RealStudioLowKey tests a low-key dramatic setup.
func TestE2E_RealStudioLowKey(t *testing.T) {
	srv := setupServer(t)
	defer srv.Close()

	resp := doGet(t, srv.URL+"/api/presets/low_key")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var preset models.Preset
	_ = json.NewDecoder(resp.Body).Decode(&preset)
	_ = resp.Body.Close()

	body, _ := json.Marshal(preset.Scene)
	analyzeResp := doPost(t, srv.URL+"/api/analyze", body)
	var analysis lighting.SceneAnalysis
	_ = json.NewDecoder(analyzeResp.Body).Decode(&analysis)
	_ = analyzeResp.Body.Close()

	// Low key: single gridded key
	if len(analysis.Contributions) != 1 {
		t.Errorf("low key: expected 1 contribution, got %d", len(analysis.Contributions))
	}

	// Shadow quality should be hard (honeycomb grid)
	if analysis.ShadowQuality != "hard" {
		t.Errorf("low key: expected hard shadows, got %q", analysis.ShadowQuality)
	}

	// Catchlight should be point (honeycomb grid)
	if analysis.CatchlightType != "point" {
		t.Errorf("low key: expected point catchlight, got %q", analysis.CatchlightType)
	}

	// All panel effects should be negative (absorbing/blocking)
	for _, pe := range analysis.PanelEffects {
		if pe.EffectIntensity > 0 {
			t.Errorf("low key: panel %q should not add light, got %f", pe.PanelID, pe.EffectIntensity)
		}
	}
}

// TestE2E_RealStudioOutdoorGoldenHour tests golden hour setup with sun.
func TestE2E_RealStudioOutdoorGoldenHour(t *testing.T) {
	srv := setupServer(t)
	defer srv.Close()

	resp := doGet(t, srv.URL+"/api/presets/outdoor_golden_hour")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var preset models.Preset
	_ = json.NewDecoder(resp.Body).Decode(&preset)
	_ = resp.Body.Close()

	body, _ := json.Marshal(preset.Scene)
	analyzeResp := doPost(t, srv.URL+"/api/analyze", body)
	var analysis lighting.SceneAnalysis
	_ = json.NewDecoder(analyzeResp.Body).Decode(&analysis)
	_ = analyzeResp.Body.Close()

	// Should have sun as the key light
	hasSun := false
	for _, c := range analysis.Contributions {
		if c.SpillAngle == 180 {
			hasSun = true
			if c.Intensity < 10 {
				t.Errorf("golden hour sun intensity %f too low", c.Intensity)
			}
		}
	}
	if !hasSun {
		t.Error("golden hour: expected sun light contribution")
	}

	// Silver reflector should provide bounce fill
	hasBounce := false
	for _, pe := range analysis.PanelEffects {
		if pe.EffectIntensity > 0 {
			hasBounce = true
		}
	}
	if !hasBounce {
		t.Error("golden hour: expected positive bounce from reflector")
	}
}
