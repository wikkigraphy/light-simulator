package imgproc

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func makeTestImage(width, height int, bg, fg color.NRGBA) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := range height {
		for x := range width {
			img.SetNRGBA(x, y, bg)
		}
	}
	cx := width / 2
	cy := height / 2
	radius := width / 4
	for y := cy - radius; y <= cy+radius; y++ {
		for x := cx - radius; x <= cx+radius; x++ {
			dx := x - cx
			dy := y - cy
			if dx*dx+dy*dy <= radius*radius {
				img.SetNRGBA(x, y, fg)
			}
		}
	}
	return img
}

func TestRemoveBackgroundFromImage(t *testing.T) {
	bg := color.NRGBA{R: 200, G: 180, B: 160, A: 255}
	fg := color.NRGBA{R: 50, G: 50, B: 50, A: 255}
	src := makeTestImage(100, 100, bg, fg)

	result := removeBackgroundFromImage(src)

	bounds := result.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 100 {
		t.Errorf("expected 100x100, got %dx%d", bounds.Dx(), bounds.Dy())
	}

	cornerPixel := result.NRGBAAt(0, 0)
	if cornerPixel.A > 50 {
		t.Errorf("corner should be mostly transparent (bg removed), got alpha=%d", cornerPixel.A)
	}

	centerPixel := result.NRGBAAt(50, 50)
	if centerPixel.A < 200 {
		t.Errorf("center should be mostly opaque (subject), got alpha=%d", centerPixel.A)
	}
}

func TestRemoveBackgroundPreservesSubject(t *testing.T) {
	bg := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	fg := color.NRGBA{R: 10, G: 10, B: 10, A: 255}
	src := makeTestImage(200, 200, bg, fg)

	result := removeBackgroundFromImage(src)

	subjectPixel := result.NRGBAAt(100, 100)
	if subjectPixel.A < 200 {
		t.Errorf("dark subject on white BG should be preserved, got alpha=%d", subjectPixel.A)
	}
	if subjectPixel.R > 30 || subjectPixel.G > 30 || subjectPixel.B > 30 {
		t.Error("subject color should be preserved as dark")
	}
}

func TestSampleBackgroundColor(t *testing.T) {
	bg := color.NRGBA{R: 180, G: 160, B: 140, A: 255}
	fg := color.NRGBA{R: 50, G: 50, B: 50, A: 255}
	img := makeTestImage(100, 100, bg, fg)

	sample := sampleBackgroundColor(img)

	if diff := colorDistance(sample.R, sample.G, sample.B, bg.R, bg.G, bg.B); diff > 15 {
		t.Errorf("sampled bg (%d,%d,%d) too far from actual bg (%d,%d,%d), dist=%.1f",
			sample.R, sample.G, sample.B, bg.R, bg.G, bg.B, diff)
	}
}

func TestSampleBackgroundColorSmallImage(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	for y := range 2 {
		for x := range 2 {
			img.SetNRGBA(x, y, color.NRGBA{R: 100, G: 100, B: 100, A: 255})
		}
	}
	sample := sampleBackgroundColor(img)
	if sample.R < 90 || sample.R > 110 {
		t.Errorf("expected ~100, got R=%d", sample.R)
	}
}

func TestColorDistance(t *testing.T) {
	cases := []struct {
		r1, g1, b1, r2, g2, b2 uint8
		wantMin, wantMax       float64
	}{
		{0, 0, 0, 0, 0, 0, 0, 0.1},
		{255, 255, 255, 0, 0, 0, 440, 445},
		{255, 0, 0, 0, 255, 0, 360, 362},
		{100, 100, 100, 110, 110, 110, 17, 18},
	}
	for _, tc := range cases {
		d := colorDistance(tc.r1, tc.g1, tc.b1, tc.r2, tc.g2, tc.b2)
		if d < tc.wantMin || d > tc.wantMax {
			t.Errorf("colorDistance(%d,%d,%d - %d,%d,%d) = %.2f, want [%.1f, %.1f]",
				tc.r1, tc.g1, tc.b1, tc.r2, tc.g2, tc.b2, d, tc.wantMin, tc.wantMax)
		}
	}
}

