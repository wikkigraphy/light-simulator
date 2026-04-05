package cheatsheet

import "github.com/srivickynesh/light-simulator/internal/models"

// AllPresets returns the complete set of professional lighting presets.
func AllPresets() []models.Preset {
	return []models.Preset{
		rembrandt(),
		butterfly(),
		splitLight(),
		loopLighting(),
		clamshell(),
		broadLight(),
		shortLight(),
		highKeyPortrait(),
		lowKeyPortrait(),
		beautyRingLight(),
		cinematicNoir(),
		crossLighting(),
		productTopDown(),
		productHero(),
		productWhiteBG(),
		productGlassware(),
		fashionEditorial(),
		fashionCatalog(),
		foodMoody(),
		foodBright(),
		headshotCorporate(),
		rimLightDramatic(),
		groupPhoto(),
		sportAction(),
		outdoorGoldenHour(),
		outdoorHarshMidDay(),
		outdoorOpenShade(),
	}
}

// PresetsByCategory groups presets by shooting category.
func PresetsByCategory() map[string][]models.Preset {
	result := make(map[string][]models.Preset)
	for _, p := range AllPresets() {
		result[p.Category] = append(result[p.Category], p)
	}
	return result
}

func rembrandt() models.Preset {
	return models.Preset{
		ID:       "rembrandt",
		Name:     "Rembrandt Lighting",
		Category: "portrait",
		Description: "Classic portrait lighting creating a triangle of light on the shadow-side cheek. " +
			"Key light at 45° angle, slightly above eye level. Named after the painter's characteristic style. " +
			"A black V-flat on the shadow side deepens the Rembrandt triangle, while a white bounce below camera lifts chin shadows.",
		Equipment: []models.EquipmentItem{
			{Role: "Key", Device: "Studio Strobe 300Ws", Modifier: "24×36″ Softbox", Power: "75%", Placement: "45° left, slightly above eye level, 2m", Recommended: "Godox AD300Pro / Profoto B10"},
			{Role: "Fill", Device: "5-in-1 Reflector (Silver)", Modifier: "42″ Reflector Disc", Power: "Passive bounce", Placement: "30° right, eye level, 1.5m", Recommended: "Neewer 42″ 5-in-1 / Profoto Collapsible Reflector"},
			{Role: "Accessory", Device: "V-Flat (Black side)", Modifier: "Negative fill panel", Power: "N/A", Placement: "Shadow side, 1m from subject, deepens triangle", Recommended: "V-Flat World Duo Board / black foamcore 4×8′"},
			{Role: "Accessory", Device: "Light Stand (Heavy Duty)", Modifier: "C-Stand with grip head", Power: "N/A", Placement: "Supporting key light", Recommended: "Avenger C-Stand C40 / Manfrotto 1004BAC"},
			{Role: "Accessory", Device: "Sandbag (15 lb)", Modifier: "Counterweight", Power: "N/A", Placement: "Base of C-stand", Recommended: "Impact Saddle Sandbag / Neewer Sandbag"},
			{Role: "Accessory", Device: "Radio Trigger", Modifier: "Wireless flash trigger", Power: "N/A", Placement: "Camera hotshoe + receiver on strobe", Recommended: "Godox X2T / Profoto Air Remote"},
		},
		Scene: models.Scene{
			Mode: models.ModePortrait,
			Lights: []models.Light{
				{
					ID: "key", Name: "Key Light", Type: models.LightTypeStrobe,
					Modifier: models.ModifierSoftbox, Role: models.RoleKey,
					Position: models.Position3D{X: -1.41, Y: 0.5, Z: -1.41, Distance: 2.0, Angle: -135},
					Power:    75, ColorTemp: 5500, CRI: 95, Enabled: true,
				},
				{
					ID: "fill", Name: "Fill Reflector", Type: models.LightTypeContinuous,
					Modifier: models.ModifierReflector, Role: models.RoleFill,
					Position: models.Position3D{X: 0.75, Y: 0.0, Z: -1.30, Distance: 1.5, Angle: 210},
					Power:    30, ColorTemp: 5500, CRI: 90, Enabled: true,
				},
			},
			Panels: []models.Panel{
				{ID: "neg_vflat", Name: "Black V-Flat (Shadow Side)", Type: models.PanelNegativeFill, Size: models.PanelSizeLarge,
					Position: models.Position3D{X: 1.0, Y: 0, Z: 0, Distance: 1.0, Angle: 90}, Enabled: true},
				{ID: "chin_bounce", Name: "White Bounce (Below Camera)", Type: models.PanelBounceWhite, Size: models.PanelSizeMedium,
					Position: models.Position3D{X: 0, Y: -0.5, Z: -0.8, Distance: 0.8, Angle: 180}, Enabled: true},
			},
			Camera:   defaultPortraitCamera(),
			Backdrop: "#1a1a1a",
			Ambient:  0.1,
		},
	}
}

func butterfly() models.Preset {
	return models.Preset{
		ID:       "butterfly",
		Name:     "Butterfly / Paramount Lighting",
		Category: "portrait",
		Description: "Key light directly above and in front of the subject, creating a butterfly-shaped shadow under the nose. " +
			"Flattering for most face shapes. A white reflector below the chin bounces the overhead key back into under-nose shadows, " +
			"completing the butterfly pattern.",
		Equipment: []models.EquipmentItem{
			{Role: "Key", Device: "Studio Strobe 500Ws", Modifier: "22″ Beauty Dish (white)", Power: "80%", Placement: "Directly above, centered, 2m", Recommended: "Godox AD600Pro + beauty dish / Profoto B10 Plus"},
			{Role: "Fill", Device: "5-in-1 Reflector (White)", Modifier: "43″ Reflector Disc", Power: "Passive bounce", Placement: "Below chin, on subject's lap or held, 1m", Recommended: "Neewer 43″ Reflector / Lastolite Triflip"},
			{Role: "Accessory", Device: "Boom Arm + C-Stand", Modifier: "Overhead mount for beauty dish", Power: "N/A", Placement: "Extends directly over subject", Recommended: "Avenger D600 Boom / Manfrotto 025BS"},
			{Role: "Accessory", Device: "Sandbag (25 lb)", Modifier: "Counterweight for boom", Power: "N/A", Placement: "Opposite end of boom arm", Recommended: "Impact Saddle Sandbag / Neewer Heavy-Duty"},
			{Role: "Accessory", Device: "Radio Trigger", Modifier: "Wireless flash trigger", Power: "N/A", Placement: "Camera hotshoe", Recommended: "Godox X2T / Profoto Air Remote"},
		},
		Scene: models.Scene{
			Mode: models.ModePortrait,
			Lights: []models.Light{
				{
					ID: "key", Name: "Key Light", Type: models.LightTypeStrobe,
					Modifier: models.ModifierBeautyDish, Role: models.RoleKey,
					Position: models.Position3D{X: 0, Y: 1.2, Z: -2.0, Distance: 2.0, Angle: 180},
					Power:    80, ColorTemp: 5600, CRI: 96, Enabled: true,
				},
				{
					ID: "fill", Name: "Fill Reflector", Type: models.LightTypeContinuous,
					Modifier: models.ModifierReflector, Role: models.RoleFill,
					Position: models.Position3D{X: 0, Y: -0.8, Z: -1.0, Distance: 1.0, Angle: 180},
					Power:    40, ColorTemp: 5600, CRI: 90, Enabled: true,
				},
			},
			Panels: []models.Panel{
				{ID: "chin_reflector", Name: "White Reflector (Below Chin)", Type: models.PanelBounceWhite, Size: models.PanelSizeMedium,
					Position: models.Position3D{X: 0, Y: -0.8, Z: -0.8, Distance: 0.8, Angle: 180}, Enabled: true},
			},
			Camera:   defaultPortraitCamera(),
			Backdrop: "#2a2a2a",
			Ambient:  0.05,
		},
	}
}

func splitLight() models.Preset {
	return models.Preset{
		ID:       "split",
		Name:     "Split Lighting",
		Category: "portrait",
		Description: "Key light at 90° to the subject, illuminating exactly half the face. " +
			"Creates dramatic, moody portraits with strong contrast. A large black V-flat on the unlit side prevents ambient bounce " +
			"from filling shadows, maintaining the hard 50/50 split. A flag above the key blocks spill onto the background.",
		Equipment: []models.EquipmentItem{
			{Role: "Key", Device: "Studio Strobe 300Ws", Modifier: "24×36″ Softbox", Power: "70%", Placement: "90° left, eye level, 2m", Recommended: "Godox AD300Pro / Elinchrom D-Lite RX4"},
			{Role: "Accessory", Device: "V-Flat (Black side)", Modifier: "Negative fill panel", Power: "N/A", Placement: "Fill side (right), 0.8m from subject, deepens contrast", Recommended: "V-Flat World Duo Board / black foamcore 4×8′"},
			{Role: "Accessory", Device: "Flag / Gobo (24×36″)", Modifier: "Spill blocker", Power: "N/A", Placement: "Above key to prevent spill onto background", Recommended: "Matthews Solid Floppy / Avenger Flag"},
			{Role: "Accessory", Device: "Black Seamless Paper (9′)", Modifier: "Background", Power: "N/A", Placement: "2m behind subject", Recommended: "Savage Seamless #20 Black"},
		},
		Scene: models.Scene{
			Mode: models.ModePortrait,
			Lights: []models.Light{
				{
					ID: "key", Name: "Key Light", Type: models.LightTypeStrobe,
					Modifier: models.ModifierSoftbox, Role: models.RoleKey,
					Position: models.Position3D{X: -2.0, Y: 0.3, Z: 0, Distance: 2.0, Angle: -90},
					Power:    70, ColorTemp: 5500, CRI: 95, Enabled: true,
				},
			},
			Panels: []models.Panel{
				{ID: "neg_vflat", Name: "Black V-Flat (Fill Side)", Type: models.PanelNegativeFill, Size: models.PanelSizeLarge,
					Position: models.Position3D{X: 0.8, Y: 0, Z: 0, Distance: 0.8, Angle: 90}, Enabled: true},
				{ID: "flag_top", Name: "Black Flag (Above Key)", Type: models.PanelFlag, Size: models.PanelSizeMedium,
					Position: models.Position3D{X: -1.2, Y: 1.5, Z: 0, Distance: 1.2, Angle: -90}, Enabled: true},
			},
			Camera:   defaultPortraitCamera(),
			Backdrop: "#0d0d0d",
			Ambient:  0.02,
		},
	}
}

func loopLighting() models.Preset {
	return models.Preset{
		ID:       "loop",
		Name:     "Loop Lighting",
		Category: "portrait",
		Description: "Key light 30-45° to one side and slightly above, creating a small shadow loop from the nose. " +
			"Versatile and flattering for most subjects. A white bounce card below camera fills chin shadows without " +
			"affecting the characteristic nose loop shadow.",
		Equipment: []models.EquipmentItem{
			{Role: "Key", Device: "Studio Strobe 300Ws", Modifier: "47″ Octabox", Power: "70%", Placement: "35° left, above eye level, 2.2m", Recommended: "Godox AD300Pro + octabox / Profoto OCF Octa"},
			{Role: "Fill", Device: "Studio Strobe 150Ws", Modifier: "Shoot-through Umbrella 43″", Power: "25%", Placement: "25° right, eye level, 2.5m", Recommended: "Godox MS200 + umbrella / Westcott 43″"},
			{Role: "Accessory", Device: "White Bounce Board (V-Flat)", Modifier: "Passive fill boost", Power: "N/A", Placement: "Below camera, angled up to fill chin shadows", Recommended: "V-Flat World Duo Board (white side) / white foamcore"},
			{Role: "Accessory", Device: "Hair Light (optional)", Modifier: "7″ Reflector + 20° Grid", Power: "30%", Placement: "Behind-above subject for separation", Recommended: "Godox MS200 + grid / speedlight with grid"},
		},
		Scene: models.Scene{
			Mode: models.ModePortrait,
			Lights: []models.Light{
				{
					ID: "key", Name: "Key Light", Type: models.LightTypeStrobe,
					Modifier: models.ModifierOctabox, Role: models.RoleKey,
					Position: models.Position3D{X: -1.26, Y: 0.6, Z: -1.80, Distance: 2.2, Angle: -145},
					Power:    70, ColorTemp: 5500, CRI: 97, Enabled: true,
				},
				{
					ID: "fill", Name: "Fill Light", Type: models.LightTypeStrobe,
					Modifier: models.ModifierUmbrella, Role: models.RoleFill,
					Position: models.Position3D{X: 1.06, Y: 0, Z: -2.27, Distance: 2.5, Angle: 205},
					Power:    25, ColorTemp: 5500, CRI: 95, Enabled: true,
				},
			},
			Panels: []models.Panel{
				{ID: "chin_bounce", Name: "White Bounce (Below Camera)", Type: models.PanelBounceWhite, Size: models.PanelSizeMedium,
					Position: models.Position3D{X: 0, Y: -0.5, Z: -0.8, Distance: 0.8, Angle: 180}, Enabled: true},
			},
			Camera:   defaultPortraitCamera(),
			Backdrop: "#333333",
			Ambient:  0.1,
		},
	}
}

