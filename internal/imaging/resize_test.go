// Purpose: Unit tests for image resizing helpers.
// Docs: internal/imaging/resize_test.doc.md
package imaging

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultResizeOptions(t *testing.T) {
	opts := DefaultResizeOptions()
	if opts.MaxWidth != 1024 {
		t.Errorf("MaxWidth = %d, want 1024", opts.MaxWidth)
	}
	if opts.JPEGQuality != 75 {
		t.Errorf("JPEGQuality = %d, want 75", opts.JPEGQuality)
	}
}

func TestResizeImagePreservesAspectRatio(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.png")
	out := filepath.Join(dir, "out.jpg")

	// Create a 2000x1000 red PNG.
	img := image.NewRGBA(image.Rect(0, 0, 2000, 1000))
	for y := 0; y < 1000; y++ {
		for x := 0; x < 2000; x++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}
	f, err := os.Create(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		t.Fatal(err)
	}
	f.Close()

	if err := ResizeImage(src, out, ResizeOptions{MaxWidth: 1024, JPEGQuality: 80}); err != nil {
		t.Fatal(err)
	}

	f, err = os.Open(out)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	decoded, _, err := image.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	bounds := decoded.Bounds()
	if bounds.Dx() != 1024 {
		t.Errorf("width = %d, want 1024", bounds.Dx())
	}
	// Height should be 512 because aspect ratio is preserved.
	if bounds.Dy() != 512 {
		t.Errorf("height = %d, want 512", bounds.Dy())
	}
}

func TestResizeImageKeepsSmallImage(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.png")
	out := filepath.Join(dir, "out.jpg")

	img := image.NewRGBA(image.Rect(0, 0, 400, 300))
	for y := 0; y < 300; y++ {
		for x := 0; x < 400; x++ {
			img.Set(x, y, color.RGBA{0, 255, 0, 255})
		}
	}
	f, err := os.Create(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		t.Fatal(err)
	}
	f.Close()

	if err := ResizeImage(src, out, ResizeOptions{MaxWidth: 1024}); err != nil {
		t.Fatal(err)
	}

	f, err = os.Open(out)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	decoded, _, err := image.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	bounds := decoded.Bounds()
	if bounds.Dx() != 400 {
		t.Errorf("width = %d, want 400", bounds.Dx())
	}
	if bounds.Dy() != 300 {
		t.Errorf("height = %d, want 300", bounds.Dy())
	}
}

func TestResizeImageMissingInput(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out.jpg")
	if err := ResizeImage(filepath.Join(t.TempDir(), "missing.png"), out, DefaultResizeOptions()); err == nil {
		t.Fatal("expected error for missing input")
	}
}

func TestResizeImagePNGOutput(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.png")
	out := filepath.Join(dir, "out.png")

	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	if err := func() error {
		f, err := os.Create(src)
		if err != nil {
			return err
		}
		defer f.Close()
		return png.Encode(f, img)
	}(); err != nil {
		t.Fatal(err)
	}

	if err := ResizeImage(src, out, DefaultResizeOptions()); err != nil {
		t.Fatal(err)
	}

	f, err := os.Open(out)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	_, format, err := image.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	if format != "png" {
		t.Errorf("format = %s, want png", format)
	}
}

func TestResizeImageJPEGInput(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.jpg")
	out := filepath.Join(dir, "out.jpg")

	img := image.NewRGBA(image.Rect(0, 0, 800, 600))
	if err := func() error {
		f, err := os.Create(src)
		if err != nil {
			return err
		}
		defer f.Close()
		return jpeg.Encode(f, img, &jpeg.Options{Quality: 80})
	}(); err != nil {
		t.Fatal(err)
	}

	if err := ResizeImage(src, out, ResizeOptions{MaxWidth: 400, JPEGQuality: 70}); err != nil {
		t.Fatal(err)
	}
	outFile, err := os.Open(out)
	if err != nil {
		t.Fatal(err)
	}
	defer outFile.Close()
	decoded, _, err := image.Decode(outFile)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Bounds().Dx() != 400 {
		t.Errorf("width = %d, want 400", decoded.Bounds().Dx())
	}
}

func TestResizeBatch(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "out")

	var paths []string
	for i := 0; i < 3; i++ {
		p := filepath.Join(dir, fmt.Sprintf("img%d.png", i))
		img := image.NewRGBA(image.Rect(0, 0, 600, 400))
		f, err := os.Create(p)
		if err != nil {
			t.Fatal(err)
		}
		if err := png.Encode(f, img); err != nil {
			f.Close()
			t.Fatal(err)
		}
		f.Close()
		paths = append(paths, p)
	}

	outputs, err := ResizeBatch(paths, outDir, ResizeOptions{MaxWidth: 300})
	if err != nil {
		t.Fatal(err)
	}
	if len(outputs) != 3 {
		t.Fatalf("got %d outputs, want 3", len(outputs))
	}
	for _, out := range outputs {
		if _, err := os.Stat(out); err != nil {
			t.Errorf("output file missing: %v", err)
		}
	}
}

func TestResizeBatchMissingInput(t *testing.T) {
	_, err := ResizeBatch([]string{filepath.Join(t.TempDir(), "missing.png")}, t.TempDir(), DefaultResizeOptions())
	if err == nil {
		t.Fatal("expected error for missing input")
	}
}
