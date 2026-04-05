package lighting

import (
	"math"
	"strings"
	"testing"

	"github.com/srivickynesh/light-simulator/internal/models"
)

func TestAnalyzeEmptyScene(t *testing.T) {
	scene := &models.Scene{
		Lights: []models.Light{},
		Camera: models.CameraSettings{
			Aperture: 2.8, ISO: 100, ShutterSpeed: "1/200",
		},
	}

	result := Analyze(scene)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Contributions) != 0 {
		t.Errorf("expected 0 contributions, got %d", len(result.Contributions))
	}
}

func TestAnalyzeKeyFillSetup(t *testing.T) {
	scene := &models.Scene{
		Lights: []models.Light{
			{
				ID: "key", Role: models.RoleKey, Enabled: true,
				Modifier: models.ModifierSoftbox,
				Position: models.Position3D{X: -1.5, Y: 0.5, Z: 1.5, Distance: 2.0, Angle: 45},
				Power:    80, ColorTemp: 5500,
			},
			{
				ID: "fill", Role: models.RoleFill, Enabled: true,
				Modifier: models.ModifierUmbrella,
				Position: models.Position3D{X: 1.0, Y: 0, Z: 2.0, Distance: 2.2, Angle: -25},
				Power:    30, ColorTemp: 5500,
			},
		},
		Camera: models.CameraSettings{
			Aperture: 2.8, ISO: 100, ShutterSpeed: "1/200",
		},
	}

	result := Analyze(scene)

	if len(result.Contributions) != 2 {
		t.Fatalf("expected 2 contributions, got %d", len(result.Contributions))
	}

	if result.KeyToFillRatio <= 0 {
		t.Error("expected positive key-to-fill ratio")
	}

	if result.ShadowQuality == "" {
		t.Error("expected non-empty shadow quality")
	}

	if result.CatchlightType != "rectangular" {
		t.Errorf("expected rectangular catchlight for softbox key, got %q", result.CatchlightType)
	}
}

func TestAnalyzeDisabledLight(t *testing.T) {
	scene := &models.Scene{
		Lights: []models.Light{
			{
				ID: "off", Role: models.RoleKey, Enabled: false,
				Modifier: models.ModifierSoftbox,
				Position: models.Position3D{Distance: 2.0},
				Power:    80, ColorTemp: 5500,
			},
		},
		Camera: models.CameraSettings{Aperture: 2.8, ISO: 100},
	}

	result := Analyze(scene)
	if len(result.Contributions) != 0 {
		t.Errorf("disabled lights should not contribute; got %d contributions", len(result.Contributions))
	}
}

func TestModifierSoftness(t *testing.T) {
	cases := []struct {
		mod  models.ModifierType
		want float64
	}{
		{models.ModifierSnoot, 0.05},
		{models.ModifierSoftbox, 0.75},
		{models.ModifierOctabox, 0.85},
		{models.ModifierDiffusion, 0.9},
		{models.ModifierNone, 0.1},
		{models.ModifierHoneycomb, 0.15},
		{models.ModifierBarnDoors, 0.1},
		{models.ModifierReflector, 0.3},
		{models.ModifierBeautyDish, 0.5},
		{models.ModifierUmbrella, 0.65},
		{models.ModifierStripbox, 0.7},
		{models.ModifierParabolic, 0.6},
	}

	for _, tc := range cases {
		got := modifierSoftness(tc.mod)
		if got != tc.want {
			t.Errorf("modifierSoftness(%q) = %f, want %f", tc.mod, got, tc.want)
		}
	}
}

func TestModifierSoftnessUnknown(t *testing.T) {
	got := modifierSoftness("unknown_modifier")
	if got != 0.5 {
		t.Errorf("unknown modifier softness = %f, want 0.5", got)
	}
}

func TestModifierSpill(t *testing.T) {
	cases := []struct {
		mod  models.ModifierType
		grid int
		want float64
	}{
		{models.ModifierNone, 0, 180},
		{models.ModifierSnoot, 0, 15},
		{models.ModifierBarnDoors, 0, 40},
		{models.ModifierHoneycomb, 0, 30},
		{models.ModifierHoneycomb, 20, 20},
		{models.ModifierHoneycomb, 40, 40},
		{models.ModifierSoftbox, 0, 90},
		{models.ModifierStripbox, 0, 60},
		{models.ModifierOctabox, 0, 100},
		{models.ModifierParabolic, 0, 50},
	}

	for _, tc := range cases {
		got := modifierSpill(tc.mod, tc.grid)
		if got != tc.want {
			t.Errorf("modifierSpill(%q, %d) = %f, want %f", tc.mod, tc.grid, got, tc.want)
		}
	}
}

func TestModifierSpillUnknown(t *testing.T) {
	got := modifierSpill("unknown", 0)
	if got != 90 {
		t.Errorf("unknown modifier spill = %f, want 90", got)
	}
}