func clamshell() models.Preset {
	return models.Preset{
		ID:       "clamshell",
		Name:     "Clamshell Lighting",
		Category: "portrait",
		Description: "Two lights sandwiching the subject vertically: one above and one below. " +
			"Creates even, beauty-style lighting ideal for beauty and headshot work. " +
			"A small white bounce on the subject's lap adds a third layer of chin fill between the two lights.",
		Equipment: []models.EquipmentItem{
			{Role: "Key (Above)", Device: "Studio Strobe 500Ws", Modifier: "47″ Octabox", Power: "75%", Placement: "Centered above, 1.8m", Recommended: "Godox AD600Pro + octa / Profoto D2 500"},
			{Role: "Fill (Below)", Device: "Studio Strobe 300Ws", Modifier: "24×36″ Softbox", Power: "45%", Placement: "Centered below chin, 1.5m", Recommended: "Godox AD300Pro + softbox / Elinchrom ELC 125"},
			{Role: "Accessory", Device: "White Bounce Board", Modifier: "Chin fill reflector", Power: "N/A", Placement: "On subject's lap, angled up toward chin", Recommended: "Lastolite Triflip / white foamcore 20×30″"},
			{Role: "Accessory", Device: "Color Checker Passport", Modifier: "Color reference card", Power: "N/A", Placement: "Subject holds for first frame, removed after", Recommended: "X-Rite ColorChecker Passport / Datacolor SpyderCheckr"},
			{Role: "Accessory", Device: "Boom Arm + C-Stand", Modifier: "Overhead mount for octabox", Power: "N/A", Placement: "Extends above subject, key hangs from boom", Recommended: "Avenger D600 Boom / Manfrotto 025BS"},
		},
		Scene: models.Scene{
			Mode: models.ModeHeadshot,
			Lights: []models.Light{
				{
					ID: "key", Name: "Key (Above)", Type: models.LightTypeStrobe,
					Modifier: models.ModifierOctabox, Role: models.RoleKey,
					Position: models.Position3D{X: 0, Y: 1.0, Z: -1.8, Distance: 1.8, Angle: 180},
					Power:    75, ColorTemp: 5500, CRI: 97, Enabled: true,
				},
				{
					ID: "fill", Name: "Fill (Below)", Type: models.LightTypeStrobe,
					Modifier: models.ModifierSoftbox, Role: models.RoleFill,
					Position: models.Position3D{X: 0, Y: -0.5, Z: -1.5, Distance: 1.5, Angle: 180},
					Power:    45, ColorTemp: 5500, CRI: 97, Enabled: true,
				},
			},
			Panels: []models.Panel{
				{ID: "lap_bounce", Name: "White Bounce (Subject's Lap)", Type: models.PanelBounceWhite, Size: models.PanelSizeSmall,
					Position: models.Position3D{X: 0, Y: -0.6, Z: -0.3, Distance: 0.3, Angle: 180}, Enabled: true},
			},
			Camera:   defaultPortraitCamera(),
			Backdrop: "#ffffff",
			Ambient:  0.15,
		},
	}
}

func broadLight() models.Preset {
	return models.Preset{
		ID:       "broad",
		Name:     "Broad Lighting",
		Category: "portrait",
		Description: "Subject turned slightly away; key light illuminates the side of the face closest to camera. " +
			"Makes face appear wider. Good for thin faces. A white V-flat on the shadow side softens the transition " +
			"edge without eliminating the face-widening shadow pattern.",
		Equipment: []models.EquipmentItem{
			{Role: "Key", Device: "Studio Strobe 300Ws", Modifier: "36″ Softbox", Power: "70%", Placement: "45° camera-side, above eye level, 2m", Recommended: "Godox AD300Pro / Profoto B10"},
			{Role: "Fill", Device: "Studio Strobe 150Ws", Modifier: "Shoot-through Umbrella 43″", Power: "20%", Placement: "25° opposite side, eye level, 2.5m", Recommended: "Godox MS200 / Westcott Umbrella"},
			{Role: "Accessory", Device: "White Bounce Card", Modifier: "Passive fill on shadow side", Power: "N/A", Placement: "Shadow side, 1m from subject, softens transition", Recommended: "V-Flat World (white side) / white foamcore 4×8′"},
		},
		Scene: models.Scene{
			Mode: models.ModePortrait,
			Lights: []models.Light{
				{
					ID: "key", Name: "Key Light", Type: models.LightTypeStrobe,
					Modifier: models.ModifierSoftbox, Role: models.RoleKey,
					Position: models.Position3D{X: 1.41, Y: 0.5, Z: -1.41, Distance: 2.0, Angle: 135},
					Power:    70, ColorTemp: 5500, CRI: 95, Enabled: true,
				},
				{
					ID: "fill", Name: "Fill Light", Type: models.LightTypeStrobe,
					Modifier: models.ModifierUmbrella, Role: models.RoleFill,
					Position: models.Position3D{X: -1.06, Y: 0, Z: -2.27, Distance: 2.5, Angle: -155},
					Power:    20, ColorTemp: 5500, CRI: 95, Enabled: true,
				},
			},
			Panels: []models.Panel{
				{ID: "shadow_bounce", Name: "White Bounce (Shadow Side)", Type: models.PanelBounceWhite, Size: models.PanelSizeLarge,
					Position: models.Position3D{X: -1.0, Y: 0, Z: 0, Distance: 1.0, Angle: -90}, Enabled: true},
			},
			Camera:   defaultPortraitCamera(),
			Backdrop: "#2b2b2b",
			Ambient:  0.1,
		},
	}
}

func shortLight() models.Preset {
	return models.Preset{
		ID:       "short",
		Name:     "Short Lighting",
		Category: "portrait",
		Description: "Subject turned slightly away; key light illuminates the far side of the face (away from camera). " +
			"Creates more shadow, slims the face. Dramatic look. A black V-flat on the camera side prevents ambient bounce " +
			"from reducing contrast, while an optional silver reflector provides subtle fill if shadows become too deep.",
		Equipment: []models.EquipmentItem{
			{Role: "Key", Device: "Studio Strobe 300Ws", Modifier: "24×36″ Softbox", Power: "70%", Placement: "45° far side, above eye level, 2m", Recommended: "Godox AD300Pro / Elinchrom D-Lite RX4"},
			{Role: "Accessory", Device: "V-Flat (Black side)", Modifier: "Negative fill panel", Power: "N/A", Placement: "Camera side, 1m from subject, prevents ambient bounce", Recommended: "V-Flat World Duo Board / black foamcore 4×8′"},
			{Role: "Accessory", Device: "5-in-1 Reflector (Silver)", Modifier: "Optional subtle fill kick", Power: "N/A", Placement: "Far below key, 2m, feathered — only if shadows too deep", Recommended: "Neewer 42″ 5-in-1 / Lastolite Triflip"},
		},
		Scene: models.Scene{
			Mode: models.ModePortrait,
			Lights: []models.Light{
				{
					ID: "key", Name: "Key Light", Type: models.LightTypeStrobe,
					Modifier: models.ModifierSoftbox, Role: models.RoleKey,
					Position: models.Position3D{X: -1.41, Y: 0.5, Z: -1.41, Distance: 2.0, Angle: -135},
					Power:    70, ColorTemp: 5500, CRI: 95, Enabled: true,
				},
			},
			Panels: []models.Panel{
				{ID: "neg_camera_side", Name: "Black V-Flat (Camera Side)", Type: models.PanelNegativeFill, Size: models.PanelSizeLarge,
					Position: models.Position3D{X: 1.0, Y: 0, Z: -0.3, Distance: 1.0, Angle: 107}, Enabled: true},
				{ID: "silver_fill", Name: "Silver Reflector (Optional)", Type: models.PanelBounceSilver, Size: models.PanelSizeMedium,
					Position: models.Position3D{X: 0.60, Y: -0.3, Z: -1.08, Distance: 1.2, Angle: 150}, Enabled: true},
			},
			Camera:   defaultPortraitCamera(),
			Backdrop: "#1a1a1a",
			Ambient:  0.05,
		},
	}
}

func highKeyPortrait() models.Preset {
	return models.Preset{
		ID:       "high_key",
		Name:     "High-Key Portrait",
		Category: "portrait",
		Description: "Bright, low-contrast setup with white background lights. " +
			"Two background lights + large key source. Clean, airy feel. White V-flats on both sides " +
			"of the subject bounce the key light for wraparound fill, ensuring minimal shadow.",
		Equipment: []models.EquipmentItem{
			{Role: "Key", Device: "Studio Strobe 500Ws", Modifier: "60″ Octabox", Power: "60%", Placement: "15° left, above eye level, 2m", Recommended: "Profoto D2 500 + Giant Octa / Godox AD600Pro"},
			{Role: "Fill", Device: "Studio Strobe 300Ws", Modifier: "Shoot-through Umbrella 60″", Power: "45%", Placement: "20° right, eye level, 2.5m", Recommended: "Godox AD300Pro / Westcott 60″"},
			{Role: "Background L", Device: "Studio Strobe 300Ws", Modifier: "Bare bulb (no modifier)", Power: "90%", Placement: "Behind subject left, aimed at backdrop, 1m", Recommended: "Godox MS300 / Elinchrom D-Lite RX4"},
			{Role: "Background R", Device: "Studio Strobe 300Ws", Modifier: "Bare bulb (no modifier)", Power: "90%", Placement: "Behind subject right, aimed at backdrop, 1m", Recommended: "Godox MS300 / Elinchrom D-Lite RX4"},
			{Role: "Accessory", Device: "White Vinyl Floor / Seamless", Modifier: "White floor covering", Power: "N/A", Placement: "Draped from backdrop down onto floor, 3m forward", Recommended: "Savage Floor Drop / white vinyl roll"},
			{Role: "Accessory", Device: "Light Meter", Modifier: "Incident meter", Power: "N/A", Placement: "Used at subject position, aimed at each light", Recommended: "Sekonic L-308X / Sekonic L-858D"},
			{Role: "Accessory", Device: "Gaffer Tape", Modifier: "Mark positions", Power: "N/A", Placement: "Floor marks for subject and light stands", Recommended: "Pro Gaff / Shurtape P-665"},
		},
		Scene: models.Scene{
			Mode: models.ModePortrait,
			Lights: []models.Light{
				{
					ID: "key", Name: "Key Light", Type: models.LightTypeStrobe,
					Modifier: models.ModifierOctabox, Role: models.RoleKey,
					Position: models.Position3D{X: -0.52, Y: 0.8, Z: -1.93, Distance: 2.0, Angle: -165},
					Power:    60, ColorTemp: 5600, CRI: 97, Enabled: true,
				},
				{
					ID: "fill", Name: "Fill Light", Type: models.LightTypeStrobe,
					Modifier: models.ModifierUmbrella, Role: models.RoleFill,
					Position: models.Position3D{X: 0.86, Y: 0, Z: -2.35, Distance: 2.5, Angle: 160},
					Power:    45, ColorTemp: 5600, CRI: 95, Enabled: true,
				},
				{
					ID: "bg1", Name: "Background Left", Type: models.LightTypeStrobe,
					Modifier: models.ModifierNone, Role: models.RoleBackground,
					Position: models.Position3D{X: -0.7, Y: 0, Z: 0.7, Distance: 1.0, Angle: -45},
					Power:    90, ColorTemp: 5600, CRI: 90, Enabled: true,
				},
				{
					ID: "bg2", Name: "Background Right", Type: models.LightTypeStrobe,
					Modifier: models.ModifierNone, Role: models.RoleBackground,
					Position: models.Position3D{X: 0.7, Y: 0, Z: 0.7, Distance: 1.0, Angle: 45},
					Power:    90, ColorTemp: 5600, CRI: 90, Enabled: true,
				},
			},
			Panels: []models.Panel{
				{ID: "white_vflat_l", Name: "White V-Flat (Left)", Type: models.PanelBounceWhite, Size: models.PanelSizeLarge,
					Position: models.Position3D{X: -1.30, Y: 0, Z: -0.75, Distance: 1.5, Angle: -120}, Enabled: true},
				{ID: "white_vflat_r", Name: "White V-Flat (Right)", Type: models.PanelBounceWhite, Size: models.PanelSizeLarge,
					Position: models.Position3D{X: 1.30, Y: 0, Z: -0.75, Distance: 1.5, Angle: 120}, Enabled: true},
			},
			Camera:   defaultPortraitCamera(),
			Backdrop: "#ffffff",
			Ambient:  0.3,
		},
	}
}

