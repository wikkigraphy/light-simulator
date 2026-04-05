package lighting

import (
	"fmt"
	"math"

	"github.com/srivickynesh/light-simulator/internal/models"
)

// LightContribution represents a single light's computed effect on the subject.
type LightContribution struct {
	LightID    string  `json:"light_id"`
	Role       string  `json:"role"`
	Intensity  float64 `json:"intensity"`   // effective intensity after falloff
	Direction  float64 `json:"direction"`   // angle of incidence on subject
	Softness   float64 `json:"softness"`    // 0 = hard, 1 = very soft
	SpillAngle float64 `json:"spill_angle"` // beam spread in degrees
	ColorTemp  int     `json:"color_temp"`
}

// SceneAnalysis is the computed result of the lighting engine.
type SceneAnalysis struct {
	Contributions  []LightContribution `json:"contributions"`
	KeyToFillRatio float64             `json:"key_to_fill_ratio"`
	OverallEV      float64             `json:"overall_ev"`
	ShadowQuality  string              `json:"shadow_quality"` // "hard", "medium", "soft"
	CatchlightType string              `json:"catchlight_type"`
	Warnings       []string            `json:"warnings"`
	CSSFilters     CSSLightingFilters  `json:"css_filters"`
}

// CSSLightingFilters are the computed CSS filter values for the preview image.
type CSSLightingFilters struct {
	Brightness     float64 `json:"brightness"`
	Contrast       float64 `json:"contrast"`
	Saturate       float64 `json:"saturate"`
	HueRotate      float64 `json:"hue_rotate"`
	ShadowGradient string  `json:"shadow_gradient"` // CSS gradient for shadow overlay
	HighlightPos   string  `json:"highlight_pos"`   // CSS radial-gradient position
	WarmthShift    float64 `json:"warmth_shift"`    // negative=cool, positive=warm
}

// Analyze computes the lighting effect for a complete scene.
func Analyze(scene *models.Scene) *SceneAnalysis {
	analysis := &SceneAnalysis{
		Contributions: make([]LightContribution, 0, len(scene.Lights)),
	}

	var keyIntensity, fillIntensity float64

	for i := range scene.Lights {
		light := &scene.Lights[i]
		if !light.Enabled {
			continue
		}

		contrib := computeContribution(light)
		analysis.Contributions = append(analysis.Contributions, contrib)

		switch light.Role {
		case models.RoleKey:
			keyIntensity = contrib.Intensity
		case models.RoleFill:
			fillIntensity = contrib.Intensity
		}
	}

	if fillIntensity > 0 {
		analysis.KeyToFillRatio = keyIntensity / fillIntensity
	}

	analysis.OverallEV = computeExposureValue(scene)
	analysis.ShadowQuality = classifyShadows(analysis)
	analysis.CatchlightType = determineCatchlight(scene)
	analysis.CSSFilters = computeCSSFilters(scene, analysis)
	analysis.Warnings = generateWarnings(scene, analysis)

	return analysis
}

func computeContribution(light *models.Light) LightContribution {
	dist := light.Position.Distance
	if dist < 0.1 {
		dist = 0.1
	}

	// Inverse-square law falloff
	rawIntensity := light.Power / (dist * dist)

	softness := modifierSoftness(light.Modifier)
	if light.Feathered {
		softness = math.Min(1.0, softness+0.15)
	}

	spillAngle := modifierSpill(light.Modifier, light.GridDegree)

	direction := math.Atan2(light.Position.X, light.Position.Z) * (180.0 / math.Pi)

	return LightContribution{
		LightID:    light.ID,
		Role:       string(light.Role),
		Intensity:  rawIntensity,
		Direction:  direction,
		Softness:   softness,
		SpillAngle: spillAngle,
		ColorTemp:  light.ColorTemp,
	}
}

func modifierSoftness(mod models.ModifierType) float64 {
	softness := map[models.ModifierType]float64{
		models.ModifierNone:       0.1,
		models.ModifierHoneycomb:  0.15,
		models.ModifierSnoot:      0.05,
		models.ModifierBarnDoors:  0.1,
		models.ModifierReflector:  0.3,
		models.ModifierBeautyDish: 0.5,
		models.ModifierUmbrella:   0.65,
		models.ModifierSoftbox:    0.75,
		models.ModifierStripbox:   0.7,
		models.ModifierOctabox:    0.85,
		models.ModifierDiffusion:  0.9,
		models.ModifierParabolic:  0.6,
	}
	if v, ok := softness[mod]; ok {
		return v
	}
	return 0.5
}

func modifierSpill(mod models.ModifierType, gridDeg int) float64 {
	if mod == models.ModifierHoneycomb && gridDeg > 0 {
		return float64(gridDeg)
	}
	spill := map[models.ModifierType]float64{
		models.ModifierNone:       180,
		models.ModifierSnoot:      15,
		models.ModifierBarnDoors:  40,
		models.ModifierHoneycomb:  30,
		models.ModifierReflector:  90,
		models.ModifierBeautyDish: 70,
		models.ModifierUmbrella:   120,
		models.ModifierSoftbox:    90,
		models.ModifierStripbox:   60,
		models.ModifierOctabox:    100,
		models.ModifierDiffusion:  140,
		models.ModifierParabolic:  50,
	}
	if v, ok := spill[mod]; ok {
		return v
	}
	return 90
}

func computeExposureValue(scene *models.Scene) float64 {
	cam := scene.Camera
	fStop := cam.Aperture
	if fStop < 1 {
		fStop = 2.8
	}

	var shutterFraction float64 = 200
	if cam.ShutterSpeed != "" {
		// Simplified: assume format "1/N"
		shutterFraction = 200
	}

	iso := float64(cam.ISO)
	if iso < 50 {
		iso = 100
	}

	ev := math.Log2((fStop*fStop*shutterFraction)/iso) + scene.Ambient*2
	return math.Round(ev*10) / 10
}