func TestWarningsNoKey(t *testing.T) {
	scene := &models.Scene{
		Lights: []models.Light{
			{ID: "fill", Role: models.RoleFill, Enabled: true, Position: models.Position3D{Distance: 2}, Power: 50, ColorTemp: 5500},
		},
		Camera: models.CameraSettings{Aperture: 2.8, ISO: 100},
	}

	result := Analyze(scene)
	found := false
	for _, w := range result.Warnings {
		if strings.Contains(w, "key light") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected warning about missing key light")
	}
}

func TestWarningsHighRatio(t *testing.T) {
	scene := &models.Scene{
		Lights: []models.Light{
			{ID: "key", Role: models.RoleKey, Enabled: true, Modifier: models.ModifierSoftbox,
				Position: models.Position3D{Distance: 0.5}, Power: 100, ColorTemp: 5500},
			{ID: "fill", Role: models.RoleFill, Enabled: true, Modifier: models.ModifierUmbrella,
				Position: models.Position3D{Distance: 5}, Power: 10, ColorTemp: 5500},
		},
		Camera: models.CameraSettings{Aperture: 2.8, ISO: 100},
	}

	result := Analyze(scene)
	found := false
	for _, w := range result.Warnings {
		if strings.Contains(w, "ratio") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected warning about high key-to-fill ratio")
	}
}

func TestWarningsColorTempMismatch(t *testing.T) {
	scene := &models.Scene{
		Lights: []models.Light{
			{ID: "key", Role: models.RoleKey, Enabled: true, Modifier: models.ModifierSoftbox,
				Position: models.Position3D{Distance: 2}, Power: 70, ColorTemp: 3200},
			{ID: "fill", Role: models.RoleFill, Enabled: true, Modifier: models.ModifierUmbrella,
				Position: models.Position3D{Distance: 2}, Power: 30, ColorTemp: 6500},
		},
		Camera: models.CameraSettings{Aperture: 2.8, ISO: 100},
	}

	result := Analyze(scene)
	found := false
	for _, w := range result.Warnings {
		if strings.Contains(w, "temperature") || strings.Contains(w, "mismatch") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected warning about color temperature mismatch (3300K spread)")
	}
}

func TestWarningsHighIntensity(t *testing.T) {
	scene := &models.Scene{
		Lights: []models.Light{
			{ID: "key", Role: models.RoleKey, Enabled: true, Modifier: models.ModifierNone,
				Position: models.Position3D{Distance: 0.1}, Power: 100, ColorTemp: 5500},
		},
		Camera: models.CameraSettings{Aperture: 2.8, ISO: 100},
	}

	result := Analyze(scene)
	found := false
	for _, w := range result.Warnings {
		if strings.Contains(w, "intensity") || strings.Contains(w, "blown") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected warning about high intensity")
	}
}

func TestNoWarningsCleanSetup(t *testing.T) {
	scene := &models.Scene{
		Lights: []models.Light{
			{ID: "key", Role: models.RoleKey, Enabled: true, Modifier: models.ModifierSoftbox,
				Position: models.Position3D{Distance: 2}, Power: 70, ColorTemp: 5500},
			{ID: "fill", Role: models.RoleFill, Enabled: true, Modifier: models.ModifierUmbrella,
				Position: models.Position3D{Distance: 2.5}, Power: 30, ColorTemp: 5500},
		},
		Camera: models.CameraSettings{Aperture: 2.8, ISO: 100},
	}

	result := Analyze(scene)
	if len(result.Warnings) > 0 {
		t.Errorf("expected no warnings for clean setup, got: %v", result.Warnings)
	}
}

func TestCSSFiltersGeneration(t *testing.T) {
	scene := &models.Scene{
		Lights: []models.Light{
			{
				ID: "key", Role: models.RoleKey, Enabled: true,
				Modifier: models.ModifierOctabox,
				Position: models.Position3D{X: -1, Y: 0.5, Z: 1.5, Distance: 2, Angle: 30},
				Power:    70, ColorTemp: 5500,
			},
		},
		Camera: models.CameraSettings{Aperture: 2.8, ISO: 100},
	}

	result := Analyze(scene)
	filters := result.CSSFilters

	if filters.Brightness <= 0 || filters.Brightness > 2 {
		t.Errorf("brightness out of range: %f", filters.Brightness)
	}
	if filters.ShadowGradient == "" {
		t.Error("expected non-empty shadow gradient")
	}
	if filters.HighlightPos == "" {
		t.Error("expected non-empty highlight position")
	}
}

func TestCSSFiltersWarmColorTemp(t *testing.T) {
	scene := &models.Scene{
		Lights: []models.Light{
			{ID: "key", Role: models.RoleKey, Enabled: true, Modifier: models.ModifierSoftbox,
				Position: models.Position3D{Distance: 2}, Power: 70, ColorTemp: 3200},
		},
		Camera: models.CameraSettings{Aperture: 2.8, ISO: 100},
	}

	result := Analyze(scene)
	if result.CSSFilters.WarmthShift >= 0 {
		t.Error("expected negative warmth shift for warm color temp")
	}
}