func lowKeyPortrait() models.Preset {
	return models.Preset{
		ID:       "low_key",
		Name:     "Low-Key Portrait",
		Category: "portrait",
		Description: "Dark, high-contrast setup. Single hard light or gridded softbox. " +
			"Black background with minimal fill. Dramatic, film-noir aesthetic. " +
			"A black V-flat opposite the key absorbs all ambient bounce, while a flag between key and background prevents light wrap.",
		Equipment: []models.EquipmentItem{
			{Role: "Key", Device: "Studio Strobe 500Ws", Modifier: "Softbox + 20° Honeycomb Grid", Power: "85%", Placement: "60° left, above eye level, 2m", Recommended: "Profoto B10 Plus + grid / Godox AD600Pro + grid"},
			{Role: "Accessory", Device: "V-Flat (Black)", Modifier: "Negative fill panel", Power: "N/A", Placement: "Opposite side of key, 1m from subject", Recommended: "V-Flat World Duo Board / DIY foamcore"},
			{Role: "Accessory", Device: "Flag / Cutter (24×36″)", Modifier: "Spill blocker", Power: "N/A", Placement: "Between key and background to prevent light wrap", Recommended: "Matthews Solid Floppy 24×36 / Avenger Cutter"},
			{Role: "Accessory", Device: "Black Seamless Paper (9′)", Modifier: "Background", Power: "N/A", Placement: "3m behind subject, unlit", Recommended: "Savage Seamless #20 Black"},
			{Role: "Accessory", Device: "Sandbag (15 lb)", Modifier: "Stand stability", Power: "N/A", Placement: "Base of all stands", Recommended: "Impact Saddle Sandbag"},
		},
		Scene: models.Scene{
			Mode: models.ModePortrait,
			Lights: []models.Light{
				{
					ID: "key", Name: "Key Light (Gridded)", Type: models.LightTypeStrobe,
					Modifier: models.ModifierHoneycomb, Role: models.RoleKey,
					Position: models.Position3D{X: -1.73, Y: 0.8, Z: -1.0, Distance: 2.0, Angle: -120},
					Power:    85, ColorTemp: 5500, CRI: 95, GridDegree: 20, Enabled: true,
				},
			},
			Panels: []models.Panel{
				{ID: "neg_vflat", Name: "Black V-Flat (Fill Side)", Type: models.PanelNegativeFill, Size: models.PanelSizeLarge,
					Position: models.Position3D{X: 1.0, Y: 0, Z: 0, Distance: 1.0, Angle: 90}, Enabled: true},
				{ID: "flag_bg", Name: "Black Flag (BG Spill Blocker)", Type: models.PanelFlag, Size: models.PanelSizeMedium,
					Position: models.Position3D{X: -0.6, Y: 1.0, Z: 1.04, Distance: 1.2, Angle: -30}, Enabled: true},
			},
			Camera:   defaultPortraitCamera(),
			Backdrop: "#000000",
			Ambient:  0.0,
		},
	}
}

func productTopDown() models.Preset {
	return models.Preset{
		ID:       "product_topdown",
		Name:     "Product Flat Lay / Top-Down",
		Category: "product",
		Description: "Camera directly above product. Two strip boxes on either side for even, " +
			"shadow-free illumination. Ideal for e-commerce flat-lay shots. " +
			"A white bounce below the product eliminates base shadows from the overhead strip lights.",
		Equipment: []models.EquipmentItem{
			{Role: "Strip Left", Device: "Studio Strobe 300Ws", Modifier: "12×48″ Strip Softbox", Power: "60%", Placement: "90° left, overhead, 1.8m", Recommended: "Godox AD300Pro + strip / Profoto RFi Strip"},
			{Role: "Strip Right", Device: "Studio Strobe 300Ws", Modifier: "12×48″ Strip Softbox", Power: "60%", Placement: "90° right, overhead, 1.8m", Recommended: "Godox AD300Pro + strip / Profoto RFi Strip"},
			{Role: "Accessory", Device: "C-Stand with Boom Arm", Modifier: "Camera mount (overhead)", Power: "N/A", Placement: "Directly above subject, camera hangs from boom", Recommended: "Avenger C-Stand + boom arm / Manfrotto 420B"},
			{Role: "Accessory", Device: "White Bounce Card (under)", Modifier: "Under-fill to eliminate base shadows", Power: "N/A", Placement: "Below product, angled up at 30°, off-camera", Recommended: "White foamcore 20×30″"},
			{Role: "Accessory", Device: "Color Checker Passport", Modifier: "Color reference", Power: "N/A", Placement: "First frame in scene for white balance", Recommended: "X-Rite ColorChecker Passport Photo 2"},
			{Role: "Accessory", Device: "Tethering Cable", Modifier: "Live view on laptop", Power: "N/A", Placement: "Camera to laptop for real-time review", Recommended: "TetherPro USB-C / Tether Tools Starter Kit"},
			{Role: "Accessory", Device: "Anti-Static Cloth", Modifier: "Dust/lint removal", Power: "N/A", Placement: "Used between shots on product surface", Recommended: "Kinetronics Anti-Static Brush"},
		},
		Scene: models.Scene{
			Mode: models.ModeProduct,
			Lights: []models.Light{
				{
					ID: "left", Name: "Strip Left", Type: models.LightTypeStrobe,
					Modifier: models.ModifierStripbox, Role: models.RoleKey,
					Position: models.Position3D{X: -1.8, Y: 1.5, Z: 0, Distance: 1.8, Angle: -90},
					Power:    60, ColorTemp: 5500, CRI: 97, Enabled: true,
				},
				{
					ID: "right", Name: "Strip Right", Type: models.LightTypeStrobe,
					Modifier: models.ModifierStripbox, Role: models.RoleFill,
					Position: models.Position3D{X: 1.8, Y: 1.5, Z: 0, Distance: 1.8, Angle: 90},
					Power:    60, ColorTemp: 5500, CRI: 97, Enabled: true,
				},
			},
			Camera: models.CameraSettings{
				FocalLength: 50, Aperture: 8, ShutterSpeed: "1/125",
				ISO: 100, WhiteBalance: 5500, SensorSize: "full_frame",
				AngleX: 90, AngleY: 0, Distance: 1.2,
			},
			Panels: []models.Panel{
				{ID: "under_fill", Name: "White Bounce (Under-Fill)", Type: models.PanelBounceWhite, Size: models.PanelSizeMedium,
					Position: models.Position3D{X: 0, Y: -0.5, Z: -0.5, Distance: 0.5, Angle: 180}, Enabled: true},
			},
			Backdrop: "#ffffff",
			Ambient:  0.05,
		},
	}
}

func productHero() models.Preset {
	return models.Preset{
		ID:       "product_hero",
		Name:     "Product Hero Shot",
		Category: "product",
		Description: "Dramatic product shot with strong key, rim light for edge definition, " +
			"and subtle fill. Creates depth and visual impact for hero/marketing imagery. " +
			"A white bounce on the fill side opens up shadows for detail, while a black card opposite adds depth and prevents flatness.",
		Equipment: []models.EquipmentItem{
			{Role: "Key", Device: "Studio Strobe 500Ws", Modifier: "36″ Softbox with Grid", Power: "70%", Placement: "45° left, above product, 1.5m", Recommended: "Profoto D2 500 / Godox AD600Pro + grid softbox"},
			{Role: "Rim", Device: "Studio Strobe 300Ws", Modifier: "12×36″ Strip Softbox", Power: "50%", Placement: "135° right-rear, 0.5m high, 1.5m", Recommended: "Godox AD300Pro + strip / Profoto RFi Strip"},
			{Role: "Fill", Device: "White Bounce Card", Modifier: "Foam core reflector", Power: "Passive bounce", Placement: "30° right, product level, 1m", Recommended: "V-Flat World Bounce / white foamcore"},
			{Role: "Accessory", Device: "Black Negative Fill Card", Modifier: "Opposite fill, adds contrast/depth", Power: "N/A", Placement: "Behind fill card, blocks ambient from other side", Recommended: "Black foamcore 20×30″ / V-Flat (black side)"},
			{Role: "Accessory", Device: "Product Posing Putty", Modifier: "Holds product at angle", Power: "N/A", Placement: "Under/behind product, hidden from camera", Recommended: "Quake Hold Museum Putty / Blu-Tack"},
			{Role: "Accessory", Device: "Canned Air / Anti-Static Brush", Modifier: "Dust removal", Power: "N/A", Placement: "Used before each shot on product", Recommended: "Falcon Dust-Off / Kinetronics Brush"},
		},
		Scene: models.Scene{
			Mode: models.ModeProduct,
			Lights: []models.Light{
				{
					ID: "key", Name: "Key Softbox", Type: models.LightTypeStrobe,
					Modifier: models.ModifierSoftbox, Role: models.RoleKey,
					Position: models.Position3D{X: -1.06, Y: 1.0, Z: -1.06, Distance: 1.5, Angle: -135},
					Power:    70, ColorTemp: 5500, CRI: 98, Enabled: true,
				},
				{
					ID: "rim", Name: "Rim Strip", Type: models.LightTypeStrobe,
					Modifier: models.ModifierStripbox, Role: models.RoleRim,
					Position: models.Position3D{X: 1.06, Y: 0.5, Z: 1.06, Distance: 1.5, Angle: 45},
					Power:    50, ColorTemp: 5500, CRI: 97, Enabled: true,
				},
				{
					ID: "fill", Name: "Fill Card", Type: models.LightTypeContinuous,
					Modifier: models.ModifierReflector, Role: models.RoleFill,
					Position: models.Position3D{X: 0.50, Y: 0, Z: -0.87, Distance: 1.0, Angle: 150},
					Power:    20, ColorTemp: 5500, CRI: 90, Enabled: true,
				},
			},
			Camera: models.CameraSettings{
				FocalLength: 85, Aperture: 5.6, ShutterSpeed: "1/160",
				ISO: 100, WhiteBalance: 5500, SensorSize: "full_frame",
				AngleX: 15, AngleY: -10, Distance: 1.5,
			},
			Panels: []models.Panel{
				{ID: "fill_bounce", Name: "White Bounce (Fill Side)", Type: models.PanelBounceWhite, Size: models.PanelSizeMedium,
					Position: models.Position3D{X: 0.50, Y: 0, Z: -0.87, Distance: 1.0, Angle: 150}, Enabled: true},
				{ID: "neg_depth", Name: "Black Neg Fill (Opposite)", Type: models.PanelNegativeFill, Size: models.PanelSizeMedium,
					Position: models.Position3D{X: -0.71, Y: 0, Z: 0.71, Distance: 1.0, Angle: -45}, Enabled: true},
			},
			Backdrop: "#1a1a2e",
			Ambient:  0.02,
		},
	}
}

