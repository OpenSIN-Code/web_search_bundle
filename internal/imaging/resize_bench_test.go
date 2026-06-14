// Purpose: Benchmark image resizing and batch resizing hot paths.
// Docs: resize.doc.md
package imaging

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func makeTestImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8((x * 255) / width),
				G: uint8((y * 255) / height),
				B: uint8(((x + y) * 255) / (width + height)),
				A: 255,
			})
		}
	}
	return img
}

func writeTestPNG(b *testing.B, img image.Image, name string) string {
	b.Helper()
	path := filepath.Join(b.TempDir(), name)
	f, err := os.Create(path)
	if err != nil {
		b.Fatalf("create test image: %v", err)
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		b.Fatalf("encode test image: %v", err)
	}
	if err := f.Close(); err != nil {
		b.Fatalf("close test image: %v", err)
	}
	return path
}

func BenchmarkResizeImage1024(b *testing.B) {
	inputPath := writeTestPNG(b, makeTestImage(1920, 1080), "input.png")
	opts := ResizeOptions{MaxWidth: 1024, JPEGQuality: 75}
	outDir := b.TempDir()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		outputPath := filepath.Join(outDir, fmt.Sprintf("out-%d.jpg", i))
		if err := ResizeImage(inputPath, outputPath, opts); err != nil {
			b.Fatalf("resize: %v", err)
		}
	}
}

func BenchmarkResizeImage512(b *testing.B) {
	inputPath := writeTestPNG(b, makeTestImage(1280, 720), "input.png")
	opts := ResizeOptions{MaxWidth: 512, JPEGQuality: 75}
	outDir := b.TempDir()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		outputPath := filepath.Join(outDir, fmt.Sprintf("out-%d.jpg", i))
		if err := ResizeImage(inputPath, outputPath, opts); err != nil {
			b.Fatalf("resize: %v", err)
		}
	}
}

func BenchmarkResizeBatch8(b *testing.B) {
	var inputPaths []string
	for i := 0; i < 8; i++ {
		inputPaths = append(inputPaths, writeTestPNG(b, makeTestImage(1280, 720), fmt.Sprintf("input-%d.png", i)))
	}
	opts := ResizeOptions{MaxWidth: 1024, JPEGQuality: 75}
	outDir := b.TempDir()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		batchDir := filepath.Join(outDir, fmt.Sprintf("batch-%d", i))
		if _, err := ResizeBatch(inputPaths, batchDir, opts); err != nil {
			b.Fatalf("resize batch: %v", err)
		}
	}
}
