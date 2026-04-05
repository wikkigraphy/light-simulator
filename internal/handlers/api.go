package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/srivickynesh/light-simulator/internal/cheatsheet"
	"github.com/srivickynesh/light-simulator/internal/lighting"
	"github.com/srivickynesh/light-simulator/internal/models"
)

// API groups all JSON API handlers.
type API struct{}

// NewAPI creates a new API handler group.
func NewAPI() *API {
	return &API{}
}

// RegisterRoutes mounts all API endpoints on the given mux.
func (a *API) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/presets", a.ListPresets)
	mux.HandleFunc("GET /api/presets/{id}", a.GetPreset)
	mux.HandleFunc("POST /api/analyze", a.AnalyzeScene)
	mux.HandleFunc("GET /api/guides/flash", a.FlashGuide)
	mux.HandleFunc("GET /api/guides/modifiers", a.ModifierGuide)
	mux.HandleFunc("GET /api/guides/lenses", a.LensGuide)
	mux.HandleFunc("GET /api/health", a.Health)
}

// ListPresets returns all lighting presets grouped by category.
func (a *API) ListPresets(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, cheatsheet.PresetsByCategory())
}

// GetPreset returns a single preset by ID.
func (a *API) GetPreset(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	for _, p := range cheatsheet.AllPresets() {
		if p.ID == id {
			writeJSON(w, http.StatusOK, p)
			return
		}
	}
	writeJSON(w, http.StatusNotFound, map[string]string{"error": "preset not found"})
}

// AnalyzeScene accepts a Scene JSON body and returns computed lighting analysis.
func (a *API) AnalyzeScene(w http.ResponseWriter, r *http.Request) {
	defer func() { _ = r.Body.Close() }()

	var scene models.Scene
	if err := json.NewDecoder(r.Body).Decode(&scene); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid scene JSON: " + err.Error()})
		return
	}

	analysis := lighting.Analyze(&scene)
	writeJSON(w, http.StatusOK, analysis)
}

// FlashGuide returns the flash selection cheatsheet.
func (a *API) FlashGuide(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, cheatsheet.FlashGuides())
}

// ModifierGuide returns the modifier cheatsheet.
func (a *API) ModifierGuide(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, cheatsheet.ModifierGuides())
}

// LensGuide returns the lens selection cheatsheet.
func (a *API) LensGuide(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, cheatsheet.LensGuides())
}

// Health returns a simple health check response.
func (a *API) Health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