func productWhiteBG() models.Preset {
	return models.Preset{
		ID:       "product_white_bg",
		Name:     "Product on White (E-Commerce)",
		Category: "product",
		Description: "Clean white-background product shot. Background overexposed 1-2 stops. " +
			"Key light with diffusion for soft, even illumination. Amazon/eBay standard. " +
			"A white bounce card camera-side reflects the overhead key back into the product front, filling forward-facing shadows.",
		Equipment: []models.EquipmentItem{
			{Role: "Key", Device: "Studio Strobe 300Ws", Modifier: "4×6′ Diffusion Panel (scrim)", Power: "65%", Placement: "Centered above, 2m", Recommended: "Profoto B10 + scrim jim / Godox AD300Pro + diffuser"},
			{Role: "Background L", Device: "Studio Strobe 300Ws", Modifier: "Bare bulb (no modifier)", Power: "95%", Placement: "Behind product left, aimed at white sweep, 1m", Recommended: "Godox MS300 / Elinchrom D-Lite RX4"},
			{Role: "Background R", Device: "Studio Strobe 300Ws", Modifier: "Bare bulb (no modifier)", Power: "95%", Placement: "Behind product right, aimed at white sweep, 1m", Recommended: "Godox MS300 / Elinchrom D-Lite RX4"},
			{Role: "Accessory", Device: "Shooting Table / White Sweep", Modifier: "Curved white surface", Power: "N/A", Placement: "Product placed on sweep", Recommended: "Foldio Studio / Neewer Shooting Table"},
			{Role: "Accessory", Device: "White Bounce Card (front)", Modifier: "Passive front fill", Power: "N/A", Placement: "Camera side, bounces key back into product front", Recommended: "White foamcore 20×30″"},
			{Role: "Accessory", Device: "Light Meter", Modifier: "Ensure BG is 1–2 stops over key", Power: "N/A", Placement: "Meter at product, then at background", Recommended: "Sekonic L-308X / smartphone app + grey card"},
			{Role: "Accessory", Device: "Masking Tape / Gaffer Tape", Modifier: "Mark product position", Power: "N/A", Placement: "Tape outline on sweep for consistent placement", Recommended: "Pro Gaff / Shurtape P-665"},
			{Role: "Accessory", Device: "Anti-Static Cloth", Modifier: "Dust/fingerprint removal", Power: "N/A", Placement: "Wipe product before each shot", Recommended: "Kinetronics Anti-Static Cloth"},
		},
		Scene: models.Scene{
			Mode: models.ModeProduct,
			Lights: []models.Light{
				{
					ID: "key", Name: "Key Light", Type: models.LightTypeStrobe,
					Modifier: models.ModifierDiffusion, Role: models.RoleKey,
					Position: models.Position3D{X: 0, Y: 1.5, Z: -2.0, Distance: 2.0, Angle: 180},
					Power:    65, ColorTemp: 5500, CRI: 98, Enabled: true,
				},
				{
					ID: "bg1", Name: "BG Light Left", Type: models.LightTypeStrobe,
					Modifier: models.ModifierNone, Role: models.RoleBackground,
					Position: models.Position3D{X: -0.7, Y: 0.5, Z: 0.7, Distance: 1.0, Angle: -45},
					Power:    95, ColorTemp: 5500, CRI: 90, Enabled: true,
				},
				{
					ID: "bg2", Name: "BG Light Right", Type: models.LightTypeStrobe,
					Modifier: models.ModifierNone, Role: models.RoleBackground,
					Position: models.Position3D{X: 0.7, Y: 0.5, Z: 0.7, Distance: 1.0, Angle: 45},
					Power:    95, ColorTemp: 5500, CRI: 90, Enabled: true,
				},
			},
			Camera: models.CameraSettings{
				FocalLength: 70, Aperture: 8, ShutterSpeed: "1/125",
				ISO: 100, WhiteBalance: 5500, SensorSize: "full_frame",
				AngleX: 10, AngleY: 0, Distance: 1.8,
			},
			Panels: []models.Panel{
				{ID: "front_bounce", Name: "White Bounce (Front Fill)", Type: models.PanelBounceWhite, Size: models.PanelSizeMedium,
					Position: models.Position3D{X: 0, Y: -0.3, Z: -1.0, Distance: 1.0, Angle: 180}, Enabled: true},
			},
			Backdrop: "#ffffff",
			Ambient:  0.1,
		},
	}
}

func fashionEditorial() models.Preset {
	return models.Preset{
		ID:       "fashion_editorial",
		Name:     "Fashion Editorial",
		Category: "fashion",
		Description: "Three-light setup: large key source for wrapping light, hair light for separation, " +
			"and kicker for edge definition. A black V-flat opposite the parabolic key creates deeper shadows on the fill side, " +
			"adding the dramatic contrast essential for editorial fashion.",
		Equipment: []models.EquipmentItem{
			{Role: "Key", Device: "Studio Strobe 1000Ws", Modifier: "86″ Parabolic Reflector", Power: "80%", Placement: "30° left, above head, 2.5m", Recommended: "Broncolor Para 88 / Profoto Giant Reflector"},
			{Role: "Hair", Device: "Studio Strobe 300Ws", Modifier: "7″ Reflector + 30° Grid", Power: "50%", Placement: "Behind-above, 2m high, aimed at hair", Recommended: "Godox AD300Pro + grid / Profoto Zoom Reflector"},
			{Role: "Kicker", Device: "Studio Strobe 300Ws", Modifier: "12×36″ Strip Softbox", Power: "40%", Placement: "120° right-rear, waist height, 2m", Recommended: "Godox AD300Pro + strip / Profoto RFi Strip"},
			{Role: "Accessory", Device: "V-Flat (Black side)", Modifier: "Negative fill", Power: "N/A", Placement: "Opposite key, 1.5m from subject", Recommended: "V-Flat World Duo Board"},
			{Role: "Accessory", Device: "Fan (Variable Speed)", Modifier: "Hair movement / fabric flow", Power: "N/A", Placement: "On floor, 3m from subject, angled up", Recommended: "Lasko Pro Performance / Dyson Air Multiplier"},
			{Role: "Accessory", Device: "Garment Steamer", Modifier: "Wrinkle removal between shots", Power: "N/A", Placement: "Off-set, used on clothing before each look", Recommended: "Jiffy J-2000 / Rowenta IS6520"},
			{Role: "Accessory", Device: "Apple Box Set", Modifier: "Height adjustment / posing aid", Power: "N/A", Placement: "Subject stands on for heel height, lean poses", Recommended: "Matthews Apple Box Set / Filmtools"},
		},
		Scene: models.Scene{
			Mode: models.ModeFashion,
			Lights: []models.Light{
				{
					ID: "key", Name: "Key Parabolic", Type: models.LightTypeStrobe,
					Modifier: models.ModifierParabolic, Role: models.RoleKey,
					Position: models.Position3D{X: -1.25, Y: 1.0, Z: -2.17, Distance: 2.5, Angle: -150},
					Power:    80, ColorTemp: 5600, CRI: 98, Enabled: true,
				},
				{
					ID: "hair", Name: "Hair Light", Type: models.LightTypeStrobe,
					Modifier: models.ModifierHoneycomb, Role: models.RoleHair,
					Position: models.Position3D{X: 0.68, Y: 2.0, Z: 1.88, Distance: 2.0, Angle: 20},
					Power:    50, ColorTemp: 5600, CRI: 95, GridDegree: 30, Enabled: true,
				},
				{
					ID: "kicker", Name: "Kicker", Type: models.LightTypeStrobe,
					Modifier: models.ModifierStripbox, Role: models.RoleKicker,
					Position: models.Position3D{X: 1.73, Y: 0.3, Z: 1.0, Distance: 2.0, Angle: 60},
					Power:    40, ColorTemp: 5600, CRI: 95, Enabled: true,
				},
			},
			Camera: models.CameraSettings{
				FocalLength: 85, Aperture: 4, ShutterSpeed: "1/200",
				ISO: 100, WhiteBalance: 5600, SensorSize: "full_frame",
				AngleX: 0, AngleY: 0, Distance: 3.5,
			},
			Panels: []models.Panel{
				{ID: "neg_vflat", Name: "Black V-Flat (Opposite Key)", Type: models.PanelNegativeFill, Size: models.PanelSizeLarge,
					Position: models.Position3D{X: 1.30, Y: 0, Z: -0.75, Distance: 1.5, Angle: 120}, Enabled: true},
			},
			Backdrop: "#e8e0d8",
			Ambient:  0.08,
		},
	}
}

func foodMoody() models.Preset {
	return models.Preset{
		ID:       "food_moody",
		Name:     "Food Photography (Dark & Moody)",
		Category: "food",
		Description: "Back-lit food with a single large diffused source behind and slightly to one side. " +
			"Creates texture, steam visibility, and appetizing highlights. Black flags on both sides control spill " +
			"and deepen shadows for mood, while a small silver card near the dish catches backlight to create appetizing sauce and glaze highlights.",
		Equipment: []models.EquipmentItem{
			{Role: "Key (Backlight)", Device: "Studio Strobe 300Ws", Modifier: "4×6′ Diffusion Panel", Power: "70%", Placement: "150° behind-left, 0.8m above, 1.5m", Recommended: "Profoto B10 + scrim / Godox AD300Pro + diffuser"},
			{Role: "Fill", Device: "White Foam Bounce Card", Modifier: "Foam core reflector", Power: "Passive bounce", Placement: "Opposite backlight, table level, 0.8m", Recommended: "Neewer bounce card / white foamcore 20×30″"},
			{Role: "Accessory", Device: "Black V-Flat (flags) ×2", Modifier: "Negative fill / spill blocker", Power: "N/A", Placement: "Left and right sides to control spill, deepen shadows", Recommended: "Matthews Flag / V-Flat World Duo Board"},
			{Role: "Accessory", Device: "Silver Bounce Card (small)", Modifier: "Specular highlight on dish/sauce", Power: "N/A", Placement: "Close to food, aimed to catch backlight into sauce/glaze", Recommended: "Small mirror 4×6″ on articulating arm / foil-covered card"},
			{Role: "Accessory", Device: "Food Styling Kit", Modifier: "Tweezers, offset spatula, brushes", Power: "N/A", Placement: "On prep table, used to adjust food placement", Recommended: "Mercer Culinary Plating Kit / Tweezerman Tweezers"},
			{Role: "Accessory", Device: "Glycerin Spray Bottle", Modifier: "Faux moisture/dew on food", Power: "N/A", Placement: "Mist on cold items for fresh-from-fridge look", Recommended: "Glycerin + water 50/50 mix in fine-mist sprayer"},
			{Role: "Accessory", Device: "Heat Gun / Torch", Modifier: "Steam simulation / caramelization", Power: "N/A", Placement: "Off-camera, used to create steam from damp cotton or brown surfaces", Recommended: "Kitchen torch / Steamer cotton-ball trick"},
			{Role: "Accessory", Device: "Dark Surface / Textured Board", Modifier: "Background/surface styling", Power: "N/A", Placement: "Food placed on styled surface", Recommended: "Replica Surfaces Dark Concrete / slate tile / dark wood"},
			{Role: "Accessory", Device: "Tethering Cable", Modifier: "Live view on laptop", Power: "N/A", Placement: "Camera to laptop for immediate food styling feedback", Recommended: "TetherPro USB-C / Capture One tethered"},
		},
		Scene: models.Scene{
			Mode: models.ModeFood,
			Lights: []models.Light{
				{
					ID: "key", Name: "Back Light", Type: models.LightTypeStrobe,
					Modifier: models.ModifierDiffusion, Role: models.RoleKey,
					Position: models.Position3D{X: -0.75, Y: 0.8, Z: 1.30, Distance: 1.5, Angle: -30},
					Power:    70, ColorTemp: 5500, CRI: 97, Enabled: true,
				},
				{
					ID: "fill", Name: "Bounce Card", Type: models.LightTypeContinuous,
					Modifier: models.ModifierReflector, Role: models.RoleFill,
					Position: models.Position3D{X: 0.40, Y: 0, Z: -0.69, Distance: 0.8, Angle: 150},
					Power:    15, ColorTemp: 5500, CRI: 90, Enabled: true,
				},
			},
			Camera: models.CameraSettings{
				FocalLength: 100, Aperture: 4, ShutterSpeed: "1/125",
				ISO: 100, WhiteBalance: 5200, SensorSize: "full_frame",
				AngleX: 25, AngleY: -5, Distance: 1.0,
			},
			Panels: []models.Panel{
				{ID: "white_bounce", Name: "White Bounce (Opposite Backlight)", Type: models.PanelBounceWhite, Size: models.PanelSizeMedium,
					Position: models.Position3D{X: 0.40, Y: 0, Z: -0.69, Distance: 0.8, Angle: 150}, Enabled: true},
				{ID: "flag_left", Name: "Black Flag (Left Side)", Type: models.PanelFlag, Size: models.PanelSizeSmall,
					Position: models.Position3D{X: -0.6, Y: 0, Z: 0, Distance: 0.6, Angle: -90}, Enabled: true},
				{ID: "flag_right", Name: "Black Flag (Right Side)", Type: models.PanelFlag, Size: models.PanelSizeSmall,
					Position: models.Position3D{X: 0.69, Y: 0, Z: 0.40, Distance: 0.8, Angle: 60}, Enabled: true},
				{ID: "silver_spec", Name: "Silver Card (Specular Highlight)", Type: models.PanelBounceSilver, Size: models.PanelSizeSmall,
					Position: models.Position3D{X: 0.13, Y: 0.2, Z: -0.48, Distance: 0.5, Angle: 165}, Enabled: true},
			},
			Backdrop: "#1a1a1a",
			Ambient:  0.02,
		},
	}
}

