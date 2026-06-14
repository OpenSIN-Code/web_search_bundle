// SPDX-License-Identifier: MIT
// Purpose: Resize extracted video frames using pure Go image processing.
// Docs: internal/imaging/resize.doc.md
package imaging

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"

	"golang.org/x/image/draw"
)

// ResizeOptions configures frame resizing.
type ResizeOptions struct {
	MaxWidth    int
	JPEGQuality int
}

// DefaultResizeOptions returns sensible defaults.
func DefaultResizeOptions() ResizeOptions {
	return ResizeOptions{MaxWidth: 1024, JPEGQuality: 75}
}

// ResizeImage resizes an image to fit within MaxWidth while preserving aspect ratio.
func ResizeImage(inputPath, outputPath string, opts ResizeOptions) error {
	in, err := os.Open(inputPath) // #nosec G304 — caller chooses image path
	if err != nil {
		return err
	}
	defer in.Close()

	src, format, err := image.Decode(in)
	if err != nil {
		return fmt.Errorf("decode %s: %w", inputPath, err)
	}

	bounds := src.Bounds()
	origWidth := bounds.Dx()
	origHeight := bounds.Dy()

	newWidth := origWidth
	newHeight := origHeight
	if opts.MaxWidth > 0 && origWidth > opts.MaxWidth {
		newWidth = opts.MaxWidth
		newHeight = int(float64(origHeight) * float64(newWidth) / float64(origWidth))
	}

	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, bounds, draw.Over, nil)

	out, err := os.Create(outputPath) // #nosec G304 — caller chooses output path
	if err != nil {
		return err
	}
	defer out.Close()

	switch format {
	case "png":
		return png.Encode(out, dst)
	default:
		return jpeg.Encode(out, dst, &jpeg.Options{Quality: opts.JPEGQuality})
	}
}

// ResizeBatch resizes a batch of images into a directory.
func ResizeBatch(inputPaths []string, outputDir string, opts ResizeOptions) ([]string, error) {
	if err := os.MkdirAll(outputDir, 0750); err != nil {
		return nil, fmt.Errorf("create output dir: %w", err)
	}
	var outputs []string
	for _, path := range inputPaths {
		name := filepath.Base(path)
		outPath := filepath.Join(outputDir, name)
		if err := ResizeImage(path, outPath, opts); err != nil {
			return nil, err
		}
		outputs = append(outputs, outPath)
	}
	return outputs, nil
}
