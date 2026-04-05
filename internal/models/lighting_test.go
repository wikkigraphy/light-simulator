package models

import (
	"encoding/json"
	"testing"
)

func TestLightTypeConstants(t *testing.T) {
	types := []LightType{
		LightTypeSpeedlight, LightTypeStrobe, LightTypeContinuous,
		LightTypeLED, LightTypeRingLight, LightTypeNatural,
	}
	seen := make(map[LightType]bool)
	for _, lt := range types {
		if lt == "" {
			t.Error("empty light type constant")
		}
		if seen[lt] {
			t.Errorf("duplicate light type: %q", lt)
		}
		seen[lt] = true
	}
}

func TestModifierTypeConstants(t *testing.T) {
	modifiers := []ModifierType{
		ModifierSoftbox, ModifierOctabox, ModifierStripbox,
		ModifierUmbrella, ModifierBeautyDish, ModifierHoneycomb,
		ModifierSnoot, ModifierBarnDoors, ModifierDiffusion,
		ModifierReflector, ModifierParabolic, ModifierNone,
	}
	seen := make(map[ModifierType]bool)
	for _, m := range modifiers {
		if m == "" {
			t.Error("empty modifier type constant")
		}
		if seen[m] {
			t.Errorf("duplicate modifier type: %q", m)
		}
		seen[m] = true
	}
	if len(modifiers) != 12 {
		t.Errorf("expected 12 modifier types, got %d", len(modifiers))
	}
}

func TestLightRoleConstants(t *testing.T) {
	roles := []LightRole{
		RoleKey, RoleFill, RoleRim, RoleHair,
		RoleBackground, RoleAccent, RoleKicker,
	}
	seen := make(map[LightRole]bool)
	for _, r := range roles {
		if r == "" {
			t.Error("empty light role constant")
		}
		if seen[r] {
			t.Errorf("duplicate light role: %q", r)
		}
		seen[r] = true
	}
}

func TestShootModeConstants(t *testing.T) {
	modes := []ShootMode{
		ModePortrait, ModeProduct, ModeFashion, ModeFood,
		ModeHeadshot, ModeGroup, ModeBoudoir, ModeSport,
	}
	seen := make(map[ShootMode]bool)
	for _, m := range modes {
		if m == "" {
			t.Error("empty shoot mode constant")
		}
		if seen[m] {
			t.Errorf("duplicate shoot mode: %q", m)
		}
		seen[m] = true
	}
	if len(modes) != 8 {
		t.Errorf("expected 8 shoot modes, got %d", len(modes))
	}
}

func TestLightJSONSerialization(t *testing.T) {
	light := Light{
		ID: "key_1", Name: "Key Light", Type: LightTypeStrobe,
		Modifier: ModifierSoftbox, Role: RoleKey,
		Position: Position3D{X: -1.5, Y: 0.5, Z: 1.5, Distance: 2.0, Angle: 45},
		Power:    75, ColorTemp: 5500, CRI: 95,
		Enabled: true,
	}

	data, err := json.Marshal(light)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded Light
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.ID != light.ID {
		t.Errorf("ID: expected %q, got %q", light.ID, decoded.ID)
	}
	if decoded.Type != light.Type {
		t.Errorf("Type: expected %q, got %q", light.Type, decoded.Type)
	}
	if decoded.Modifier != light.Modifier {
		t.Errorf("Modifier: expected %q, got %q", light.Modifier, decoded.Modifier)
	}
	if decoded.Position.Distance != light.Position.Distance {
		t.Errorf("Distance: expected %f, got %f", light.Position.Distance, decoded.Position.Distance)
	}
	if decoded.Power != light.Power {
		t.Errorf("Power: expected %f, got %f", light.Power, decoded.Power)
	}
}