func headshotCorporate() models.Preset {
	return models.Preset{
		ID:       "headshot_corporate",
		Name:     "Corporate Headshot",
		Category: "headshot",
		Description: "Clean, professional two-light setup. Large octabox key with smaller fill. " +
			"Even, flattering light with gentle shadow modeling. White or grey seamless. " +
			"A white bounce card below the camera softens chin and under-eye shadows for a polished, professional result.",
		Equipment: []models.EquipmentItem{
			{Role: "Key", Device: "Studio Strobe 500Ws", Modifier: "47″ Octabox", Power: "65%", Placement: "25° left, slightly above eye level, 1.8m", Recommended: "Profoto D2 500 + OCF Octa / Godox AD600Pro + octa"},
			{Role: "Fill", Device: "Studio Strobe 300Ws", Modifier: "24×36″ Softbox", Power: "35%", Placement: "15° right, eye level, 2.2m", Recommended: "Godox AD300Pro + softbox / Elinchrom ELC 125"},
			{Role: "Accessory", Device: "Grey Seamless Paper (9′ wide)", Modifier: "Background", Power: "N/A", Placement: "3m behind subject, unlit for natural fall-off", Recommended: "Savage Seamless #56 / Lastolite Roll"},
			{Role: "Accessory", Device: "Hair Light (optional)", Modifier: "Gridded strip or snoot", Power: "25%", Placement: "Behind-above subject for hair separation", Recommended: "Godox MS200 + snoot / speedlight + grid"},
			{Role: "Accessory", Device: "White Bounce Board (V-Flat)", Modifier: "Chin fill panel", Power: "N/A", Placement: "Below camera, bounces key into under-chin shadows", Recommended: "V-Flat World (white side) / white foamcore"},
			{Role: "Accessory", Device: "Posing Stool (adjustable)", Modifier: "Height-adjustable seat", Power: "N/A", Placement: "Subject sits with consistent height", Recommended: "Savage Posing Stool / director-style chair"},
			{Role: "Accessory", Device: "Color Checker Passport", Modifier: "Color calibration reference", Power: "N/A", Placement: "First shot of each session", Recommended: "X-Rite ColorChecker Passport Photo 2"},
			{Role: "Accessory", Device: "Tethering Cable + Laptop", Modifier: "Client review in real-time", Power: "N/A", Placement: "Camera to laptop for immediate approval", Recommended: "TetherPro USB-C / Capture One Pro"},
		},
		Scene: models.Scene{
			Mode: models.ModeHeadshot,
			Lights: []models.Light{
				{
					ID: "key", Name: "Key Octabox", Type: models.LightTypeStrobe,
					Modifier: models.ModifierOctabox, Role: models.RoleKey,
					Position: models.Position3D{X: -0.76, Y: 0.6, Z: -1.63, Distance: 1.8, Angle: -155},
					Power:    65, ColorTemp: 5500, CRI: 97, Enabled: true,
				},
				{
					ID: "fill", Name: "Fill Softbox", Type: models.LightTypeStrobe,
					Modifier: models.ModifierSoftbox, Role: models.RoleFill,
					Position: models.Position3D{X: 0.57, Y: 0.2, Z: -2.13, Distance: 2.2, Angle: 165},
					Power:    35, ColorTemp: 5500, CRI: 97, Enabled: true,
				},
			},
			Camera: models.CameraSettings{
				FocalLength: 105, Aperture: 5.6, ShutterSpeed: "1/160",
				ISO: 100, WhiteBalance: 5500, SensorSize: "full_frame",
				AngleX: 0, AngleY: 0, Distance: 2.5,
			},
			Panels: []models.Panel{
				{ID: "chin_bounce", Name: "White Bounce (Below Camera)", Type: models.PanelBounceWhite, Size: models.PanelSizeMedium,
					Position: models.Position3D{X: 0, Y: -0.5, Z: -0.8, Distance: 0.8, Angle: 180}, Enabled: true},
			},
			Backdrop: "#c8c8c8",
			Ambient:  0.1,
		},
	}
}

func rimLightDramatic() models.Preset {
	return models.Preset{
		ID:       "rim_dramatic",
		Name:     "Dramatic Rim / Edge Lighting",
		Category: "portrait",
		Description: "Two rim lights behind the subject with minimal or no front fill. " +
			"Creates a silhouette-like effect with glowing edges. Very cinematic. " +
			"A very subtle white bounce below camera provides just enough fill for eye detail without breaking the silhouette effect.",
		Equipment: []models.EquipmentItem{
			{Role: "Rim Left", Device: "Studio Strobe 300Ws", Modifier: "12×36″ Strip Softbox", Power: "70%", Placement: "135° behind-left, 1.8m", Recommended: "Godox AD300Pro + strip / Profoto RFi 1×3′"},
			{Role: "Rim Right", Device: "Studio Strobe 300Ws", Modifier: "12×36″ Strip Softbox", Power: "70%", Placement: "135° behind-right, 1.8m", Recommended: "Godox AD300Pro + strip / Profoto RFi 1×3′"},
			{Role: "Accessory", Device: "White Bounce Board (subtle)", Modifier: "Front fill for eye detail", Power: "N/A", Placement: "Below camera, angled up — very subtle, preserves silhouette", Recommended: "White foamcore 20×30″ / Lastolite Triflip"},
			{Role: "Accessory", Device: "Black Seamless Paper (9′)", Modifier: "Background", Power: "N/A", Placement: "2m behind subject, unlit", Recommended: "Savage Seamless #20 Black"},
			{Role: "Accessory", Device: "Haze Machine (optional)", Modifier: "Atmospheric volume for rim beams", Power: "N/A", Placement: "Set floor, light haze to reveal rim light rays", Recommended: "Rosco V-Hazer / Ultratec Radiance"},
		},
		Scene: models.Scene{
			Mode: models.ModePortrait,
			Lights: []models.Light{
				{
					ID: "rim_l", Name: "Rim Left", Type: models.LightTypeStrobe,
					Modifier: models.ModifierStripbox, Role: models.RoleRim,
					Position: models.Position3D{X: -1.27, Y: 0.3, Z: 1.27, Distance: 1.8, Angle: -45},
					Power:    70, ColorTemp: 5500, CRI: 95, Enabled: true,
				},
				{
					ID: "rim_r", Name: "Rim Right", Type: models.LightTypeStrobe,
					Modifier: models.ModifierStripbox, Role: models.RoleRim,
					Position: models.Position3D{X: 1.27, Y: 0.3, Z: 1.27, Distance: 1.8, Angle: 45},
					Power:    70, ColorTemp: 5500, CRI: 95, Enabled: true,
				},
			},
			Panels: []models.Panel{
				{ID: "subtle_bounce", Name: "White Bounce (Below Camera)", Type: models.PanelBounceWhite, Size: models.PanelSizeMedium,
					Position: models.Position3D{X: 0, Y: -0.5, Z: -1.0, Distance: 1.0, Angle: 180}, Enabled: true},
			},
			Camera:   defaultPortraitCamera(),
			Backdrop: "#000000",
			Ambient:  0.0,
		},
	}
}

func beautyRingLight() models.Preset {
	return models.Preset{
		ID:       "beauty_ring",
		Name:     "Beauty Ring Light + Accents",
		Category: "portrait",
		Description: "Ring light as key for shadowless front illumination with signature circular catchlights. " +
			"Two accent strips behind for edge separation. Classic beauty/cosmetic look. " +
			"A small white bounce card below the chin catches the ring light's downward spill for full face coverage.",
		Equipment: []models.EquipmentItem{
			{Role: "Key", Device: "18″ LED Ring Light (Bi-Color)", Modifier: "Built-in diffuser", Power: "65%", Placement: "Centered, camera shoots through ring, 1.5m", Recommended: "Neewer 18″ Ring Light / Godox LR160"},
			{Role: "Accent Left", Device: "Studio Strobe 150Ws", Modifier: "12×36″ Strip Softbox", Power: "35%", Placement: "125° behind-left, 1.5m", Recommended: "Godox MS200 + strip / Elinchrom D-Lite RX ONE"},
			{Role: "Accent Right", Device: "Studio Strobe 150Ws", Modifier: "12×36″ Strip Softbox", Power: "35%", Placement: "125° behind-right, 1.5m", Recommended: "Godox MS200 + strip / Elinchrom D-Lite RX ONE"},
			{Role: "Accessory", Device: "White Bounce Board (small)", Modifier: "Chin fill reflector", Power: "N/A", Placement: "Below chin, on table or subject's lap", Recommended: "Lastolite Triflip / white foamcore 12×16″"},
			{Role: "Accessory", Device: "Makeup Table + Mirror", Modifier: "Subject prep station", Power: "N/A", Placement: "Off-set, for beauty/cosmetic touch-ups between shots", Recommended: "Portable director-style makeup table"},
			{Role: "Accessory", Device: "Color Checker", Modifier: "Color reference", Power: "N/A", Placement: "First frame reference shot", Recommended: "X-Rite ColorChecker Classic Mini"},
		},
		Scene: models.Scene{
			Mode: models.ModeHeadshot,
			Lights: []models.Light{
				{
					ID: "ring", Name: "Ring Light", Type: models.LightTypeRingLight,
					Modifier: models.ModifierNone, Role: models.RoleKey,
					Position: models.Position3D{X: 0, Y: 0.3, Z: -1.5, Distance: 1.5, Angle: 180},
					Power:    65, ColorTemp: 5500, CRI: 97, Enabled: true,
				},
				{
					ID: "accent_l", Name: "Accent Left", Type: models.LightTypeStrobe,
					Modifier: models.ModifierStripbox, Role: models.RoleAccent,
					Position: models.Position3D{X: -0.86, Y: 0.5, Z: 1.24, Distance: 1.5, Angle: -35},
					Power:    35, ColorTemp: 5500, CRI: 95, Enabled: true,
				},
				{
					ID: "accent_r", Name: "Accent Right", Type: models.LightTypeStrobe,
					Modifier: models.ModifierStripbox, Role: models.RoleAccent,
					Position: models.Position3D{X: 0.86, Y: 0.5, Z: 1.24, Distance: 1.5, Angle: 35},
					Power:    35, ColorTemp: 5500, CRI: 95, Enabled: true,
				},
			},
			Panels: []models.Panel{
				{ID: "chin_card", Name: "White Bounce (Below Chin)", Type: models.PanelBounceWhite, Size: models.PanelSizeSmall,
					Position: models.Position3D{X: 0, Y: -0.6, Z: -0.5, Distance: 0.5, Angle: 180}, Enabled: true},
			},
			Camera:   defaultPortraitCamera(),
			Backdrop: "#1a1a1a",
			Ambient:  0.05,
		},
	}
}

func cinematicNoir() models.Preset {
	return models.Preset{
		ID:       "cinematic_noir",
		Name:     "Cinematic Film Noir",
		Category: "portrait",
		Description: "Hard key through barn doors creating venetian-blind slit patterns. " +
			"Minimal fill for deep shadows. Warm tungsten color for period atmosphere. " +
			"A black flag camera-side prevents lens flare from the barn-door key, while a V-flat on the fill side ensures deep film-noir shadows.",
		Equipment: []models.EquipmentItem{
			{Role: "Key", Device: "Fresnel Spotlight 500W", Modifier: "4-Leaf Barn Doors", Power: "85%", Placement: "65° left, overhead 1.2m, 2.5m", Recommended: "Godox SL150II + barn doors / Arri 650 Plus Fresnel"},
			{Role: "Background", Device: "Studio Strobe 150Ws", Modifier: "Snoot", Power: "30%", Placement: "Behind subject, aimed at wall, 2m", Recommended: "Godox AD200Pro + snoot / Profoto Snoot"},
			{Role: "Accessory", Device: "CTO Gel (Full)", Modifier: "Tungsten color correction", Power: "N/A", Placement: "On key light", Recommended: "Rosco #3407 Full CTO / Lee 204"},
			{Role: "Accessory", Device: "Cucoloris (Cookie) / Gobo", Modifier: "Shadow pattern projector", Power: "N/A", Placement: "Between key and subject for venetian-blind pattern", Recommended: "Matthews Cucoloris 24×36 / DIY foam cutter"},
			{Role: "Accessory", Device: "Haze Machine", Modifier: "Atmospheric haze for light beams", Power: "N/A", Placement: "Set floor, create even haze before shooting", Recommended: "Rosco V-Hazer / Ultratec Radiance Hazer"},
			{Role: "Accessory", Device: "Black Flag (18×24″)", Modifier: "Negative fill / spill control", Power: "N/A", Placement: "Camera side, prevents lens flare from key", Recommended: "Matthews Solid Flag / Avenger Flag"},
		},
		Scene: models.Scene{
			Mode: models.ModePortrait,
			Lights: []models.Light{
				{
					ID: "key", Name: "Key (Barn Doors)", Type: models.LightTypeStrobe,
					Modifier: models.ModifierBarnDoors, Role: models.RoleKey,
					Position: models.Position3D{X: -2.27, Y: 1.2, Z: -1.06, Distance: 2.5, Angle: -115},
					Power:    85, ColorTemp: 3800, CRI: 92, Enabled: true,
				},
				{
					ID: "bg", Name: "Background Spot", Type: models.LightTypeStrobe,
					Modifier: models.ModifierSnoot, Role: models.RoleBackground,
					Position: models.Position3D{X: 0.68, Y: 1.5, Z: 1.88, Distance: 2.0, Angle: 20},
					Power:    30, ColorTemp: 3800, CRI: 90, Enabled: true,
				},
			},
			Camera: models.CameraSettings{
				FocalLength: 50, Aperture: 2.8, ShutterSpeed: "1/125",
				ISO: 400, WhiteBalance: 3800, SensorSize: "full_frame",
				AngleX: -5, AngleY: 0, Distance: 2.0,
			},
			Panels: []models.Panel{
				{ID: "flag_camera", Name: "Black Flag (Camera Side, Anti-Flare)", Type: models.PanelFlag, Size: models.PanelSizeMedium,
					Position: models.Position3D{X: -0.50, Y: 0.5, Z: -0.87, Distance: 1.0, Angle: -150}, Enabled: true},
				{ID: "neg_vflat", Name: "Black V-Flat (Fill Side)", Type: models.PanelNegativeFill, Size: models.PanelSizeLarge,
					Position: models.Position3D{X: 1.2, Y: 0, Z: 0, Distance: 1.2, Angle: 90}, Enabled: true},
			},
			Backdrop: "#0a0a0a",
			Ambient:  0.0,
		},
	}
}