func TestEdgeWeight(t *testing.T) {
	edge := edgeWeight(0, 50, 100, 100)
	if edge < 0.9 {
		t.Errorf("edge at x=0 should be ~1.0, got %f", edge)
	}

	center := edgeWeight(50, 50, 100, 100)
	if center > 0.01 {
		t.Errorf("center should be ~0, got %f", center)
	}

	nearEdge := edgeWeight(5, 50, 100, 100)
	if nearEdge < 0.3 || nearEdge > 0.8 {
		t.Errorf("near edge weight unexpected: %f", nearEdge)
	}
}

func TestEdgeWeightZeroDimensions(t *testing.T) {
	w := edgeWeight(0, 0, 0, 0)
	if w != 0 {
		t.Errorf("zero dimensions should return 0, got %f", w)
	}
}

func TestRemoveBackgroundStream(t *testing.T) {
	bg := color.NRGBA{R: 200, G: 200, B: 200, A: 255}
	fg := color.NRGBA{R: 30, G: 30, B: 30, A: 255}
	src := makeTestImage(80, 80, bg, fg)

	var input bytes.Buffer
	if err := png.Encode(&input, src); err != nil {
		t.Fatalf("encode: %v", err)
	}

	var output bytes.Buffer
	if err := RemoveBackgroundStream(&input, &output); err != nil {
		t.Fatalf("RemoveBackgroundStream: %v", err)
	}

	result, err := png.Decode(&output)
	if err != nil {
		t.Fatalf("decode result: %v", err)
	}

	bounds := result.Bounds()
	if bounds.Dx() != 80 || bounds.Dy() != 80 {
		t.Errorf("expected 80x80, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestRemoveBackgroundStreamInvalidInput(t *testing.T) {
	input := bytes.NewReader([]byte("not an image"))
	var output bytes.Buffer
	if err := RemoveBackgroundStream(input, &output); err == nil {
		t.Error("expected error for invalid input")
	}
}

func TestRemoveBackgroundFile(t *testing.T) {
	dir := t.TempDir()
	srcPath := filepath.Join(dir, "test.png")
	dstPath := filepath.Join(dir, "result.png")

	bg := color.NRGBA{R: 200, G: 180, B: 160, A: 255}
	fg := color.NRGBA{R: 30, G: 30, B: 30, A: 255}
	src := makeTestImage(60, 60, bg, fg)

	f, err := os.Create(srcPath)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := png.Encode(f, src); err != nil {
		t.Fatalf("encode: %v", err)
	}
	_ = f.Close()

	if err := RemoveBackground(srcPath, dstPath); err != nil {
		t.Fatalf("RemoveBackground: %v", err)
	}

	resultFile, err := os.Open(dstPath)
	if err != nil {
		t.Fatalf("open result: %v", err)
	}
	defer func() { _ = resultFile.Close() }()

	result, err := png.Decode(resultFile)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	if result.Bounds().Dx() != 60 {
		t.Errorf("expected width 60, got %d", result.Bounds().Dx())
	}
}

func TestRemoveBackgroundFileMissing(t *testing.T) {
	if err := RemoveBackground("/nonexistent/path.png", "/tmp/out.png"); err == nil {
		t.Error("expected error for missing source file")
	}
}

func TestRemoveBackgroundJPEGInput(t *testing.T) {
	bg := color.NRGBA{R: 200, G: 200, B: 200, A: 255}
	fg := color.NRGBA{R: 30, G: 30, B: 30, A: 255}
	src := makeTestImage(50, 50, bg, fg)

	var input bytes.Buffer
	if err := jpeg.Encode(&input, src, nil); err != nil {
		t.Fatalf("encode jpeg: %v", err)
	}

	var output bytes.Buffer
	if err := RemoveBackgroundStream(&input, &output); err != nil {
		t.Fatalf("RemoveBackgroundStream JPEG: %v", err)
	}

	if output.Len() == 0 {
		t.Error("expected non-empty output")
	}
}