func TestSceneJSONSerialization(t *testing.T) {
	scene := Scene{
		ID: "test", Name: "Test Scene", Mode: ModePortrait,
		Lights: []Light{
			{ID: "key", Name: "Key", Type: LightTypeStrobe, Modifier: ModifierSoftbox,
				Role: RoleKey, Position: Position3D{Distance: 2}, Power: 70, ColorTemp: 5500, Enabled: true},
		},
		Camera: CameraSettings{
			FocalLength: 85, Aperture: 2.8, ShutterSpeed: "1/200",
			ISO: 100, WhiteBalance: 5500, SensorSize: "full_frame", Distance: 2.5,
		},
		Backdrop: "#1a1a1a",
		Ambient:  0.1,
	}

	data, err := json.Marshal(scene)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded Scene
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.Mode != scene.Mode {
		t.Errorf("Mode: expected %q, got %q", scene.Mode, decoded.Mode)
	}
	if len(decoded.Lights) != 1 {
		t.Fatalf("expected 1 light, got %d", len(decoded.Lights))
	}
	if decoded.Camera.FocalLength != 85 {
		t.Errorf("FocalLength: expected 85, got %d", decoded.Camera.FocalLength)
	}
	if decoded.Backdrop != scene.Backdrop {
		t.Errorf("Backdrop: expected %q, got %q", scene.Backdrop, decoded.Backdrop)
	}
}

func TestPresetJSONSerialization(t *testing.T) {
	preset := Preset{
		ID: "rembrandt", Name: "Rembrandt", Category: "portrait",
		Description: "Classic portrait lighting",
		Equipment: []EquipmentItem{
			{Role: "Key", Device: "Strobe 300Ws", Modifier: "Softbox", Power: "75%", Placement: "45° left", Recommended: "Godox AD300Pro"},
			{Role: "Fill", Device: "Reflector", Modifier: "Silver disc", Power: "Passive", Placement: "Right side", Recommended: "Neewer 42″"},
		},
		Scene: Scene{
			Mode:   ModePortrait,
			Lights: []Light{{ID: "key", Enabled: true}},
		},
	}

	data, err := json.Marshal(preset)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded Preset
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.ID != preset.ID {
		t.Errorf("ID: expected %q, got %q", preset.ID, decoded.ID)
	}
	if decoded.Category != preset.Category {
		t.Errorf("Category: expected %q, got %q", preset.Category, decoded.Category)
	}
	if len(decoded.Equipment) != 2 {
		t.Fatalf("Equipment: expected 2 items, got %d", len(decoded.Equipment))
	}
	if decoded.Equipment[0].Role != "Key" {
		t.Errorf("Equipment[0].Role: expected %q, got %q", "Key", decoded.Equipment[0].Role)
	}
	if decoded.Equipment[1].Recommended != "Neewer 42″" {
		t.Errorf("Equipment[1].Recommended: expected %q, got %q", "Neewer 42″", decoded.Equipment[1].Recommended)
	}
}

