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

// PanelEffect represents a passive panel's computed influence on the scene.
type PanelEffect struct {
	PanelID          string  `json:"panel_id"`
	Type             string  `json:"type"`
	EffectIntensity  float64 `json:"effect_intensity"`
	SoftnessModifier float64 `json:"softness_modifier"`
	ColorTempShift   int     `json:"color_temp_shift"`
	Description      string  `json:"description"`
}

// SceneAnalysis is the computed result of the lighting engine.
type SceneAnalysis struct {
	Contributions  []LightContribution `json:"contributions"`
	PanelEffects   []PanelEffect       `json:"panel_effects,omitempty"`
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

	analysis.PanelEffects = computePanelEffects(scene, analysis.Contributions)

	for _, pe := range analysis.PanelEffects {
		switch {
		case pe.EffectIntensity > 0:
			fillIntensity += pe.EffectIntensity
		case pe.EffectIntensity < 0:
			fillIntensity = math.Max(0, fillIntensity+pe.EffectIntensity)
		}
	}

	if fillIntensity > 0 {
		analysis.KeyToFillRatio = keyIntensity / fillIntensity
	} else if keyIntensity > 0 {
		analysis.KeyToFillRatio = keyIntensity * 16
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

	var rawIntensity float64
	if light.Type == models.LightTypeSun {
		// The sun is effectively at infinite distance, so inverse-square
		// falloff doesn't apply. Power represents solar intensity scaled
		// by time of day / cloud cover (100% = direct noon sun ≈ 100k lux).
		// Height (Y) encodes sun elevation; higher Y = higher sun = more intense.
		elevation := math.Max(light.Position.Y, 0.5)
		elevFactor := math.Min(elevation/3.0, 1.0)
		rawIntensity = light.Power * 1.2 * (0.3 + 0.7*elevFactor)
	} else {
		// Inverse-square law falloff for point/area sources
		rawIntensity = light.Power / (dist * dist)
	}

	softness := modifierSoftness(light.Modifier)
	if light.Type == models.LightTypeSun {
		softness = 0.15
	}
	if light.Feathered {
		softness = math.Min(1.0, softness+0.15)
	}

	spillAngle := modifierSpill(light.Modifier, light.GridDegree)
	if light.Type == models.LightTypeSun {
		spillAngle = 180
	}

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
		var num, denom float64
		if n, _ := fmt.Sscanf(cam.ShutterSpeed, "%f/%f", &num, &denom); n == 2 && num > 0 {
			shutterFraction = denom / num
		} else if n, _ := fmt.Sscanf(cam.ShutterSpeed, "%f", &num); n == 1 && num > 0 {
			shutterFraction = 1.0 / num
		}
	}
	if shutterFraction < 1 {
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

	// Apply panel effects to total intensity and color temperature.
	// Bounce panels add reflected light; negative fill and flags subtract;
	// diffusion reduces intensity. Gold bounce shifts color temperature.
	var panelIntensityDelta float64
	var panelTempShift float64
	for _, pe := range analysis.PanelEffects {
		panelIntensityDelta += pe.EffectIntensity
		if pe.ColorTempShift != 0 && pe.EffectIntensity > 0 {
			panelTempShift += float64(pe.ColorTempShift) * pe.EffectIntensity
		}
	}

	totalIntensity = math.Max(0, totalIntensity+panelIntensityDelta)

	if totalIntensity > 0 {
		if panelTempShift != 0 {
			weightedTemp += panelTempShift
		}
		avgTemp := weightedTemp / math.Max(totalIntensity, 0.001)

		normalizedIntensity := math.Min(totalIntensity/50.0, 1.0)
		filters.Brightness = 0.3 + normalizedIntensity*1.5

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

	// Panel softness modifiers affect shadow quality.
	// Diffusion panels soften shadows; negative fill hardens the shadow side.
	panelSoftnessAdj := 0.0
	for _, pe := range analysis.PanelEffects {
		if pe.SoftnessModifier > 0.8 {
			panelSoftnessAdj += 0.1
		}
		if pe.EffectIntensity < 0 && pe.SoftnessModifier == 0 {
			panelSoftnessAdj -= 0.05
		}
	}

	shadowSide := keyDir + 180
	if shadowSide > 360 {
		shadowSide -= 360
	}
	shadowQuality := analysis.ShadowQuality
	if panelSoftnessAdj > 0.05 && shadowQuality == "hard" {
		shadowQuality = "medium"
	} else if panelSoftnessAdj < -0.03 && shadowQuality == "soft" {
		shadowQuality = "medium"
	}
	filters.ShadowGradient = buildShadowGradient(shadowSide, shadowQuality)

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

	warnings = append(warnings, generatePanelWarnings(scene)...)

	return warnings
}

func generatePanelWarnings(scene *models.Scene) []string {
	var warnings []string

	var keyAngle float64
	hasKey := false
	for _, l := range scene.Lights {
		if l.Role == models.RoleKey && l.Enabled {
			keyAngle = l.Position.Angle
			hasKey = true
			break
		}
	}

	for _, p := range scene.Panels {
		if !p.Enabled {
			continue
		}
		if p.Type == models.PanelNegativeFill && hasKey {
			angleDiff := math.Abs(p.Position.Angle - keyAngle)
			if angleDiff < 45 || angleDiff > 315 {
				warnings = append(warnings, fmt.Sprintf("Panel '%s': negative fill on key-light side will reduce subject illumination", p.Name))
			}
		}
	}

	return warnings
}

func computePanelEffects(scene *models.Scene, contribs []LightContribution) []PanelEffect {
	if len(scene.Panels) == 0 {
		return nil
	}

	// Compute how much light actually reaches each panel by checking geometric
	// alignment between every light source and the panel surface. Light that
	// falls outside a light's spill cone contributes nothing.
	effects := make([]PanelEffect, 0, len(scene.Panels))
	for i := range scene.Panels {
		p := &scene.Panels[i]
		if !p.Enabled {
			continue
		}

		incidentIntensity := computeIncidentLight(p, scene.Lights, contribs)
		effect := computeSinglePanelEffect(p, contribs, incidentIntensity)
		effects = append(effects, effect)
	}
	return effects
}

// computeIncidentLight calculates the total light energy arriving at a panel's
// surface using inverse-square falloff from each light to the panel, cosine
// attenuation based on the angle of incidence on the panel surface, and the
// light's spill cone coverage. This replaces the simplistic "total scene
// intensity" approach with proper radiometric calculation.
func computeIncidentLight(panel *models.Panel, lights []models.Light, contribs []LightContribution) float64 {
	var total float64

	panelAngleRad := panel.Position.Angle * math.Pi / 180
	panelX := math.Sin(panelAngleRad) * panel.Position.Distance
	panelZ := math.Cos(panelAngleRad) * panel.Position.Distance

	panelNormalAngle := panel.Position.Angle + 180 + panel.Rotation
	panelNX := math.Sin(panelNormalAngle * math.Pi / 180)
	panelNZ := math.Cos(panelNormalAngle * math.Pi / 180)

	for li, light := range lights {
		if !light.Enabled || li >= len(contribs) {
			continue
		}

		if light.Type == models.LightTypeSun {
			// Sun produces parallel rays from its angular direction.
			// All panels are "in the beam"; cosine attenuation applies.
			sunAngle := light.Position.Angle * math.Pi / 180
			sunDirX := -math.Sin(sunAngle)
			sunDirZ := -math.Cos(sunAngle)
			cosIncidence := math.Abs(sunDirX*panelNX + sunDirZ*panelNZ)
			total += contribs[li].Intensity * cosIncidence
			continue
		}

		lightAngleRad := light.Position.Angle * math.Pi / 180
		lightX := math.Sin(lightAngleRad) * light.Position.Distance
		lightZ := math.Cos(lightAngleRad) * light.Position.Distance

		dx := panelX - lightX
		dz := panelZ - lightZ
		distLP := math.Sqrt(dx*dx + dz*dz)
		if distLP < 0.05 {
			distLP = 0.05
		}

		dirX := dx / distLP
		dirZ := dz / distLP

		aimDist := math.Sqrt(lightX*lightX + lightZ*lightZ)
		if aimDist < 0.01 {
			aimDist = 0.01
		}
		aimX := -lightX / aimDist
		aimZ := -lightZ / aimDist

		cosAim := dirX*aimX + dirZ*aimZ
		aimAngle := math.Acos(math.Max(-1, math.Min(1, cosAim))) * 180 / math.Pi

		spillHalf := modifierSpill(light.Modifier, light.GridDegree) / 2
		if aimAngle > spillHalf {
			continue
		}

		cosIncidence := math.Abs(dirX*panelNX + dirZ*panelNZ)
		intensityAtPanel := light.Power * cosIncidence / (distLP * distLP)

		spillFraction := aimAngle / spillHalf
		edgeFalloff := 1.0 - spillFraction*spillFraction
		intensityAtPanel *= edgeFalloff

		total += intensityAtPanel
	}

	// Minimum floor so panels still register a small effect from ambient bounce
	if total < 0.5 && len(contribs) > 0 {
		var totalScene float64
		for _, c := range contribs {
			totalScene += c.Intensity
		}
		total = math.Max(total, totalScene*0.05)
	}

	return total
}

func computeSinglePanelEffect(p *models.Panel, contribs []LightContribution, incidentLight float64) PanelEffect {
	panelDist := p.Position.Distance
	if panelDist < 0.1 {
		panelDist = 0.1
	}

	sizeFactor := panelSizeFactor(p.Size)

	switch p.Type {
	case models.PanelBounceWhite:
		return computeBounceEffect(p, 0.60, 0, sizeFactor, panelDist, incidentLight)
	case models.PanelBounceSilver:
		return computeBounceEffect(p, 0.85, 0, sizeFactor, panelDist, incidentLight)
	case models.PanelBounceGold:
		return computeBounceEffect(p, 0.75, 500, sizeFactor, panelDist, incidentLight)
	case models.PanelNegativeFill:
		return computeNegativeFillEffect(p, sizeFactor, panelDist, incidentLight)
	case models.PanelFlag:
		return computeFlagEffect(p, sizeFactor, panelDist, incidentLight)
	case models.PanelDiffusion:
		return computeDiffusionEffect(p, sizeFactor, panelDist, incidentLight)
	default:
		return PanelEffect{PanelID: p.ID, Type: string(p.Type)}
	}
}

func computeBounceEffect(p *models.Panel, reflectivity float64, tempShift int, sizeFactor, panelDist, incidentLight float64) PanelEffect {
	// Bounced light = incident energy * panel reflectivity * panel area factor,
	// attenuated by inverse-square from panel to subject (panelDist).
	bounced := incidentLight * reflectivity * sizeFactor / (panelDist * panelDist)
	bounced = math.Min(bounced, incidentLight*reflectivity)

	desc := fmt.Sprintf("Reflects %.0f%% of incident light (%.1f) as soft fill → +%.1f intensity",
		reflectivity*100, incidentLight, bounced)
	if tempShift > 0 {
		desc += fmt.Sprintf(" (+%dK warmth)", tempShift)
	}

	return PanelEffect{
		PanelID:          p.ID,
		Type:             string(p.Type),
		EffectIntensity:  bounced,
		SoftnessModifier: 0.85 + sizeFactor*0.05,
		ColorTempShift:   tempShift,
		Description:      desc,
	}
}

func computeNegativeFillEffect(p *models.Panel, sizeFactor, panelDist, incidentLight float64) PanelEffect {
	// Negative fill absorbs ambient bounce light. The closer and larger the
	// panel, the more shadow-side light it removes. Uses solid-angle
	// approximation: the panel subtends more of the subject's view when
	// closer and larger.
	solidAngle := sizeFactor / (panelDist * panelDist)
	absorption := incidentLight * 0.3 * solidAngle
	absorption = math.Min(absorption, incidentLight*0.25)

	desc := fmt.Sprintf("Absorbs ambient bounce (incident %.1f), deepens shadows → −%.1f intensity",
		incidentLight, absorption)

	return PanelEffect{
		PanelID:         p.ID,
		Type:            string(p.Type),
		EffectIntensity: -absorption,
		Description:     desc,
	}
}

func computeFlagEffect(p *models.Panel, sizeFactor, panelDist, incidentLight float64) PanelEffect {
	// Flags block direct spill from lights. The blocking fraction depends
	// on how much of the light's beam the flag intercepts (solid angle).
	solidAngle := sizeFactor / (panelDist * panelDist)
	blocked := incidentLight * 0.15 * solidAngle
	blocked = math.Min(blocked, incidentLight*0.15)

	desc := fmt.Sprintf("Blocks light spill (incident %.1f) → −%.1f intensity", incidentLight, blocked)

	return PanelEffect{
		PanelID:         p.ID,
		Type:            string(p.Type),
		EffectIntensity: -blocked,
		Description:     desc,
	}
}

func computeDiffusionEffect(p *models.Panel, sizeFactor, _, incidentLight float64) PanelEffect {
	// Diffusion panels transmit ~50-65% of light while converting it from
	// specular to diffuse. Larger panels diffuse more area but also pass
	// more total light. Net result: ~1-1.5 stop reduction.
	reduction := incidentLight * 0.35 * sizeFactor
	reduction = math.Min(reduction, incidentLight*0.5)

	desc := fmt.Sprintf("Diffuses light (incident %.1f), reduces ~1-1.5 stops → −%.1f, greatly increases softness",
		incidentLight, reduction)

	return PanelEffect{
		PanelID:          p.ID,
		Type:             string(p.Type),
		EffectIntensity:  -reduction,
		SoftnessModifier: 0.95,
		Description:      desc,
	}
}

func panelSizeFactor(size models.PanelSize) float64 {
	switch size {
	case models.PanelSizeSmall:
		return 0.3
	case models.PanelSizeMedium:
		return 0.55
	case models.PanelSizeLarge:
		return 0.85
	case models.PanelSizeXLarge:
		return 1.0
	default:
		return 0.5
	}
}

func formatFloat(f float64) string {
	rounded := math.Round(f*100) / 100
	if rounded == math.Trunc(rounded) {
		return fmt.Sprintf("%d", int(rounded))
	}
	return fmt.Sprintf("%.2f", rounded)
}