func crossLighting() models.Preset {
	return models.Preset{
		ID:       "cross_light",
		Name:     "Cross Lighting (Dual Key)",
		Category: "portrait",
		Description: "Two opposing key lights at equal power creating minimal shadows and maximum dimension. " +
			"Each light fills the other's shadow side. Dramatic, symmetrical look. " +
			"Black V-flats behind each softbox prevent rear spill and keep the two-light interplay clean.",
		Equipment: []models.EquipmentItem{
			{Role: "Key Left", Device: "Studio Strobe 300Ws", Modifier: "24×36″ Softbox", Power: "70%", Placement: "60° left, above eye level, 2m", Recommended: "Godox AD300Pro + softbox / Profoto B10"},
			{Role: "Key Right", Device: "Studio Strobe 300Ws", Modifier: "24×36″ Softbox", Power: "70%", Placement: "60° right, above eye level, 2m", Recommended: "Godox AD300Pro + softbox / Profoto B10"},
			{Role: "Hair", Device: "Studio Strobe 150Ws", Modifier: "7″ Reflector + 20° Grid", Power: "40%", Placement: "Directly above, behind subject, 2m", Recommended: "Godox MS200 + grid / Profoto Zoom Reflector"},
			{Role: "Accessory", Device: "V-Flat (Black side) ×2", Modifier: "Negative fill panels", Power: "N/A", Placement: "Behind each softbox, prevents rear spill", Recommended: "V-Flat World Duo Board / black foamcore"},
			{Role: "Accessory", Device: "Light Meter", Modifier: "Match power between both keys", Power: "N/A", Placement: "Subject position, meter each key individually", Recommended: "Sekonic L-308X / L-858D"},
		},
		Scene: models.Scene{
			Mode: models.ModePortrait,
			Lights: []models.Light{
				{
					ID: "key_l", Name: "Key Left", Type: models.LightTypeStrobe,
					Modifier: models.ModifierSoftbox, Role: models.RoleKey,
					Position: models.Position3D{X: -1.73, Y: 0.5, Z: -1.0, Distance: 2.0, Angle: -120},
					Power:    70, ColorTemp: 5500, CRI: 96, Enabled: true,
				},
				{
					ID: "key_r", Name: "Key Right", Type: models.LightTypeStrobe,
					Modifier: models.ModifierSoftbox, Role: models.RoleFill,
					Position: models.Position3D{X: 1.73, Y: 0.5, Z: -1.0, Distance: 2.0, Angle: 120},
					Power:    70, ColorTemp: 5500, CRI: 96, Enabled: true,
				},
				{
					ID: "hair", Name: "Hair Light", Type: models.LightTypeStrobe,
					Modifier: models.ModifierHoneycomb, Role: models.RoleHair,
					Position: models.Position3D{X: 0, Y: 2.0, Z: 2.0, Distance: 2.0, Angle: 0},
					Power:    40, ColorTemp: 5500, CRI: 95, GridDegree: 20, Enabled: true,
				},
			},
			Panels: []models.Panel{
				{ID: "neg_behind_l", Name: "Black V-Flat (Behind Left Key)", Type: models.PanelNegativeFill, Size: models.PanelSizeLarge,
					Position: models.Position3D{X: -2.0, Y: 0, Z: 0, Distance: 2.0, Angle: -90}, Enabled: true},
				{ID: "neg_behind_r", Name: "Black V-Flat (Behind Right Key)", Type: models.PanelNegativeFill, Size: models.PanelSizeLarge,
					Position: models.Position3D{X: 2.0, Y: 0, Z: 0, Distance: 2.0, Angle: 90}, Enabled: true},
			},
			Camera:   defaultPortraitCamera(),
			Backdrop: "#222222",
			Ambient:  0.05,
		},
	}
}

func productGlassware() models.Preset {
	return models.Preset{
		ID:       "product_glass",
		Name:     "Product Glassware / Bottles",
		Category: "product",
		Description: "Back-lit translucent products with two strip boxes behind a diffusion panel. " +
			"Reveals liquid color and glass shape. Tight black cards on both sides create the characteristic dark edge lines " +
			"that define glass form, while a small white card directed at the label ensures text readability.",
		Equipment: []models.EquipmentItem{
			{Role: "Back Strip Left", Device: "Studio Strobe 300Ws", Modifier: "12×48″ Strip Softbox", Power: "60%", Placement: "150° behind-left, 1.5m", Recommended: "Godox AD300Pro + strip / Profoto RFi Strip"},
			{Role: "Back Strip Right", Device: "Studio Strobe 300Ws", Modifier: "12×48″ Strip Softbox", Power: "60%", Placement: "150° behind-right, 1.5m", Recommended: "Godox AD300Pro + strip / Profoto RFi Strip"},
			{Role: "Top Highlight", Device: "Studio Strobe 150Ws", Modifier: "Snoot", Power: "25%", Placement: "Directly above, 2m", Recommended: "Godox MS200 + snoot / Profoto Snoot"},
			{Role: "Accessory", Device: "Black Acrylic Reflection Surface", Modifier: "Mirror-like base for reflection", Power: "N/A", Placement: "Product sits on black acrylic", Recommended: "Black plexiglass sheet 24×36″"},
			{Role: "Accessory", Device: "Black Card Flags (2×)", Modifier: "Edge definition panels (negative fill)", Power: "N/A", Placement: "Both sides, 0.3m from product, creates dark edge lines", Recommended: "Matthews Solid Floppy / black foamcore"},
			{Role: "Accessory", Device: "Diffusion Panel (behind product)", Modifier: "Large scrim between strips and product", Power: "N/A", Placement: "Directly behind product, strips fire through it", Recommended: "Westcott Scrim Jim 4×6′ / DIY diffusion fabric"},
			{Role: "Accessory", Device: "White Card (small)", Modifier: "Label/logo highlight bounce", Power: "N/A", Placement: "Aimed at bottle label to bounce light onto text", Recommended: "Small white card 5×7″ on articulating arm"},
			{Role: "Accessory", Device: "Glycerin Spray + Misting Bottle", Modifier: "Condensation/dew effect on glass", Power: "N/A", Placement: "Spray on bottle before shooting for fresh-pour look", Recommended: "Glycerin + water mix / commercial dew spray"},
			{Role: "Accessory", Device: "Posing Putty / Museum Wax", Modifier: "Holds bottle at exact angle", Power: "N/A", Placement: "Under bottle base, hidden from camera", Recommended: "Quake Hold Museum Putty"},
		},
		Scene: models.Scene{
			Mode: models.ModeProduct,
			Lights: []models.Light{
				{
					ID: "back_l", Name: "Back Strip Left", Type: models.LightTypeStrobe,
					Modifier: models.ModifierStripbox, Role: models.RoleKey,
					Position: models.Position3D{X: -0.75, Y: 0.5, Z: 1.30, Distance: 1.5, Angle: -30},
					Power:    60, ColorTemp: 5500, CRI: 98, Enabled: true,
				},
				{
					ID: "back_r", Name: "Back Strip Right", Type: models.LightTypeStrobe,
					Modifier: models.ModifierStripbox, Role: models.RoleFill,
					Position: models.Position3D{X: 0.75, Y: 0.5, Z: 1.30, Distance: 1.5, Angle: 30},
					Power:    60, ColorTemp: 5500, CRI: 98, Enabled: true,
				},
				{
					ID: "top", Name: "Top Highlight", Type: models.LightTypeStrobe,
					Modifier: models.ModifierSnoot, Role: models.RoleAccent,
					Position: models.Position3D{X: 0, Y: 2.0, Z: 0, Distance: 0.01, Angle: 0},
					Power:    25, ColorTemp: 5500, CRI: 97, Enabled: true,
				},
			},
			Camera: models.CameraSettings{
				FocalLength: 100, Aperture: 11, ShutterSpeed: "1/125",
				ISO: 100, WhiteBalance: 5500, SensorSize: "full_frame",
				AngleX: 5, AngleY: 0, Distance: 1.2,
			},
			Panels: []models.Panel{
				{ID: "edge_l", Name: "Black Card (Left Edge Definition)", Type: models.PanelNegativeFill, Size: models.PanelSizeMedium,
					Position: models.Position3D{X: -0.3, Y: 0, Z: 0, Distance: 0.3, Angle: -90}, Enabled: true},
				{ID: "edge_r", Name: "Black Card (Right Edge Definition)", Type: models.PanelNegativeFill, Size: models.PanelSizeMedium,
					Position: models.Position3D{X: 0.3, Y: 0, Z: 0, Distance: 0.3, Angle: 90}, Enabled: true},
				{ID: "label_bounce", Name: "White Card (Label Highlight)", Type: models.PanelBounceWhite, Size: models.PanelSizeSmall,
					Position: models.Position3D{X: 0, Y: 0, Z: -0.3, Distance: 0.3, Angle: 180}, Enabled: true},
			},
			Backdrop: "#0d0d0d",
			Ambient:  0.0,
		},
	}
}