func TestCSSFiltersCoolColorTemp(t *testing.T) {
	scene := &models.Scene{
		Lights: []models.Light{
			{ID: "key", Role: models.RoleKey, Enabled: true, Modifier: models.ModifierSoftbox,
				Position: models.Position3D{Distance: 2}, Power: 70, ColorTemp: 7500},
		},
		Camera: models.CameraSettings{Aperture: 2.8, ISO: 100},
	}

	result := Analyze(scene)
	if result.CSSFilters.WarmthShift <= 0 {
		t.Error("expected positive warmth shift for cool color temp")
	}
}

func TestCSSFiltersContrastWithHighRatio(t *testing.T) {
	scene := &models.Scene{
		Lights: []models.Light{
			{ID: "key", Role: models.RoleKey, Enabled: true, Modifier: models.ModifierSoftbox,
				Position: models.Position3D{Distance: 1}, Power: 80, ColorTemp: 5500},
			{ID: "fill", Role: models.RoleFill, Enabled: true, Modifier: models.ModifierUmbrella,
				Position: models.Position3D{Distance: 3}, Power: 20, ColorTemp: 5500},
		},
		Camera: models.CameraSettings{Aperture: 2.8, ISO: 100},
	}

	result := Analyze(scene)
	if result.CSSFilters.Contrast <= 1.0 {
		t.Errorf("expected contrast > 1.0 with key:fill ratio, got %f", result.CSSFilters.Contrast)
	}
}

func TestClassifyShadows(t *testing.T) {
	cases := []struct {
		name     string
		modifier models.ModifierType
		want     string
	}{
		{"hard_snoot", models.ModifierSnoot, "hard"},
		{"hard_none", models.ModifierNone, "hard"},
		{"medium_beauty_dish", models.ModifierBeautyDish, "medium"},
		{"soft_octabox", models.ModifierOctabox, "soft"},
		{"soft_diffusion", models.ModifierDiffusion, "soft"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			scene := &models.Scene{
				Lights: []models.Light{
					{ID: "key", Role: models.RoleKey, Enabled: true, Modifier: tc.modifier,
						Position: models.Position3D{Distance: 2}, Power: 70, ColorTemp: 5500},
				},
				Camera: models.CameraSettings{Aperture: 2.8, ISO: 100},
			}
			result := Analyze(scene)
			if result.ShadowQuality != tc.want {
				t.Errorf("expected shadow quality %q for %q, got %q", tc.want, tc.modifier, result.ShadowQuality)
			}
		})
	}
}

func TestCatchlightTypes(t *testing.T) {
	cases := []struct {
		modifier models.ModifierType
		want     string
	}{
		{models.ModifierOctabox, "octagonal"},
		{models.ModifierSoftbox, "rectangular"},
		{models.ModifierStripbox, "rectangular"},
		{models.ModifierBeautyDish, "circular_ring"},
		{models.ModifierUmbrella, "circular"},
		{models.ModifierParabolic, "parabolic"},
		{models.ModifierNone, "point"},
		{models.ModifierSnoot, "point"},
	}

	for _, tc := range cases {
		t.Run(string(tc.modifier), func(t *testing.T) {
			scene := &models.Scene{
				Lights: []models.Light{
					{ID: "key", Role: models.RoleKey, Enabled: true, Modifier: tc.modifier,
						Position: models.Position3D{Distance: 2}, Power: 70, ColorTemp: 5500},
				},
			}
			got := determineCatchlight(scene)
			if got != tc.want {
				t.Errorf("determineCatchlight(%q) = %q, want %q", tc.modifier, got, tc.want)
			}
		})
	}
}

func TestCatchlightNoKey(t *testing.T) {
	scene := &models.Scene{
		Lights: []models.Light{
			{ID: "fill", Role: models.RoleFill, Enabled: true, Modifier: models.ModifierSoftbox,
				Position: models.Position3D{Distance: 2}, Power: 50, ColorTemp: 5500},
		},
	}
	got := determineCatchlight(scene)
	if got != "none" {
		t.Errorf("expected 'none' when no key light, got %q", got)
	}
}

func TestComputeContributionInverseSquare(t *testing.T) {
	lightClose := &models.Light{
		ID: "close", Role: models.RoleKey, Enabled: true, Modifier: models.ModifierSoftbox,
		Position: models.Position3D{Distance: 1.0}, Power: 100, ColorTemp: 5500,
	}
	lightFar := &models.Light{
		ID: "far", Role: models.RoleKey, Enabled: true, Modifier: models.ModifierSoftbox,
		Position: models.Position3D{Distance: 2.0}, Power: 100, ColorTemp: 5500,
	}

	closeContrib := computeContribution(lightClose)
	farContrib := computeContribution(lightFar)

	expectedRatio := 4.0
	actualRatio := closeContrib.Intensity / farContrib.Intensity
	if math.Abs(actualRatio-expectedRatio) > 0.01 {
		t.Errorf("expected intensity ratio %.1f (inverse square), got %.2f", expectedRatio, actualRatio)
	}
}