func classifyShadows(analysis *SceneAnalysis) string {
	if len(analysis.Contributions) == 0 {
		return "soft"
	}

	var avgSoftness float64
	for _, c := range analysis.Contributions {
		avgSoftness += c.Softness
	}
	avgSoftness /= float64(len(analysis.Contributions))

	switch {
	case avgSoftness > 0.6:
		return "soft"
	case avgSoftness > 0.3:
		return "medium"
	default:
		return "hard"
	}
}

func determineCatchlight(scene *models.Scene) string {
	for _, l := range scene.Lights {
		if l.Role == models.RoleKey && l.Enabled {
			switch l.Modifier {
			case models.ModifierOctabox:
				return "octagonal"
			case models.ModifierSoftbox, models.ModifierStripbox:
				return "rectangular"
			case models.ModifierBeautyDish:
				return "circular_ring"
			case models.ModifierUmbrella:
				return "circular"
			case models.ModifierParabolic:
				return "parabolic"
			default:
				return "point"
			}
		}
	}
	return "none"
}

func computeCSSFilters(scene *models.Scene, analysis *SceneAnalysis) CSSLightingFilters {
	filters := CSSLightingFilters{
		Brightness: 1.0,
		Contrast:   1.0,
		Saturate:   1.0,
	}

	var totalIntensity float64
	var weightedTemp float64
	var keyDir float64

	for _, c := range analysis.Contributions {
		totalIntensity += c.Intensity
		weightedTemp += float64(c.ColorTemp) * c.Intensity
		if c.Role == "key" {
			keyDir = c.Direction
		}
	}

	if totalIntensity > 0 {
		avgTemp := weightedTemp / totalIntensity

		// Brightness: scale to reasonable CSS range [0.3 - 1.8]
		normalizedIntensity := math.Min(totalIntensity/50.0, 1.0)
		filters.Brightness = 0.3 + normalizedIntensity*1.5

		// Color temperature: 5500K is neutral, below=warm, above=cool
		filters.WarmthShift = (avgTemp - 5500) / 3000.0
		if filters.WarmthShift > 0 {
			filters.HueRotate = -filters.WarmthShift * 15
		} else {
			filters.HueRotate = -filters.WarmthShift * 20
		}
	}

	// Contrast from key:fill ratio
	if analysis.KeyToFillRatio > 1 {
		filters.Contrast = 1.0 + math.Min(analysis.KeyToFillRatio/8.0, 0.5)
	}

	// Shadow gradient position based on key light direction
	shadowSide := keyDir + 180
	if shadowSide > 360 {
		shadowSide -= 360
	}
	filters.ShadowGradient = buildShadowGradient(shadowSide, analysis.ShadowQuality)

	// Highlight position from key light
	hlX := 50 + math.Sin(keyDir*math.Pi/180)*30
	hlY := 50 - math.Cos(keyDir*math.Pi/180)*30
	filters.HighlightPos = buildHighlightPos(hlX, hlY)

	return filters
}

func buildShadowGradient(angle float64, quality string) string {
	var opacity float64
	switch quality {
	case "hard":
		opacity = 0.6
	case "medium":
		opacity = 0.35
	default:
		opacity = 0.15
	}
	return formatShadowGradient(angle, opacity)
}

func formatShadowGradient(angle, opacity float64) string {
	return "linear-gradient(" +
		formatFloat(angle) + "deg, " +
		"rgba(0,0,0," + formatFloat(opacity) + ") 0%, " +
		"rgba(0,0,0,0) 60%)"
}

func buildHighlightPos(x, y float64) string {
	return "radial-gradient(circle at " +
		formatFloat(x) + "% " + formatFloat(y) + "%, " +
		"rgba(255,255,255,0.15) 0%, rgba(255,255,255,0) 50%)"
}

func generateWarnings(scene *models.Scene, analysis *SceneAnalysis) []string {
	var warnings []string

	hasKey := false
	for _, l := range scene.Lights {
		if l.Role == models.RoleKey && l.Enabled {
			hasKey = true
			break
		}
	}
	if !hasKey && len(scene.Lights) > 0 {
		warnings = append(warnings, "No key light defined — consider assigning a key role for primary illumination")
	}

	if analysis.KeyToFillRatio > 8 {
		warnings = append(warnings, "Key-to-fill ratio is very high (>8:1) — shadows may lose all detail")
	}

	for _, c := range analysis.Contributions {
		if c.Intensity > 200 {
			warnings = append(warnings, fmt.Sprintf("Light '%s' has very high intensity — risk of blown highlights", c.LightID))
		}
	}

	// Color temperature mismatch check
	var temps []int
	for _, l := range scene.Lights {
		if l.Enabled {
			temps = append(temps, l.ColorTemp)
		}
	}
	if len(temps) > 1 {
		minT, maxT := temps[0], temps[0]
		for _, t := range temps[1:] {
			if t < minT {
				minT = t
			}
			if t > maxT {
				maxT = t
			}
		}
		if maxT-minT > 1500 {
			warnings = append(warnings, fmt.Sprintf("Color temperature mismatch: %dK spread — may cause unwanted color casts unless intentional", maxT-minT))
		}
	}

	return warnings
}

func formatFloat(f float64) string {
	rounded := math.Round(f*100) / 100
	if rounded == math.Trunc(rounded) {
		return fmt.Sprintf("%d", int(rounded))
	}
	return fmt.Sprintf("%.2f", rounded)
}
