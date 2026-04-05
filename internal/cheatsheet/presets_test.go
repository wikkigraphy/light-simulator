package cheatsheet

import (
	"testing"
)

func TestAllPresetsNotEmpty(t *testing.T) {
	presets := AllPresets()
	if len(presets) == 0 {
		t.Fatal("AllPresets returned empty slice")
	}
}

func TestAllPresetsCount(t *testing.T) {
	presets := AllPresets()
	if got := len(presets); got < 24 {
		t.Errorf("expected at least 24 presets, got %d", got)
	}
}

func TestAllPresetsHaveRequiredFields(t *testing.T) {
	for _, p := range AllPresets() {
		if p.ID == "" {
			t.Errorf("preset has empty ID: %+v", p)
		}
		if p.Name == "" {
			t.Errorf("preset %q has empty Name", p.ID)
		}
		if p.Category == "" {
			t.Errorf("preset %q has empty Category", p.ID)
		}
		if p.Description == "" {
			t.Errorf("preset %q has empty Description", p.ID)
		}
		if len(p.Scene.Lights) == 0 {
			t.Errorf("preset %q has no lights", p.ID)
		}
	}
}

func TestAllPresetsUniqueIDs(t *testing.T) {
	seen := make(map[string]bool)
	for _, p := range AllPresets() {
		if seen[p.ID] {
			t.Errorf("duplicate preset ID: %q", p.ID)
		}
		seen[p.ID] = true
	}
}

func TestAllPresetsLightsHaveIDs(t *testing.T) {
	for _, p := range AllPresets() {
		for i, l := range p.Scene.Lights {
			if l.ID == "" {
				t.Errorf("preset %q light[%d] has empty ID", p.ID, i)
			}
			if l.Name == "" {
				t.Errorf("preset %q light[%d] has empty Name", p.ID, i)
			}
			if !l.Enabled {
				t.Errorf("preset %q light[%d] %q is disabled by default", p.ID, i, l.Name)
			}
		}
	}
}

func TestAllPresetsHaveValidPower(t *testing.T) {
	for _, p := range AllPresets() {
		for _, l := range p.Scene.Lights {
			if l.Power < 0 || l.Power > 100 {
				t.Errorf("preset %q light %q has power %f outside [0,100]", p.ID, l.Name, l.Power)
			}
		}
	}
}

func TestAllPresetsHaveValidColorTemp(t *testing.T) {
	for _, p := range AllPresets() {
		for _, l := range p.Scene.Lights {
			if l.ColorTemp < 1800 || l.ColorTemp > 10000 {
				t.Errorf("preset %q light %q has color temp %d outside [1800,10000]", p.ID, l.Name, l.ColorTemp)
			}
		}
	}
}

func TestPresetsByCategory(t *testing.T) {
	categories := PresetsByCategory()
	if len(categories) == 0 {
		t.Fatal("PresetsByCategory returned empty map")
	}

	expectedCategories := []string{"portrait", "product", "fashion", "food", "headshot"}
	for _, cat := range expectedCategories {
		if _, ok := categories[cat]; !ok {
			t.Errorf("missing expected category: %q", cat)
		}
	}
}

func TestPresetsByCategoryTotalMatchesAll(t *testing.T) {
	categories := PresetsByCategory()
	total := 0
	for _, presets := range categories {
		total += len(presets)
	}

	allCount := len(AllPresets())
	if total != allCount {
		t.Errorf("category total %d != AllPresets count %d", total, allCount)
	}
}

func TestNewCategoriesExist(t *testing.T) {
	categories := PresetsByCategory()
	for _, cat := range []string{"group", "sport"} {
		if presets, ok := categories[cat]; !ok || len(presets) == 0 {
			t.Errorf("expected category %q with presets", cat)
		}
	}
}

func TestAllPresetsHaveEquipment(t *testing.T) {
	for _, p := range AllPresets() {
		if len(p.Equipment) == 0 {
			t.Errorf("preset %q has no equipment list", p.ID)
		}
	}
}

func TestAllPresetsEquipmentHaveRequiredFields(t *testing.T) {
	for _, p := range AllPresets() {
		for i, eq := range p.Equipment {
			if eq.Role == "" {
				t.Errorf("preset %q equipment[%d] has empty Role", p.ID, i)
			}
			if eq.Device == "" {
				t.Errorf("preset %q equipment[%d] has empty Device", p.ID, i)
			}
			if eq.Modifier == "" {
				t.Errorf("preset %q equipment[%d] has empty Modifier", p.ID, i)
			}
			if eq.Placement == "" {
				t.Errorf("preset %q equipment[%d] has empty Placement", p.ID, i)
			}
			if eq.Recommended == "" {
				t.Errorf("preset %q equipment[%d] has empty Recommended", p.ID, i)
			}
		}
	}
}

func TestAllPresetsEquipmentHasAtLeastOneLight(t *testing.T) {
	for _, p := range AllPresets() {
		hasLight := false
		for _, eq := range p.Equipment {
			if eq.Role != "Accessory" {
				hasLight = true
				break
			}
		}
		if !hasLight {
			t.Errorf("preset %q equipment has no light device (only accessories)", p.ID)
		}
	}
}

func TestSpecificPresetsExist(t *testing.T) {
	ids := map[string]bool{
		"rembrandt": false, "beauty_ring": false, "cinematic_noir": false,
		"cross_light": false, "product_glass": false, "fashion_catalog": false,
		"food_bright": false, "group_photo": false, "sport_action": false,
	}
	for _, p := range AllPresets() {
		if _, want := ids[p.ID]; want {
			ids[p.ID] = true
		}
	}
	for id, found := range ids {
		if !found {
			t.Errorf("preset %q not found in AllPresets", id)
		}
	}
}