func TestComputeContributionMinDistance(t *testing.T) {
	light := &models.Light{
		ID: "zero", Role: models.RoleKey, Enabled: true, Modifier: models.ModifierSoftbox,
		Position: models.Position3D{Distance: 0.0}, Power: 100, ColorTemp: 5500,
	}

	contrib := computeContribution(light)
	if math.IsInf(contrib.Intensity, 0) || math.IsNaN(contrib.Intensity) {
		t.Error("distance=0 should clamp to 0.1, not produce Inf/NaN")
	}
}

func TestComputeContributionFeathered(t *testing.T) {
	light := &models.Light{
		ID: "f", Role: models.RoleKey, Enabled: true, Modifier: models.ModifierSoftbox,
		Position: models.Position3D{Distance: 2.0}, Power: 70, ColorTemp: 5500,
		Feathered: true,
	}
	contrib := computeContribution(light)
	baseSoftness := modifierSoftness(models.ModifierSoftbox)
	if contrib.Softness <= baseSoftness {
		t.Errorf("feathered should increase softness beyond base %f, got %f", baseSoftness, contrib.Softness)
	}
}

func TestComputeExposureValue(t *testing.T) {
	scene := &models.Scene{
		Camera: models.CameraSettings{
			Aperture: 2.8, ISO: 100, ShutterSpeed: "1/200",
		},
		Ambient: 0.1,
	}

	ev := computeExposureValue(scene)
	if ev < 3 || ev > 20 {
		t.Errorf("EV=%f seems out of reasonable range", ev)
	}
}

func TestComputeExposureValueShutterSpeeds(t *testing.T) {
	makeScene := func(shutter string) *models.Scene {
		return &models.Scene{
			Camera: models.CameraSettings{
				Aperture: 2.8, ISO: 100, ShutterSpeed: shutter,
			},
		}
	}

	ev200 := computeExposureValue(makeScene("1/200"))
	ev60 := computeExposureValue(makeScene("1/60"))
	ev500 := computeExposureValue(makeScene("1/500"))
	evEmpty := computeExposureValue(makeScene(""))

	// Faster shutter = higher EV
	if ev500 <= ev200 {
		t.Errorf("1/500 EV (%.1f) should be > 1/200 EV (%.1f)", ev500, ev200)
	}
	if ev200 <= ev60 {
		t.Errorf("1/200 EV (%.1f) should be > 1/60 EV (%.1f)", ev200, ev60)
	}

	// Empty shutter speed should default to 1/200
	if math.Abs(evEmpty-ev200) > 0.1 {
		t.Errorf("empty shutter should default to 1/200; empty=%.1f, 1/200=%.1f", evEmpty, ev200)
	}
}

func TestComputeExposureValueDefaultAperture(t *testing.T) {
	scene := &models.Scene{
		Camera: models.CameraSettings{
			Aperture: 0, ISO: 100,
		},
	}
	ev := computeExposureValue(scene)
	if ev == 0 {
		t.Error("should use default aperture when 0 provided")
	}
}

func TestComputeExposureValueLowISO(t *testing.T) {
	scene := &models.Scene{
		Camera: models.CameraSettings{
			Aperture: 2.8, ISO: 0,
		},
	}
	ev := computeExposureValue(scene)
	if ev == 0 {
		t.Error("should use default ISO when 0 provided")
	}
}

