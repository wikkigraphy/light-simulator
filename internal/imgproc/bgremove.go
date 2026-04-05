package imgproc

import (
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"io"
	"math"
	"os"
)

// RemoveBackground reads an image from src, removes the background using
// edge-aware chroma keying, and writes a transparent PNG to dst.
func RemoveBackground(srcPath, dstPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer func() { _ = srcFile.Close() }()

	img, _, err := image.Decode(srcFile)
	if err != nil {
		return err
	}

	result := removeBackgroundFromImage(img)

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer func() { _ = dstFile.Close() }()

	return png.Encode(dstFile, result)
}

// RemoveBackgroundStream reads from r, removes background, writes PNG to w.
func RemoveBackgroundStream(r io.Reader, w io.Writer) error {
	img, _, err := image.Decode(r)
	if err != nil {
		return err
	}

	result := removeBackgroundFromImage(img)
	return png.Encode(w, result)
}

func removeBackgroundFromImage(img image.Image) *image.NRGBA {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	result := image.NewNRGBA(bounds)

	bgColor := sampleBackgroundColor(img)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			pr := uint8(r >> 8)
			pg := uint8(g >> 8)
			pb := uint8(b >> 8)

			dist := colorDistance(pr, pg, pb, bgColor.R, bgColor.G, bgColor.B)

			edgeFactor := edgeWeight(x-bounds.Min.X, y-bounds.Min.Y, width, height)

			threshold := 35.0 + edgeFactor*25.0
			softRange := 20.0

			var alpha uint8
			if dist < threshold {
				alpha = 0
			} else if dist < threshold+softRange {
				alpha = uint8(((dist - threshold) / softRange) * 255)
			} else {
				alpha = 255
			}

			result.SetNRGBA(x, y, color.NRGBA{R: pr, G: pg, B: pb, A: alpha})
		}
	}

	return result
}

type bgSample struct {
	R, G, B uint8
}

// sampleBackgroundColor samples the corners and edges to determine the
// dominant background color.
func sampleBackgroundColor(img image.Image) bgSample {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	sampleSize := 10
	if sampleSize > w/4 {
		sampleSize = w / 4
	}
	if sampleSize > h/4 {
		sampleSize = h / 4
	}
	if sampleSize < 1 {
		sampleSize = 1
	}

	var totalR, totalG, totalB float64
	var count float64

	for dy := 0; dy < sampleSize; dy++ {
		for dx := 0; dx < sampleSize; dx++ {
			corners := []image.Point{
				{X: bounds.Min.X + dx, Y: bounds.Min.Y + dy},
				{X: bounds.Max.X - 1 - dx, Y: bounds.Min.Y + dy},
				{X: bounds.Min.X + dx, Y: bounds.Max.Y - 1 - dy},
				{X: bounds.Max.X - 1 - dx, Y: bounds.Max.Y - 1 - dy},
			}
			for _, pt := range corners {
				r, g, b, _ := img.At(pt.X, pt.Y).RGBA()
				totalR += float64(r >> 8)
				totalG += float64(g >> 8)
				totalB += float64(b >> 8)
				count++
			}
		}
	}

	if count == 0 {
		return bgSample{128, 128, 128}
	}

	return bgSample{
		R: uint8(totalR / count),
		G: uint8(totalG / count),
		B: uint8(totalB / count),
	}
}

func colorDistance(r1, g1, b1, r2, g2, b2 uint8) float64 {
	dr := float64(r1) - float64(r2)
	dg := float64(g1) - float64(g2)
	db := float64(b1) - float64(b2)
	return math.Sqrt(dr*dr + dg*dg + db*db)
}

// edgeWeight returns a value 0-1 based on proximity to image edges.
// Higher at edges, lower toward center.
func edgeWeight(x, y, width, height int) float64 {
	if width == 0 || height == 0 {
		return 0
	}
	fx := float64(x) / float64(width)
	fy := float64(y) / float64(height)

	distFromEdge := math.Min(
		math.Min(fx, 1-fx),
		math.Min(fy, 1-fy),
	)

	if distFromEdge > 0.15 {
		return 0
	}
	return 1 - (distFromEdge / 0.15)
}
