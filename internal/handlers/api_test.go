package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/srivickynesh/light-simulator/internal/lighting"
	"github.com/srivickynesh/light-simulator/internal/models"
)

func newTestRequest(t *testing.T, method, target string, body *bytes.Reader) *http.Request {
	t.Helper()
	if body == nil {
		body = bytes.NewReader(nil)
	}
	return httptest.NewRequestWithContext(context.Background(), method, target, body)
}

func TestHealthEndpoint(t *testing.T) {
	api := NewAPI()
	req := newTestRequest(t, http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()

	api.Health(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["status"] != "ok" {
		t.Errorf("expected status 'ok', got %q", resp["status"])
	}
}

func TestHealthReturnsJSON(t *testing.T) {
	api := NewAPI()
	req := newTestRequest(t, http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()
	api.Health(w, req)

	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
}

func TestListPresets(t *testing.T) {
	api := NewAPI()
	req := newTestRequest(t, http.MethodGet, "/api/presets", nil)
	w := httptest.NewRecorder()

	api.ListPresets(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var presets map[string]json.RawMessage
	if err := json.NewDecoder(w.Body).Decode(&presets); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(presets) == 0 {
		t.Error("expected at least one preset category")
	}
}

func TestListPresetsContainsCategories(t *testing.T) {
	api := NewAPI()
	req := newTestRequest(t, http.MethodGet, "/api/presets", nil)
	w := httptest.NewRecorder()
	api.ListPresets(w, req)

	var categories map[string][]models.Preset
	if err := json.NewDecoder(w.Body).Decode(&categories); err != nil {
		t.Fatalf("decode error: %v", err)
	}

	for _, cat := range []string{"portrait", "product", "fashion", "food", "headshot", "group", "sport"} {
		if _, ok := categories[cat]; !ok {
			t.Errorf("missing category %q", cat)
		}
	}
}

func TestGetPresetNotFound(t *testing.T) {
	api := NewAPI()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/presets/{id}", api.GetPreset)

	req := newTestRequest(t, http.MethodGet, "/api/presets/nonexistent", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestGetPresetFound(t *testing.T) {
	api := NewAPI()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/presets/{id}", api.GetPreset)

	req := newTestRequest(t, http.MethodGet, "/api/presets/rembrandt", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestGetPresetReturnsCorrectID(t *testing.T) {
	api := NewAPI()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/presets/{id}", api.GetPreset)

	for _, id := range []string{"rembrandt", "butterfly", "beauty_ring", "sport_action", "product_glass"} {
		req := newTestRequest(t, http.MethodGet, "/api/presets/"+id, nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("preset %q: expected 200, got %d", id, w.Code)
			continue
		}

		var preset models.Preset
		if err := json.NewDecoder(w.Body).Decode(&preset); err != nil {
			t.Errorf("preset %q: decode error: %v", id, err)
			continue
		}
		if preset.ID != id {
			t.Errorf("expected preset ID %q, got %q", id, preset.ID)
		}
	}
}

func TestAnalyzeEndpoint(t *testing.T) {
	api := NewAPI()
	scene := map[string]any{
		"lights": []map[string]any{
			{
				"id": "key", "role": "key", "enabled": true,
				"modifier": "softbox", "power": 70, "color_temp": 5500,
				"position": map[string]any{"x": -1.5, "y": 0.5, "z": 1.5, "distance": 2.0, "angle": 45},
			},
		},
		"camera": map[string]any{"aperture": 2.8, "iso": 100, "shutter_speed": "1/200"},
	}

	body, _ := json.Marshal(scene)
	req := newTestRequest(t, http.MethodPost, "/api/analyze", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	api.AnalyzeScene(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestAnalyzeReturnsAnalysis(t *testing.T) {
	api := NewAPI()
	scene := models.Scene{
		Lights: []models.Light{
			{ID: "key", Role: models.RoleKey, Enabled: true, Modifier: models.ModifierSoftbox,
				Position: models.Position3D{X: -1, Z: 1, Distance: 2}, Power: 70, ColorTemp: 5500},
		},
		Camera: models.CameraSettings{Aperture: 2.8, ISO: 100, ShutterSpeed: "1/200"},
	}

	body, _ := json.Marshal(scene)
	req := newTestRequest(t, http.MethodPost, "/api/analyze", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	api.AnalyzeScene(w, req)

	var analysis lighting.SceneAnalysis
	if err := json.NewDecoder(w.Body).Decode(&analysis); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(analysis.Contributions) != 1 {
		t.Errorf("expected 1 contribution, got %d", len(analysis.Contributions))
	}
	if analysis.OverallEV == 0 {
		t.Error("expected non-zero EV")
	}
	if analysis.CatchlightType == "" {
		t.Error("expected non-empty catchlight type")
	}
}

func TestAnalyzeInvalidJSON(t *testing.T) {
	api := NewAPI()
	req := newTestRequest(t, http.MethodPost, "/api/analyze", bytes.NewReader([]byte("not json")))
	w := httptest.NewRecorder()

	api.AnalyzeScene(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAnalyzeEmptyBody(t *testing.T) {
	api := NewAPI()
	req := newTestRequest(t, http.MethodPost, "/api/analyze", bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	api.AnalyzeScene(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for empty scene, got %d", w.Code)
	}
}

func TestFlashGuideEndpoint(t *testing.T) {
	api := NewAPI()
	req := newTestRequest(t, http.MethodGet, "/api/guides/flash", nil)
	w := httptest.NewRecorder()

	api.FlashGuide(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestFlashGuideReturnsData(t *testing.T) {
	api := NewAPI()
	req := newTestRequest(t, http.MethodGet, "/api/guides/flash", nil)
	w := httptest.NewRecorder()
	api.FlashGuide(w, req)

	var guides []json.RawMessage
	if err := json.NewDecoder(w.Body).Decode(&guides); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(guides) == 0 {
		t.Error("expected at least one flash guide")
	}
}

func TestModifierGuideEndpoint(t *testing.T) {
	api := NewAPI()
	req := newTestRequest(t, http.MethodGet, "/api/guides/modifiers", nil)
	w := httptest.NewRecorder()

	api.ModifierGuide(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestModifierGuideReturnsData(t *testing.T) {
	api := NewAPI()
	req := newTestRequest(t, http.MethodGet, "/api/guides/modifiers", nil)
	w := httptest.NewRecorder()
	api.ModifierGuide(w, req)

	var guides []json.RawMessage
	if err := json.NewDecoder(w.Body).Decode(&guides); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(guides) == 0 {
		t.Error("expected at least one modifier guide")
	}
}

func TestLensGuideEndpoint(t *testing.T) {
	api := NewAPI()
	req := newTestRequest(t, http.MethodGet, "/api/guides/lenses", nil)
	w := httptest.NewRecorder()

	api.LensGuide(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestLensGuideReturnsData(t *testing.T) {
	api := NewAPI()
	req := newTestRequest(t, http.MethodGet, "/api/guides/lenses", nil)
	w := httptest.NewRecorder()
	api.LensGuide(w, req)

	var guides []json.RawMessage
	if err := json.NewDecoder(w.Body).Decode(&guides); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(guides) == 0 {
		t.Error("expected at least one lens guide")
	}
}

func TestAnalyzeWithPanels(t *testing.T) {
	api := NewAPI()
	scene := models.Scene{
		Lights: []models.Light{
			{ID: "key", Role: models.RoleKey, Enabled: true, Modifier: models.ModifierSoftbox,
				Position: models.Position3D{X: -1, Z: 1, Distance: 2, Angle: 45}, Power: 70, ColorTemp: 5500},
			{ID: "fill", Role: models.RoleFill, Enabled: true, Modifier: models.ModifierUmbrella,
				Position: models.Position3D{X: 1, Z: 2, Distance: 2.2, Angle: -25}, Power: 30, ColorTemp: 5500},
		},
		Panels: []models.Panel{
			{ID: "neg", Name: "Neg Fill", Type: models.PanelNegativeFill, Size: models.PanelSizeLarge,
				Position: models.Position3D{X: 1.5, Y: 0, Z: 0.5, Distance: 1.0, Angle: -90}, Enabled: true},
			{ID: "bounce", Name: "White Bounce", Type: models.PanelBounceWhite, Size: models.PanelSizeMedium,
				Position: models.Position3D{X: 0, Y: -0.5, Z: 1.0, Distance: 0.8, Angle: 0}, Enabled: true},
		},
		Camera: models.CameraSettings{Aperture: 2.8, ISO: 100, ShutterSpeed: "1/200"},
	}

	body, _ := json.Marshal(scene)
	req := newTestRequest(t, http.MethodPost, "/api/analyze", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	api.AnalyzeScene(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var analysis lighting.SceneAnalysis
	if err := json.NewDecoder(w.Body).Decode(&analysis); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(analysis.PanelEffects) != 2 {
		t.Errorf("expected 2 panel effects, got %d", len(analysis.PanelEffects))
	}
	if len(analysis.Contributions) != 2 {
		t.Errorf("expected 2 contributions, got %d", len(analysis.Contributions))
	}
}

func TestGetPresetReturnsPanels(t *testing.T) {
	api := NewAPI()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/presets/{id}", api.GetPreset)

	req := newTestRequest(t, http.MethodGet, "/api/presets/rembrandt", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var preset models.Preset
	if err := json.NewDecoder(w.Body).Decode(&preset); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(preset.Scene.Panels) == 0 {
		t.Error("rembrandt preset should have panels")
	}
}

func TestRegisterRoutes(t *testing.T) {
	api := NewAPI()
	mux := http.NewServeMux()
	api.RegisterRoutes(mux)

	endpoints := []struct {
		method string
		path   string
		want   int
	}{
		{http.MethodGet, "/api/health", http.StatusOK},
		{http.MethodGet, "/api/presets", http.StatusOK},
		{http.MethodGet, "/api/guides/flash", http.StatusOK},
		{http.MethodGet, "/api/guides/modifiers", http.StatusOK},
		{http.MethodGet, "/api/guides/lenses", http.StatusOK},
	}

	for _, ep := range endpoints {
		req := newTestRequest(t, ep.method, ep.path, nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != ep.want {
			t.Errorf("%s %s: expected %d, got %d", ep.method, ep.path, ep.want, w.Code)
		}
	}
}
