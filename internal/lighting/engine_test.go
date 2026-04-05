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