func TestEquipmentItemJSONSerialization(t *testing.T) {
	item := EquipmentItem{
		Role: "Key", Device: "Studio Strobe 500Ws", Modifier: "24×36″ Softbox",
		Power: "80%", Placement: "45° left, 2m", Recommended: "Profoto B10 Plus",
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	s := string(data)
	for _, field := range []string{`"role"`, `"device"`, `"modifier"`, `"power"`, `"placement"`, `"recommended"`} {
		if !contains(s, field) {
			t.Errorf("expected JSON to contain field %s", field)
		}
	}

	var decoded EquipmentItem
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.Device != item.Device {
		t.Errorf("Device: expected %q, got %q", item.Device, decoded.Device)
	}
}

func TestPresetWithEmptyEquipment(t *testing.T) {
	preset := Preset{
		ID: "test", Name: "Test", Category: "test",
		Description: "Test preset",
		Scene:       Scene{Mode: ModePortrait},
	}

	data, err := json.Marshal(preset)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded Preset
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.Equipment != nil {
		t.Errorf("expected nil Equipment, got %v", decoded.Equipment)
	}
}

func TestPanelTypeConstants(t *testing.T) {
	types := []PanelType{
		PanelNegativeFill, PanelBounceWhite, PanelBounceSilver,
		PanelBounceGold, PanelDiffusion, PanelFlag,
	}
	seen := make(map[PanelType]bool)
	for _, pt := range types {
		if pt == "" {
			t.Error("empty panel type constant")
		}
		if seen[pt] {
			t.Errorf("duplicate panel type: %q", pt)
		}
		seen[pt] = true
	}
	if len(types) != 6 {
		t.Errorf("expected 6 panel types, got %d", len(types))
	}
}

func TestPanelSizeConstants(t *testing.T) {
	sizes := []PanelSize{
		PanelSizeSmall, PanelSizeMedium, PanelSizeLarge, PanelSizeXLarge,
	}
	seen := make(map[PanelSize]bool)
	for _, ps := range sizes {
		if ps == "" {
			t.Error("empty panel size constant")
		}
		if seen[ps] {
			t.Errorf("duplicate panel size: %q", ps)
		}
		seen[ps] = true
	}
	if len(sizes) != 4 {
		t.Errorf("expected 4 panel sizes, got %d", len(sizes))
	}
}

func TestPanelJSONSerialization(t *testing.T) {
	panel := Panel{
		ID: "neg1", Name: "Black V-Flat", Type: PanelNegativeFill,
		Size:     PanelSizeLarge,
		Position: Position3D{X: 1.0, Y: 0, Z: 0.5, Distance: 1.0, Angle: -90},
		Rotation: 90,
		Enabled:  true,
	}

	data, err := json.Marshal(panel)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded Panel
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.ID != panel.ID {
		t.Errorf("ID: expected %q, got %q", panel.ID, decoded.ID)
	}
	if decoded.Type != panel.Type {
		t.Errorf("Type: expected %q, got %q", panel.Type, decoded.Type)
	}
	if decoded.Size != panel.Size {
		t.Errorf("Size: expected %q, got %q", panel.Size, decoded.Size)
	}
	if decoded.Position.Distance != panel.Position.Distance {
		t.Errorf("Distance: expected %f, got %f", panel.Position.Distance, decoded.Position.Distance)
	}
	if decoded.Rotation != panel.Rotation {
		t.Errorf("Rotation: expected %f, got %f", panel.Rotation, decoded.Rotation)
	}
	if !decoded.Enabled {
		t.Error("expected Enabled=true")
	}
}

func TestSceneWithPanelsJSONSerialization(t *testing.T) {
	scene := Scene{
		ID: "test_panels", Name: "Scene with Panels", Mode: ModePortrait,
		Lights: []Light{
			{ID: "key", Name: "Key", Type: LightTypeStrobe, Modifier: ModifierSoftbox,
				Role: RoleKey, Position: Position3D{Distance: 2}, Power: 70, ColorTemp: 5500, Enabled: true},
		},
		Panels: []Panel{
			{ID: "vflat", Name: "Black V-Flat", Type: PanelNegativeFill, Size: PanelSizeLarge,
				Position: Position3D{X: 1, Y: 0, Z: 0.5, Distance: 1.0, Angle: -90}, Enabled: true},
			{ID: "bounce", Name: "White Card", Type: PanelBounceWhite, Size: PanelSizeMedium,
				Position: Position3D{X: 0, Y: -0.5, Z: 1.0, Distance: 0.8, Angle: 0}, Enabled: true},
		},
		Camera:   CameraSettings{FocalLength: 85, Aperture: 2.8, ShutterSpeed: "1/200", ISO: 100, WhiteBalance: 5500, SensorSize: "full_frame", Distance: 2.5},
		Backdrop: "#1a1a1a",
		Ambient:  0.1,
	}

	data, err := json.Marshal(scene)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded Scene
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(decoded.Panels) != 2 {
		t.Fatalf("expected 2 panels, got %d", len(decoded.Panels))
	}
	if decoded.Panels[0].Type != PanelNegativeFill {
		t.Errorf("Panel[0].Type: expected %q, got %q", PanelNegativeFill, decoded.Panels[0].Type)
	}
	if decoded.Panels[1].Type != PanelBounceWhite {
		t.Errorf("Panel[1].Type: expected %q, got %q", PanelBounceWhite, decoded.Panels[1].Type)
	}
}

func TestSceneWithoutPanelsOmitsField(t *testing.T) {
	scene := Scene{
		ID: "no_panels", Mode: ModePortrait,
		Lights: []Light{{ID: "key", Enabled: true}},
	}

	data, err := json.Marshal(scene)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	s := string(data)
	if contains(s, `"panels"`) {
		t.Error("expected panels field to be omitted when empty")
	}
}

func TestPosition3DJSONFields(t *testing.T) {
	pos := Position3D{X: 1.5, Y: 0.5, Z: -2.0, Distance: 3.0, Angle: 135}
	data, err := json.Marshal(pos)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	s := string(data)
	for _, field := range []string{`"x"`, `"y"`, `"z"`, `"distance"`, `"angle"`} {
		if !contains(s, field) {
			t.Errorf("expected JSON to contain field %s", field)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