func fashionCatalog() models.Preset {
	return models.Preset{
		ID:       "fashion_catalog",
		Name:     "Fashion Catalog (Clean)",
		Category: "fashion",
		Description: "Even, clean lighting for catalog/lookbook work. Large overhead key with " +
			"fill from below. White seamless background with two BG lights. Shows garment detail. " +
			"White V-flats flanking the camera bounce the overhead key back onto the garment front for even, detail-revealing illumination.",
		Equipment: []models.EquipmentItem{
			{Role: "Key", Device: "Studio Strobe 500Ws", Modifier: "60″ Octabox", Power: "70%", Placement: "Centered overhead, 2.5m", Recommended: "Profoto D2 500 + Giant Octa / Godox AD600Pro + octa"},
			{Role: "Fill", Device: "Studio Strobe 300Ws", Modifier: "24×36″ Softbox", Power: "40%", Placement: "Below key, 2m", Recommended: "Godox AD300Pro + softbox / Elinchrom ELC 125"},
			{Role: "Background L", Device: "Studio Strobe 300Ws", Modifier: "Bare bulb", Power: "90%", Placement: "Behind left, aimed at white sweep", Recommended: "Godox MS300 / Elinchrom D-Lite RX4"},
			{Role: "Background R", Device: "Studio Strobe 300Ws", Modifier: "Bare bulb", Power: "90%", Placement: "Behind right, aimed at white sweep", Recommended: "Godox MS300 / Elinchrom D-Lite RX4"},
			{Role: "Accessory", Device: "White Seamless Paper (9′ wide)", Modifier: "Full-length backdrop", Power: "N/A", Placement: "Backdrop sweeps onto floor, 4m wide coverage", Recommended: "Savage Seamless #01 Super White"},
			{Role: "Accessory", Device: "White Bounce Board (V-Flat)", Modifier: "Front fill panel", Power: "N/A", Placement: "Both sides of camera, bounce key onto garment front", Recommended: "V-Flat World Duo Board (white side) ×2"},
			{Role: "Accessory", Device: "Garment Steamer", Modifier: "Wrinkle removal", Power: "N/A", Placement: "Off-set, steams each garment before shooting", Recommended: "Jiffy J-2000 / Rowenta IS6520"},
			{Role: "Accessory", Device: "Posing Stool / Apple Box", Modifier: "Seated poses / height variation", Power: "N/A", Placement: "Subject uses for seated catalog poses", Recommended: "Matthews Apple Box / director-style posing stool"},
			{Role: "Accessory", Device: "Clothing Clips / Clamps", Modifier: "Garment fit adjustment", Power: "N/A", Placement: "Back of garment, hidden from camera", Recommended: "Fashion styling clips / binder clips"},
		},
		Scene: models.Scene{
			Mode: models.ModeFashion,
			Lights: []models.Light{
				{
					ID: "key", Name: "Key Overhead", Type: models.LightTypeStrobe,
					Modifier: models.ModifierOctabox, Role: models.RoleKey,
					Position: models.Position3D{X: 0, Y: 1.5, Z: -2.5, Distance: 2.5, Angle: 180},
					Power:    70, ColorTemp: 5600, CRI: 97, Enabled: true,
				},
				{
					ID: "fill", Name: "Fill Below", Type: models.LightTypeStrobe,
					Modifier: models.ModifierSoftbox, Role: models.RoleFill,
					Position: models.Position3D{X: 0, Y: -0.3, Z: -2.0, Distance: 2.0, Angle: 180},
					Power:    40, ColorTemp: 5600, CRI: 97, Enabled: true,
				},
				{
					ID: "bg1", Name: "BG Left", Type: models.LightTypeStrobe,
					Modifier: models.ModifierNone, Role: models.RoleBackground,
					Position: models.Position3D{X: -0.7, Y: 0, Z: 0.7, Distance: 1.0, Angle: -45},
					Power:    90, ColorTemp: 5600, CRI: 90, Enabled: true,
				},
				{
					ID: "bg2", Name: "BG Right", Type: models.LightTypeStrobe,
					Modifier: models.ModifierNone, Role: models.RoleBackground,
					Position: models.Position3D{X: 0.7, Y: 0, Z: 0.7, Distance: 1.0, Angle: 45},
					Power:    90, ColorTemp: 5600, CRI: 90, Enabled: true,
				},
			},
			Camera: models.CameraSettings{
				FocalLength: 70, Aperture: 8, ShutterSpeed: "1/200",
				ISO: 100, WhiteBalance: 5600, SensorSize: "full_frame",
				AngleX: 0, AngleY: 0, Distance: 4.0,
			},
			Panels: []models.Panel{
				{ID: "front_vflat_l", Name: "White V-Flat (Left of Camera)", Type: models.PanelBounceWhite, Size: models.PanelSizeLarge,
					Position: models.Position3D{X: -1.30, Y: 0, Z: -0.75, Distance: 1.5, Angle: -120}, Enabled: true},
				{ID: "front_vflat_r", Name: "White V-Flat (Right of Camera)", Type: models.PanelBounceWhite, Size: models.PanelSizeLarge,
					Position: models.Position3D{X: 1.30, Y: 0, Z: -0.75, Distance: 1.5, Angle: 120}, Enabled: true},
			},
			Backdrop: "#ffffff",
			Ambient:  0.15,
		},
	}
}

func foodBright() models.Preset {
	return models.Preset{
		ID:       "food_bright",
		Name:     "Food Photography (Bright & Airy)",
		Category: "food",
		Description: "Side-lit bright food photography simulating window light. Large diffusion panel as key. " +
			"Clean, Instagram-friendly aesthetic. A white bounce opposite the key fills shadows for an airy look, " +
			"an overhead white panel prevents top shadows, and a small black flag at camera-side adds front depth.",
		Equipment: []models.EquipmentItem{
			{Role: "Key (Window Sim)", Device: "Studio Strobe 300Ws", Modifier: "4×6′ Diffusion Panel", Power: "60%", Placement: "70° left side, 0.8m above, 1.5m", Recommended: "Profoto B10 + scrim jim / Godox AD300Pro + diffuser"},
			{Role: "Fill", Device: "White Foam Bounce Card", Modifier: "Foam core reflector", Power: "Passive bounce", Placement: "Opposite key, table level, 0.8m", Recommended: "V-Flat World (white side) / white foamcore 20×30″"},
			{Role: "Accessory", Device: "Secondary Bounce (above)", Modifier: "Overhead fill reflector", Power: "N/A", Placement: "Above food, bounces key down to fill top shadows", Recommended: "White foamcore on overhead arm / Lastolite Skylite"},
			{Role: "Accessory", Device: "Black Flag (small)", Modifier: "Subtle negative fill", Power: "N/A", Placement: "Camera side, adds depth to front of dish", Recommended: "Black foamcore 12×16″ on mini stand"},
			{Role: "Accessory", Device: "Styling Surface / Wooden Board", Modifier: "Background surface styling", Power: "N/A", Placement: "Food placed on styled surface", Recommended: "Replica Surfaces Light Wood / marble tile / linen cloth"},
			{Role: "Accessory", Device: "Food Styling Kit", Modifier: "Tweezers, offset spatula, brushes, squeeze bottles", Power: "N/A", Placement: "On prep table for garnish and sauce placement", Recommended: "Mercer Culinary Plating Kit / fine-tip squeeze bottles"},
			{Role: "Accessory", Device: "Props Kit", Modifier: "Plates, napkins, cutlery, herbs, ingredients", Power: "N/A", Placement: "Arranged around hero dish for context", Recommended: "Thrift store ceramics / Crate & Barrel props"},
			{Role: "Accessory", Device: "Spray Bottle (water)", Modifier: "Mist freshness on greens/herbs", Power: "N/A", Placement: "Spray herbs and salad just before shooting", Recommended: "Fine-mist spray bottle"},
			{Role: "Accessory", Device: "Overhead Arm / C-Stand Boom", Modifier: "Camera mount for 45° top-down", Power: "N/A", Placement: "Camera mounted on boom for consistent angle", Recommended: "Manfrotto Overhead Kit / Tether Tools Rock Solid"},
			{Role: "Accessory", Device: "Tethering Cable", Modifier: "Live view on laptop", Power: "N/A", Placement: "Real-time composition and styling feedback", Recommended: "TetherPro USB-C / Capture One tethered"},
		},
		Scene: models.Scene{
			Mode: models.ModeFood,
			Lights: []models.Light{
				{
					ID: "key", Name: "Window Light (Side)", Type: models.LightTypeStrobe,
					Modifier: models.ModifierDiffusion, Role: models.RoleKey,
					Position: models.Position3D{X: -1.5, Y: 0.8, Z: 0, Distance: 1.5, Angle: -90},
					Power:    60, ColorTemp: 5800, CRI: 97, Enabled: true,
				},
				{
					ID: "fill", Name: "Bounce Card", Type: models.LightTypeContinuous,
					Modifier: models.ModifierReflector, Role: models.RoleFill,
					Position: models.Position3D{X: 0.57, Y: 0, Z: -0.57, Distance: 0.8, Angle: 135},
					Power:    25, ColorTemp: 5800, CRI: 90, Enabled: true,
				},
			},
			Camera: models.CameraSettings{
				FocalLength: 50, Aperture: 2.8, ShutterSpeed: "1/160",
				ISO: 100, WhiteBalance: 5800, SensorSize: "full_frame",
				AngleX: 45, AngleY: 0, Distance: 0.8,
			},
			Panels: []models.Panel{
				{ID: "bounce_opposite", Name: "White Bounce (Opposite Key)", Type: models.PanelBounceWhite, Size: models.PanelSizeMedium,
					Position: models.Position3D{X: 0.57, Y: 0, Z: -0.57, Distance: 0.8, Angle: 135}, Enabled: true},
				{ID: "overhead_panel", Name: "White Panel (Overhead Fill)", Type: models.PanelBounceWhite, Size: models.PanelSizeLarge,
					Position: models.Position3D{X: 0, Y: 1.0, Z: 0, Distance: 0.01, Angle: 0}, Enabled: true},
				{ID: "front_flag", Name: "Black Flag (Front Depth)", Type: models.PanelFlag, Size: models.PanelSizeSmall,
					Position: models.Position3D{X: 0, Y: -0.2, Z: -0.6, Distance: 0.6, Angle: 180}, Enabled: true},
			},
			Backdrop: "#f5f0eb",
			Ambient:  0.2,
		},
	}
}

func groupPhoto() models.Preset {
	return models.Preset{
		ID:       "group_photo",
		Name:     "Group / Team Photo",
		Category: "group",
		Description: "Wide, even coverage for groups of 4-12 people. Two large umbrellas for broad fill, " +
			"overhead strip for separation. Even illumination across wide area. " +
			"White V-flats on both sides of the camera bounce umbrella light back into the group, ensuring even front-row illumination.",
		Equipment: []models.EquipmentItem{
			{Role: "Key Left", Device: "Studio Strobe 600Ws", Modifier: "60″ Shoot-through Umbrella", Power: "75%", Placement: "45° left, 1m above eye level, 3.5m", Recommended: "Godox AD600Pro + umbrella / Profoto D2 1000"},
			{Role: "Fill Right", Device: "Studio Strobe 600Ws", Modifier: "60″ Shoot-through Umbrella", Power: "60%", Placement: "45° right, 1m above eye level, 3.5m", Recommended: "Godox AD600Pro + umbrella / Profoto D2 1000"},
			{Role: "Hair/Separation", Device: "Studio Strobe 300Ws", Modifier: "12×60″ Strip Softbox", Power: "40%", Placement: "Overhead centered, 2.5m high", Recommended: "Godox AD300Pro + long strip / Profoto RFi Strip"},
			{Role: "Accessory", Device: "Grey/White Seamless (12′ wide)", Modifier: "Wide backdrop for full group", Power: "N/A", Placement: "Full-width backdrop, 4m behind front row", Recommended: "Savage 12′ Seamless / large muslin"},
			{Role: "Accessory", Device: "Step Platform / Risers", Modifier: "Back row elevation", Power: "N/A", Placement: "Back row stands 30cm higher, middle row 15cm", Recommended: "Portable staging risers / sturdy benches / apple boxes"},
			{Role: "Accessory", Device: "White Bounce Board (V-Flat) ×2", Modifier: "Front fill panels", Power: "N/A", Placement: "Both sides of camera, bounces umbrella light back into group", Recommended: "V-Flat World Duo Board (white side)"},
			{Role: "Accessory", Device: "Gaffer Tape (floor marks)", Modifier: "Position markers", Power: "N/A", Placement: "Mark feet positions for each row on floor", Recommended: "Pro Gaff / Shurtape P-665"},
			{Role: "Accessory", Device: "Sandbags (15 lb) ×4", Modifier: "Stand stability (high stands)", Power: "N/A", Placement: "Base of all tall light stands", Recommended: "Impact Saddle Sandbag"},
		},
		Scene: models.Scene{
			Mode: models.ModeGroup,
			Lights: []models.Light{
				{
					ID: "key_l", Name: "Key Left Umbrella", Type: models.LightTypeStrobe,
					Modifier: models.ModifierUmbrella, Role: models.RoleKey,
					Position: models.Position3D{X: -2.47, Y: 1.0, Z: -2.47, Distance: 3.5, Angle: -135},
					Power:    75, ColorTemp: 5500, CRI: 95, Enabled: true,
				},
				{
					ID: "key_r", Name: "Fill Right Umbrella", Type: models.LightTypeStrobe,
					Modifier: models.ModifierUmbrella, Role: models.RoleFill,
					Position: models.Position3D{X: 2.47, Y: 1.0, Z: -2.47, Distance: 3.5, Angle: 135},
					Power:    60, ColorTemp: 5500, CRI: 95, Enabled: true,
				},
				{
					ID: "hair", Name: "Overhead Strip", Type: models.LightTypeStrobe,
					Modifier: models.ModifierStripbox, Role: models.RoleHair,
					Position: models.Position3D{X: 0, Y: 2.5, Z: 0, Distance: 0.01, Angle: 0},
					Power:    40, ColorTemp: 5500, CRI: 95, Enabled: true,
				},
			},
			Camera: models.CameraSettings{
				FocalLength: 35, Aperture: 5.6, ShutterSpeed: "1/200",
				ISO: 200, WhiteBalance: 5500, SensorSize: "full_frame",
				AngleX: 0, AngleY: 0, Distance: 5.0,
			},
			Panels: []models.Panel{
				{ID: "front_vflat_l", Name: "White V-Flat (Left of Camera)", Type: models.PanelBounceWhite, Size: models.PanelSizeLarge,
					Position: models.Position3D{X: -1.73, Y: 0, Z: -1.0, Distance: 2.0, Angle: -120}, Enabled: true},
				{ID: "front_vflat_r", Name: "White V-Flat (Right of Camera)", Type: models.PanelBounceWhite, Size: models.PanelSizeLarge,
					Position: models.Position3D{X: 1.73, Y: 0, Z: -1.0, Distance: 2.0, Angle: 120}, Enabled: true},
			},
			Backdrop: "#c8c8c8",
			Ambient:  0.15,
		},
	}
}

