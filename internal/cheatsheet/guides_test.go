package cheatsheet

import (
	"testing"
)

func TestFlashGuidesNotEmpty(t *testing.T) {
	guides := FlashGuides()
	if len(guides) == 0 {
		t.Fatal("FlashGuides returned empty slice")
	}
}

func TestFlashGuidesHaveRequiredFields(t *testing.T) {
	for i, g := range FlashGuides() {
		if g.Category == "" {
			t.Errorf("guide[%d] has empty Category", i)
		}
		if g.Title == "" {
			t.Errorf("guide[%d] has empty Title", i)
		}
		if g.Description == "" {
			t.Errorf("guide[%d] %q has empty Description", i, g.Title)
		}
		if len(g.BestFor) == 0 {
			t.Errorf("guide[%d] %q has empty BestFor", i, g.Title)
		}
		if g.PowerRange == "" {
			t.Errorf("guide[%d] %q has empty PowerRange", i, g.Title)
		}
		if len(g.Tips) == 0 {
			t.Errorf("guide[%d] %q has empty Tips", i, g.Title)
		}
	}
}

func TestFlashGuidesUniqueCategories(t *testing.T) {
	seen := make(map[string]bool)
	for _, g := range FlashGuides() {
		if seen[g.Category] {
			t.Errorf("duplicate flash guide category: %q", g.Category)
		}
		seen[g.Category] = true
	}
}

func TestModifierGuidesNotEmpty(t *testing.T) {
	guides := ModifierGuides()
	if len(guides) == 0 {
		t.Fatal("ModifierGuides returned empty slice")
	}
}

func TestModifierGuidesHaveRequiredFields(t *testing.T) {
	for i, g := range ModifierGuides() {
		if g.Name == "" {
			t.Errorf("modifier[%d] has empty Name", i)
		}
		if g.Type == "" {
			t.Errorf("modifier[%d] %q has empty Type", i, g.Name)
		}
		if g.Softness == "" {
			t.Errorf("modifier[%d] %q has empty Softness", i, g.Name)
		}
		if g.Catchlight == "" {
			t.Errorf("modifier[%d] %q has empty Catchlight", i, g.Name)
		}
		if len(g.BestFor) == 0 {
			t.Errorf("modifier[%d] %q has empty BestFor", i, g.Name)
		}
		if len(g.ProTips) == 0 {
			t.Errorf("modifier[%d] %q has empty ProTips", i, g.Name)
		}
	}
}

func TestModifierGuidesUniqueTypes(t *testing.T) {
	seen := make(map[string]bool)
	for _, g := range ModifierGuides() {
		if seen[g.Type] {
			t.Errorf("duplicate modifier type: %q", g.Type)
		}
		seen[g.Type] = true
	}
}

func TestLensGuidesNotEmpty(t *testing.T) {
	guides := LensGuides()
	if len(guides) == 0 {
		t.Fatal("LensGuides returned empty slice")
	}
}

func TestLensGuidesHaveRequiredFields(t *testing.T) {
	for i, g := range LensGuides() {
		if g.FocalLength == "" {
			t.Errorf("lens[%d] has empty FocalLength", i)
		}
		if g.Type == "" {
			t.Errorf("lens[%d] %q has empty Type", i, g.FocalLength)
		}
		if len(g.BestFor) == 0 {
			t.Errorf("lens[%d] %q has empty BestFor", i, g.FocalLength)
		}
		if g.DOFNotes == "" {
			t.Errorf("lens[%d] %q has empty DOFNotes", i, g.FocalLength)
		}
		if g.Distortion == "" {
			t.Errorf("lens[%d] %q has empty Distortion", i, g.FocalLength)
		}
	}
}