func TestFormatFloat(t *testing.T) {
	cases := []struct {
		in   float64
		want string
	}{
		{0, "0"},
		{1, "1"},
		{100, "100"},
		{1.5, "1.50"},
		{3.14159, "3.14"},
		{0.15, "0.15"},
	}

	for _, tc := range cases {
		got := formatFloat(tc.in)
		if got != tc.want {
			t.Errorf("formatFloat(%f) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestAnalyzeMultipleLights(t *testing.T) {
	scene := &models.Scene{
		Lights: []models.Light{
			{ID: "key", Role: models.RoleKey, Enabled: true, Modifier: models.ModifierSoftbox,
				Position: models.Position3D{X: -1, Z: 1, Distance: 2}, Power: 80, ColorTemp: 5500},
			{ID: "fill", Role: models.RoleFill, Enabled: true, Modifier: models.ModifierUmbrella,
				Position: models.Position3D{X: 1, Z: 1, Distance: 2}, Power: 30, ColorTemp: 5500},
			{ID: "rim", Role: models.RoleRim, Enabled: true, Modifier: models.ModifierStripbox,
				Position: models.Position3D{X: 1, Z: -1, Distance: 2}, Power: 50, ColorTemp: 5500},
			{ID: "hair", Role: models.RoleHair, Enabled: true, Modifier: models.ModifierHoneycomb,
				Position: models.Position3D{X: 0, Z: -1, Distance: 2}, Power: 40, ColorTemp: 5500, GridDegree: 20},
		},
		Camera: models.CameraSettings{Aperture: 2.8, ISO: 100},
	}

	result := Analyze(scene)
	if len(result.Contributions) != 4 {
		t.Errorf("expected 4 contributions, got %d", len(result.Contributions))
	}
	if result.OverallEV == 0 {
		t.Error("expected non-zero EV")
	}
}

// --- Panel physics tests ---

func baseSceneWithPanels(panels []models.Panel) *models.Scene {
	return &models.Scene{
		Lights: []models.Light{
			{ID: "key", Role: models.RoleKey, Enabled: true, Modifier: models.ModifierSoftbox,
				Position: models.Position3D{X: -1.5, Z: 1.5, Distance: 2, Angle: 45}, Power: 80, ColorTemp: 5500},
			{ID: "fill", Role: models.RoleFill, Enabled: true, Modifier: models.ModifierUmbrella,
				Position: models.Position3D{X: 1, Z: 2, Distance: 2.2, Angle: -25}, Power: 30, ColorTemp: 5500},
		},
		Panels: panels,
		Camera: models.CameraSettings{Aperture: 2.8, ISO: 100, ShutterSpeed: "1/200"},
	}
}

func TestBounceWhitePanelAddsIntensity(t *testing.T) {
	panels := []models.Panel{
		{ID: "wb", Name: "White Bounce", Type: models.PanelBounceWhite, Size: models.PanelSizeLarge,
			Position: models.Position3D{X: 1, Y: 0, Z: 0.5, Distance: 1.0, Angle: -90}, Enabled: true},
	}
	result := Analyze(baseSceneWithPanels(panels))

	if len(result.PanelEffects) != 1 {
		t.Fatalf("expected 1 panel effect, got %d", len(result.PanelEffects))
	}
	pe := result.PanelEffects[0]
	if pe.EffectIntensity <= 0 {
		t.Errorf("white bounce should add positive intensity, got %f", pe.EffectIntensity)
	}
	if pe.Type != string(models.PanelBounceWhite) {
		t.Errorf("expected type %q, got %q", models.PanelBounceWhite, pe.Type)
	}
}

func TestBounceSilverHigherThanWhite(t *testing.T) {
	whitePanels := []models.Panel{
		{ID: "w", Name: "White", Type: models.PanelBounceWhite, Size: models.PanelSizeLarge,
			Position: models.Position3D{Distance: 1.0}, Enabled: true},
	}
	silverPanels := []models.Panel{
		{ID: "s", Name: "Silver", Type: models.PanelBounceSilver, Size: models.PanelSizeLarge,
			Position: models.Position3D{Distance: 1.0}, Enabled: true},
	}

	whiteResult := Analyze(baseSceneWithPanels(whitePanels))
	silverResult := Analyze(baseSceneWithPanels(silverPanels))

	if len(whiteResult.PanelEffects) == 0 || len(silverResult.PanelEffects) == 0 {
		t.Fatal("expected panel effects for both")
	}

	if silverResult.PanelEffects[0].EffectIntensity <= whiteResult.PanelEffects[0].EffectIntensity {
		t.Errorf("silver (%.2f) should reflect more than white (%.2f)",
			silverResult.PanelEffects[0].EffectIntensity, whiteResult.PanelEffects[0].EffectIntensity)
	}
}

func TestBounceGoldWarmsColorTemp(t *testing.T) {
	panels := []models.Panel{
		{ID: "g", Name: "Gold", Type: models.PanelBounceGold, Size: models.PanelSizeMedium,
			Position: models.Position3D{Distance: 1.0}, Enabled: true},
	}
	result := Analyze(baseSceneWithPanels(panels))

	if len(result.PanelEffects) == 0 {
		t.Fatal("expected panel effects")
	}
	pe := result.PanelEffects[0]
	if pe.ColorTempShift != 500 {
		t.Errorf("gold bounce should shift +500K, got %d", pe.ColorTempShift)
	}
	if pe.EffectIntensity <= 0 {
		t.Errorf("gold bounce should add positive intensity, got %f", pe.EffectIntensity)
	}
}

func TestNegativeFillReducesFillIntensity(t *testing.T) {
	panels := []models.Panel{
		{ID: "nf", Name: "Black V-Flat", Type: models.PanelNegativeFill, Size: models.PanelSizeLarge,
			Position: models.Position3D{X: 1, Y: 0, Z: 0.5, Distance: 1.0, Angle: -90}, Enabled: true},
	}
	result := Analyze(baseSceneWithPanels(panels))

	if len(result.PanelEffects) == 0 {
		t.Fatal("expected panel effects")
	}
	pe := result.PanelEffects[0]
	if pe.EffectIntensity >= 0 {
		t.Errorf("negative fill should subtract intensity, got %f", pe.EffectIntensity)
	}
}

func TestNegativeFillIncreasesKeyFillRatio(t *testing.T) {
	sceneNoPanels := baseSceneWithPanels(nil)
	resultNoPanels := Analyze(sceneNoPanels)

	panels := []models.Panel{
		{ID: "nf", Name: "Black V-Flat", Type: models.PanelNegativeFill, Size: models.PanelSizeLarge,
			Position: models.Position3D{X: 1, Y: 0, Z: 0.5, Distance: 1.0, Angle: -90}, Enabled: true},
	}
	resultWithPanels := Analyze(baseSceneWithPanels(panels))

	if resultWithPanels.KeyToFillRatio <= resultNoPanels.KeyToFillRatio {
		t.Errorf("negative fill should increase key:fill ratio; without=%.2f, with=%.2f",
			resultNoPanels.KeyToFillRatio, resultWithPanels.KeyToFillRatio)
	}
}

func TestFlagBlocksSpill(t *testing.T) {
	panels := []models.Panel{
		{ID: "fl", Name: "Flag", Type: models.PanelFlag, Size: models.PanelSizeSmall,
			Position: models.Position3D{X: -1, Y: 1, Z: 0, Distance: 0.8, Angle: 90}, Enabled: true},
	}
	result := Analyze(baseSceneWithPanels(panels))

	if len(result.PanelEffects) == 0 {
		t.Fatal("expected panel effects for flag")
	}
	pe := result.PanelEffects[0]
	if pe.EffectIntensity >= 0 {
		t.Errorf("flag should subtract intensity, got %f", pe.EffectIntensity)
	}
}

func TestDiffusionScrimReducesIntensity(t *testing.T) {
	panels := []models.Panel{
		{ID: "ds", Name: "Scrim Jim", Type: models.PanelDiffusion, Size: models.PanelSizeXLarge,
			Position: models.Position3D{X: 0, Y: 2, Z: 1, Distance: 1.5, Angle: 0}, Enabled: true},
	}
	result := Analyze(baseSceneWithPanels(panels))

	if len(result.PanelEffects) == 0 {
		t.Fatal("expected panel effects for diffusion scrim")
	}
	pe := result.PanelEffects[0]
	if pe.EffectIntensity >= 0 {
		t.Errorf("diffusion scrim should reduce intensity, got %f", pe.EffectIntensity)
	}
	if pe.SoftnessModifier < 0.9 {
		t.Errorf("diffusion scrim should have high softness modifier, got %f", pe.SoftnessModifier)
	}
}

func TestPanelEffectsInAnalysis(t *testing.T) {
	panels := []models.Panel{
		{ID: "wb", Name: "White Bounce", Type: models.PanelBounceWhite, Size: models.PanelSizeMedium,
			Position: models.Position3D{Distance: 1.0}, Enabled: true},
		{ID: "nf", Name: "Neg Fill", Type: models.PanelNegativeFill, Size: models.PanelSizeLarge,
			Position: models.Position3D{Distance: 1.0}, Enabled: true},
	}
	result := Analyze(baseSceneWithPanels(panels))

	if len(result.PanelEffects) != 2 {
		t.Fatalf("expected 2 panel effects, got %d", len(result.PanelEffects))
	}

	var hasBounce, hasNeg bool
	for _, pe := range result.PanelEffects {
		switch pe.Type {
		case string(models.PanelBounceWhite):
			hasBounce = true
		case string(models.PanelNegativeFill):
			hasNeg = true
		}
	}
	if !hasBounce || !hasNeg {
		t.Error("expected both bounce and negative fill effects")
	}
}

func TestPanelWarningNegFillOnKeyLightSide(t *testing.T) {
	panels := []models.Panel{
		{ID: "nf", Name: "Bad Neg Fill", Type: models.PanelNegativeFill, Size: models.PanelSizeLarge,
			Position: models.Position3D{X: -1.5, Z: 1.5, Distance: 1.0, Angle: 45}, Enabled: true},
	}
	result := Analyze(baseSceneWithPanels(panels))

	found := false
	for _, w := range result.Warnings {
		if strings.Contains(w, "negative fill on key-light side") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected warning about negative fill on key side, got: %v", result.Warnings)
	}
}

func TestDisabledPanelHasNoEffect(t *testing.T) {
	panels := []models.Panel{
		{ID: "off", Name: "Disabled", Type: models.PanelBounceWhite, Size: models.PanelSizeLarge,
			Position: models.Position3D{Distance: 1.0}, Enabled: false},
	}
	result := Analyze(baseSceneWithPanels(panels))

	if len(result.PanelEffects) != 0 {
		t.Errorf("disabled panel should not produce effects, got %d", len(result.PanelEffects))
	}
}

func TestNoPanelsProducesNilEffects(t *testing.T) {
	result := Analyze(baseSceneWithPanels(nil))
	if result.PanelEffects != nil {
		t.Errorf("scene without panels should have nil PanelEffects, got %v", result.PanelEffects)
	}
}

func TestIncidentLightGeometry(t *testing.T) {
	panel := &models.Panel{
		ID: "p1", Type: models.PanelBounceWhite, Size: models.PanelSizeLarge,
		Position: models.Position3D{X: 1, Y: 0, Z: 0.5, Distance: 1.0, Angle: -90},
		Rotation: 0, Enabled: true,
	}
	lights := []models.Light{
		{ID: "key", Role: models.RoleKey, Enabled: true, Modifier: models.ModifierSoftbox,
			Position: models.Position3D{X: -1.5, Z: 1.5, Distance: 2, Angle: 45}, Power: 80, ColorTemp: 5500},
	}
	contribs := []LightContribution{
		{LightID: "key", Intensity: 20, Direction: 45, SpillAngle: 90},
	}

	incident := computeIncidentLight(panel, lights, contribs)
	if incident <= 0 {
		t.Errorf("expected positive incident light, got %f", incident)
	}
}

func TestIncidentLightOutsideSpillCone(t *testing.T) {
	// Panel placed behind the light, well outside its forward-facing spill cone
	panel := &models.Panel{
		ID: "p1", Type: models.PanelBounceWhite, Size: models.PanelSizeLarge,
		Position: models.Position3D{X: -2, Y: 0, Z: -2, Distance: 2.83, Angle: 225},
		Enabled:  true,
	}
	lights := []models.Light{
		{ID: "key", Role: models.RoleKey, Enabled: true, Modifier: models.ModifierSnoot,
			Position: models.Position3D{X: -1.5, Z: 1.5, Distance: 2, Angle: 45}, Power: 80, ColorTemp: 5500},
	}
	contribs := []LightContribution{
		{LightID: "key", Intensity: 20, Direction: 45, SpillAngle: 15},
	}

	incident := computeIncidentLight(panel, lights, contribs)
	// Should still get at least ambient floor since there are light contributions
	if incident <= 0 {
		t.Errorf("expected positive incident (at least ambient floor), got %f", incident)
	}
	// But direct contribution should be very small (mostly ambient floor)
	if incident > 5 {
		t.Errorf("panel outside snoot cone should get minimal light, got %f", incident)
	}
}

func TestBounceCloserPanelGetsMoreLight(t *testing.T) {
	// A bounce panel closer to the key light should receive more incident
	// light and produce a stronger bounce effect
	nearPanels := []models.Panel{
		{ID: "near", Name: "Near", Type: models.PanelBounceWhite, Size: models.PanelSizeLarge,
			Position: models.Position3D{X: 0.5, Y: 0, Z: 0.5, Distance: 0.7, Angle: -45}, Enabled: true},
	}
	farPanels := []models.Panel{
		{ID: "far", Name: "Far", Type: models.PanelBounceWhite, Size: models.PanelSizeLarge,
			Position: models.Position3D{X: 2, Y: 0, Z: 2, Distance: 2.83, Angle: -45}, Enabled: true},
	}

	nearResult := Analyze(baseSceneWithPanels(nearPanels))
	farResult := Analyze(baseSceneWithPanels(farPanels))

	if len(nearResult.PanelEffects) == 0 || len(farResult.PanelEffects) == 0 {
		t.Fatal("expected panel effects for both")
	}

	if nearResult.PanelEffects[0].EffectIntensity <= farResult.PanelEffects[0].EffectIntensity {
		t.Errorf("closer panel (%.2f) should bounce more than far panel (%.2f)",
			nearResult.PanelEffects[0].EffectIntensity, farResult.PanelEffects[0].EffectIntensity)
	}
}

func TestPanelEffectsModifyCSSFilters(t *testing.T) {
	sceneNoPanels := baseSceneWithPanels(nil)
	resultNoPanels := Analyze(sceneNoPanels)

	panels := []models.Panel{
		{ID: "neg", Name: "Neg Fill", Type: models.PanelNegativeFill, Size: models.PanelSizeLarge,
			Position: models.Position3D{X: 1, Y: 0, Z: 0.5, Distance: 1.0, Angle: -90}, Enabled: true},
		{ID: "neg2", Name: "Neg Fill 2", Type: models.PanelNegativeFill, Size: models.PanelSizeLarge,
			Position: models.Position3D{X: 1.5, Y: 0, Z: 0, Distance: 1.5, Angle: -90}, Enabled: true},
	}
	resultWithPanels := Analyze(baseSceneWithPanels(panels))

	// Negative fill panels should reduce brightness since they absorb light
	if resultWithPanels.CSSFilters.Brightness >= resultNoPanels.CSSFilters.Brightness {
		t.Errorf("negative fill should reduce CSS brightness; without=%.3f, with=%.3f",
			resultNoPanels.CSSFilters.Brightness, resultWithPanels.CSSFilters.Brightness)
	}
}

func TestGoldBounceAffectsColorTemp(t *testing.T) {
	sceneNoPanels := baseSceneWithPanels(nil)
	resultNoPanels := Analyze(sceneNoPanels)

	panels := []models.Panel{
		{ID: "gold", Name: "Gold Bounce", Type: models.PanelBounceGold, Size: models.PanelSizeXLarge,
			Position: models.Position3D{X: 0.5, Y: 0, Z: 0.5, Distance: 0.7, Angle: -45}, Enabled: true},
	}
	resultGold := Analyze(baseSceneWithPanels(panels))

	// Gold bounce should shift the hue toward warm (the warmth shift should
	// change compared to no panels)
	if resultGold.CSSFilters.WarmthShift == resultNoPanels.CSSFilters.WarmthShift {
		t.Error("gold bounce panel should affect the warmth shift in CSS filters")
	}
}

func TestDiffusionPanelSoftensShadows(t *testing.T) {
	scene := &models.Scene{
		Lights: []models.Light{
			{ID: "key", Role: models.RoleKey, Enabled: true, Modifier: models.ModifierNone,
				Position: models.Position3D{X: -1.5, Z: 1.5, Distance: 2, Angle: 45}, Power: 80, ColorTemp: 5500},
		},
		Camera: models.CameraSettings{Aperture: 2.8, ISO: 100, ShutterSpeed: "1/200"},
	}
	resultHard := Analyze(scene)
	if resultHard.ShadowQuality != "hard" {
		t.Skipf("bare bulb shadow should be hard, got %s", resultHard.ShadowQuality)
	}

	sceneWithDiffusion := &models.Scene{
		Lights: scene.Lights,
		Panels: []models.Panel{
			{ID: "diff", Name: "Scrim", Type: models.PanelDiffusion, Size: models.PanelSizeXLarge,
				Position: models.Position3D{X: -1, Z: 1, Distance: 1.4, Angle: 45}, Enabled: true},
		},
		Camera: scene.Camera,
	}
	resultDiffused := Analyze(sceneWithDiffusion)

	// The CSS filters shadow gradient should have been softened
	if resultDiffused.CSSFilters.ShadowGradient == resultHard.CSSFilters.ShadowGradient {
		t.Error("diffusion panel should modify the shadow gradient compared to no panels")
	}
}

func TestPanelSizeFactor(t *testing.T) {
	cases := []struct {
		size models.PanelSize
		want float64
	}{
		{models.PanelSizeSmall, 0.3},
		{models.PanelSizeMedium, 0.55},
		{models.PanelSizeLarge, 0.85},
		{models.PanelSizeXLarge, 1.0},
		{"unknown", 0.5},
	}
	for _, tc := range cases {
		got := panelSizeFactor(tc.size)
		if math.Abs(got-tc.want) > 0.001 {
			t.Errorf("panelSizeFactor(%q) = %f, want %f", tc.size, got, tc.want)
		}
	}
}

func TestSunLightContribution(t *testing.T) {
	sun := &models.Light{
		ID: "sun", Name: "Sun", Type: models.LightTypeSun,
		Modifier: models.ModifierNone, Role: models.RoleKey,
		Position: models.Position3D{X: -1, Y: 2, Z: -2, Distance: 3.0, Angle: 210},
		Power:    100, ColorTemp: 5600, CRI: 100, Enabled: true,
	}

	contrib := computeContribution(sun)

	if contrib.SpillAngle != 180 {
		t.Errorf("sun spill should be 180°, got %f", contrib.SpillAngle)
	}
	if contrib.Softness != 0.15 {
		t.Errorf("sun softness should be 0.15, got %f", contrib.Softness)
	}
	if contrib.Intensity < 30 {
		t.Errorf("sun intensity too low: %f, expected > 30 for power=100", contrib.Intensity)
	}
	// Sun shouldn't use inverse-square; at distance=3 a 100-power strobe
	// would give ~11, but sun should give much higher
	strobe := &models.Light{
		ID: "strobe", Name: "Strobe", Type: models.LightTypeStrobe,
		Modifier: models.ModifierNone, Role: models.RoleKey,
		Position: sun.Position,
		Power:    100, Enabled: true,
	}
	strobeContrib := computeContribution(strobe)
	if contrib.Intensity <= strobeContrib.Intensity {
		t.Errorf("sun intensity (%f) should be much greater than strobe at same distance (%f)",
			contrib.Intensity, strobeContrib.Intensity)
	}
}

func TestSunPanelIncidentLight(t *testing.T) {
	sun := models.Light{
		ID: "sun", Name: "Sun", Type: models.LightTypeSun,
		Modifier: models.ModifierNone, Role: models.RoleKey,
		Position: models.Position3D{X: 0, Y: 2, Z: -2, Distance: 3.0, Angle: 180},
		Power:    100, ColorTemp: 5600, CRI: 100, Enabled: true,
	}

	panel := &models.Panel{
		ID: "refl", Name: "Reflector", Type: models.PanelBounceSilver,
		Size:     models.PanelSizeMedium,
		Position: models.Position3D{X: 1, Y: 0, Z: 1, Distance: 1.0, Angle: 45},
		Rotation: 225, Enabled: true,
	}

	contribs := []LightContribution{computeContribution(&sun)}
	incident := computeIncidentLight(panel, []models.Light{sun}, contribs)

	if incident <= 0 {
		t.Errorf("sun should produce incident light on panel, got %f", incident)
	}
}

func TestSunAnalyzeScene(t *testing.T) {
	scene := &models.Scene{
		ID: "sun_scene", Name: "Outdoor Test", Mode: models.ModeOutdoor,
		Lights: []models.Light{
			{
				ID: "sun", Name: "Sun", Type: models.LightTypeSun,
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

	analysis := Analyze(scene)

	if len(analysis.Contributions) != 1 {
		t.Fatalf("expected 1 contribution, got %d", len(analysis.Contributions))
	}
	if analysis.CSSFilters.Brightness < 0.5 {
		t.Errorf("sun scene brightness too low: %f", analysis.CSSFilters.Brightness)
	}
	if analysis.OverallEV == 0 {
		t.Error("expected non-zero EV for sun scene")
	}
}