func sportAction() models.Preset {
	return models.Preset{
		ID:       "sport_action",
		Name:     "Sport / Action Portrait",
		Category: "sport",
		Description: "High-power setup for freezing motion. Hard key with grid for directional punch, " +
			"two rim lights for edge separation against dark background. Short flash duration critical. " +
			"A white V-flat below camera bounces the gridded key into the athlete's face, preserving eye detail and catchlights.",
		Equipment: []models.EquipmentItem{
			{Role: "Key", Device: "Studio Strobe 1000Ws (Action mode)", Modifier: "7″ Reflector + 30° Grid", Power: "90%", Placement: "35° left, above head, 2.5m", Recommended: "Profoto Pro-11 / Broncolor Siros 800 (t0.1 ≤ 1/5000s)"},
			{Role: "Rim Left", Device: "Studio Strobe 500Ws", Modifier: "12×36″ Strip Softbox", Power: "65%", Placement: "125° behind-left, 2.5m", Recommended: "Godox AD600Pro + strip / Elinchrom ELC 500"},
			{Role: "Rim Right", Device: "Studio Strobe 500Ws", Modifier: "12×36″ Strip Softbox", Power: "65%", Placement: "125° behind-right, 2.5m", Recommended: "Godox AD600Pro + strip / Elinchrom ELC 500"},
			{Role: "Accessory", Device: "Radio Trigger (HSS capable)", Modifier: "Wireless flash trigger", Power: "N/A", Placement: "Camera hotshoe + receivers on all strobes", Recommended: "Godox X2T / Profoto Air Remote TTL / PocketWizard Plus IV"},
			{Role: "Accessory", Device: "Fog / Haze Machine", Modifier: "Atmosphere for light beam visibility", Power: "N/A", Placement: "Set floor, create even haze before shooting", Recommended: "Rosco V-Hazer / Chauvet Hurricane Haze 2D"},
			{Role: "Accessory", Device: "Black Seamless Paper (12′ wide)", Modifier: "Dark background", Power: "N/A", Placement: "3m behind subject, unlit", Recommended: "Savage Seamless #20 Black (extra-wide)"},
			{Role: "Accessory", Device: "Sandbags (25 lb) ×3", Modifier: "Heavy-duty stand stability", Power: "N/A", Placement: "Base of all tall stands (high-power strobes)", Recommended: "Impact Saddle Sandbag / sand-filled shot bags"},
			{Role: "Accessory", Device: "V-Flat (White side)", Modifier: "Subtle front fill bounce", Power: "N/A", Placement: "Below camera, bounces key into face for eye detail", Recommended: "V-Flat World Duo Board (white side)"},
			{Role: "Accessory", Device: "Gaffer Tape (action marks)", Modifier: "Mark subject position", Power: "N/A", Placement: "Floor marks for jump/action start point", Recommended: "Pro Gaff bright color for visibility"},
		},
		Scene: models.Scene{
			Mode: models.ModeSport,
			Lights: []models.Light{
				{
					ID: "key", Name: "Key (Gridded)", Type: models.LightTypeStrobe,
					Modifier: models.ModifierHoneycomb, Role: models.RoleKey,
					Position: models.Position3D{X: -1.43, Y: 1.0, Z: -2.05, Distance: 2.5, Angle: -145},
					Power:    90, ColorTemp: 5600, CRI: 96, GridDegree: 30, Enabled: true,
				},
				{
					ID: "rim_l", Name: "Rim Left", Type: models.LightTypeStrobe,
					Modifier: models.ModifierStripbox, Role: models.RoleRim,
					Position: models.Position3D{X: -1.43, Y: 0.5, Z: 2.05, Distance: 2.5, Angle: -35},
					Power:    65, ColorTemp: 5600, CRI: 95, Enabled: true,
				},
				{
					ID: "rim_r", Name: "Rim Right", Type: models.LightTypeStrobe,
					Modifier: models.ModifierStripbox, Role: models.RoleRim,
					Position: models.Position3D{X: 1.43, Y: 0.5, Z: 2.05, Distance: 2.5, Angle: 35},
					Power:    65, ColorTemp: 5600, CRI: 95, Enabled: true,
				},
			},
			Camera: models.CameraSettings{
				FocalLength: 70, Aperture: 4, ShutterSpeed: "1/250",
				ISO: 200, WhiteBalance: 5600, SensorSize: "full_frame",
				AngleX: -5, AngleY: 0, Distance: 3.0,
			},
			Panels: []models.Panel{
				{ID: "face_bounce", Name: "White V-Flat (Below Camera)", Type: models.PanelBounceWhite, Size: models.PanelSizeLarge,
					Position: models.Position3D{X: 0, Y: -0.5, Z: -1.5, Distance: 1.5, Angle: 180}, Enabled: true},
			},
			Backdrop: "#0a0a0a",
			Ambient:  0.02,
		},
	}
}

func outdoorGoldenHour() models.Preset {
	return models.Preset{
		ID:       "outdoor_golden_hour",
		Name:     "Outdoor Golden Hour Portrait",
		Category: "outdoor",
		Description: "Beautiful warm backlit portrait shot during golden hour. " +
			"Sun low on the horizon behind the subject creates a warm rim, " +
			"while a silver reflector bounces fill into the face.",
		Equipment: []models.EquipmentItem{
			{Role: "Key / Rim", Device: "Sun", Modifier: "—", Power: "Natural", Placement: "Behind subject, 30° above horizon", Recommended: "Shoot within 1hr of sunset"},
			{Role: "Fill", Device: "Silver Reflector", Modifier: "42″ disc", Power: "Reflected sun", Placement: "Camera-left at 1m, angled up", Recommended: "Silver for punch, white for softer fill"},
		},
		Scene: models.Scene{
			ID: "outdoor_golden_hour", Name: "Golden Hour Portrait", Mode: models.ModeOutdoor,
			Lights: []models.Light{
				{
					ID: "sun", Name: "Sun (Golden Hour)", Type: models.LightTypeSun,
					Modifier: models.ModifierNone, Role: models.RoleKey,
					Position: models.Position3D{X: -1.50, Y: 2.0, Z: 2.60, Distance: 3.0, Angle: -30},
					Power:    75, ColorTemp: 3500, CRI: 100, Enabled: true,
				},
			},
			Camera: models.CameraSettings{
				FocalLength: 85, Aperture: 2.0, ShutterSpeed: "1/250",
				ISO: 100, WhiteBalance: 5000, SensorSize: "full_frame",
				AngleX: -5, AngleY: 0, Distance: 2.5,
			},
			Panels: []models.Panel{
				{ID: "fill_refl", Name: "Silver Reflector", Type: models.PanelBounceSilver, Size: models.PanelSizeMedium,
					Position: models.Position3D{X: -0.71, Y: 0, Z: -0.71, Distance: 1.0, Angle: -135}, Rotation: 45, Enabled: true},
			},
			Backdrop: "#4a3728",
			Ambient:  0.35,
		},
	}
}

func outdoorHarshMidDay() models.Preset {
	return models.Preset{
		ID:       "outdoor_harsh_midday",
		Name:     "Outdoor Harsh Mid-Day",
		Category: "outdoor",
		Description: "Managing harsh overhead sunlight with a diffusion scrim " +
			"to soften shadows and a white bounce card for fill. " +
			"Essential technique for shooting when you can't choose the time of day.",
		Equipment: []models.EquipmentItem{
			{Role: "Key", Device: "Sun", Modifier: "—", Power: "Natural", Placement: "Overhead, near-zenith", Recommended: "Use scrim to tame harsh top light"},
			{Role: "Diffusion", Device: "Scrim Jim", Modifier: "4×6′ overhead frame", Power: "—", Placement: "Directly above subject", Recommended: "1-stop diffusion silk"},
			{Role: "Fill", Device: "White Bounce", Modifier: "V-Flat", Power: "Reflected", Placement: "Camera-right at 1.2m", Recommended: "White foamcore for natural-looking fill"},
		},
		Scene: models.Scene{
			ID: "outdoor_harsh_midday", Name: "Harsh Mid-Day", Mode: models.ModeOutdoor,
			Lights: []models.Light{
				{
					ID: "sun", Name: "Sun (Noon)", Type: models.LightTypeSun,
					Modifier: models.ModifierNone, Role: models.RoleKey,
					Position: models.Position3D{X: 0.17, Y: 3.0, Z: 0.98, Distance: 1.0, Angle: 10},
					Power:    100, ColorTemp: 5600, CRI: 100, Enabled: true,
				},
			},
			Camera: models.CameraSettings{
				FocalLength: 70, Aperture: 4.0, ShutterSpeed: "1/500",
				ISO: 100, WhiteBalance: 5500, SensorSize: "full_frame",
				AngleX: 0, AngleY: 0, Distance: 2.0,
			},
			Panels: []models.Panel{
				{ID: "scrim", Name: "Overhead Scrim", Type: models.PanelDiffusion, Size: models.PanelSizeXLarge,
					Position: models.Position3D{X: 0, Y: 2.0, Z: 0, Distance: 0.01, Angle: 0}, Rotation: 0, Enabled: true},
				{ID: "bounce", Name: "White V-Flat Fill", Type: models.PanelBounceWhite, Size: models.PanelSizeLarge,
					Position: models.Position3D{X: 1.04, Y: 0, Z: -0.60, Distance: 1.2, Angle: 120}, Rotation: 300, Enabled: true},
			},
			Backdrop: "#87CEEB",
			Ambient:  0.45,
		},
	}
}

func outdoorOpenShade() models.Preset {
	return models.Preset{
		ID:       "outdoor_open_shade",
		Name:     "Outdoor Open Shade Portrait",
		Category: "outdoor",
		Description: "Portrait in open shade with a reflector to add directional light. " +
			"Open shade provides even, soft ambient light, while a gold reflector " +
			"adds warmth and dimensionality to the face.",
		Equipment: []models.EquipmentItem{
			{Role: "Ambient", Device: "Sun (blocked by structure)", Modifier: "—", Power: "Natural", Placement: "Diffused by shade", Recommended: "Position subject at edge of shade facing open sky"},
			{Role: "Key Fill", Device: "Gold Reflector", Modifier: "42″ disc", Power: "Reflected sky", Placement: "Camera-right at 1m, high", Recommended: "Gold adds warmth in cool shade"},
		},
		Scene: models.Scene{
			ID: "outdoor_open_shade", Name: "Open Shade", Mode: models.ModeOutdoor,
			Lights: []models.Light{
				{
					ID: "ambient_sky", Name: "Ambient Sky", Type: models.LightTypeSun,
					Modifier: models.ModifierNone, Role: models.RoleFill,
					Position: models.Position3D{X: 0, Y: 3.0, Z: 0.5, Distance: 3.5, Angle: 0},
					Power:    30, ColorTemp: 6500, CRI: 100, Enabled: true,
				},
			},
			Camera: models.CameraSettings{
				FocalLength: 85, Aperture: 1.8, ShutterSpeed: "1/500",
				ISO: 200, WhiteBalance: 5800, SensorSize: "full_frame",
				AngleX: -5, AngleY: 0, Distance: 2.5,
			},
			Panels: []models.Panel{
				{ID: "gold_bounce", Name: "Gold Reflector", Type: models.PanelBounceGold, Size: models.PanelSizeMedium,
					Position: models.Position3D{X: 0.82, Y: 0.5, Z: -0.57, Distance: 1.0, Angle: 125}, Rotation: 240, Enabled: true},
			},
			Backdrop: "#5a7a4a",
			Ambient:  0.4,
		},
	}
}

func defaultPortraitCamera() models.CameraSettings {
	return models.CameraSettings{
		FocalLength:  85,
		Aperture:     2.8,
		ShutterSpeed: "1/200",
		ISO:          100,
		WhiteBalance: 5500,
		SensorSize:   "full_frame",
		AngleX:       0,
		AngleY:       0,
		Distance:     2.5,
	}
}
