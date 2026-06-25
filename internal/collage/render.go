package collage

import (
	"fmt"
	"image"
	"image/draw"
)

const collageGridSize = 3

// ResizeToSquare scales src to a square image with the requested size.
func ResizeToSquare(src image.Image, size int) (*image.RGBA, error) {
	if size <= 0 {
		return nil, fmt.Errorf("size must be positive, got %d", size)
	}
	if src == nil {
		return nil, fmt.Errorf("source image must not be nil")
	}

	bounds := src.Bounds()
	if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
		return nil, fmt.Errorf("source image must have non-zero bounds")
	}

	dst := image.NewRGBA(image.Rect(0, 0, size, size))
	srcW := bounds.Dx()
	srcH := bounds.Dy()

	for y := 0; y < size; y++ {
		srcY := bounds.Min.Y + (y*srcH)/size
		if srcY >= bounds.Max.Y {
			srcY = bounds.Max.Y - 1
		}
		for x := 0; x < size; x++ {
			srcX := bounds.Min.X + (x*srcW)/size
			if srcX >= bounds.Max.X {
				srcX = bounds.Max.X - 1
			}
			dst.Set(x, y, src.At(srcX, srcY))
		}
	}

	return dst, nil
}

// DrawCollage resizes each source image to the grid cell size and draws them
// onto canvas in row-major order.
func DrawCollage(canvas draw.Image, images []image.Image, layout LayoutConfig) error {
	if canvas == nil {
		return fmt.Errorf("canvas must not be nil")
	}
	if len(images) != collageGridSize*collageGridSize {
		return fmt.Errorf("expected %d images, got %d", collageGridSize*collageGridSize, len(images))
	}

	if bounds := canvas.Bounds(); bounds.Dx() != layout.ArtifactSizePx || bounds.Dy() != layout.ArtifactSizePx {
		return fmt.Errorf("canvas size %dx%d does not match artifact-size-px %d", bounds.Dx(), bounds.Dy(), layout.ArtifactSizePx)
	}

	cellSize, err := layout.CellSize()
	if err != nil {
		return err
	}

	for idx, src := range images {
		resized, err := ResizeToSquare(src, cellSize)
		if err != nil {
			return fmt.Errorf("resize image %d: %w", idx, err)
		}

		row := idx / collageGridSize
		col := idx % collageGridSize
		x, y, err := layout.CellPlacement(row, col)
		if err != nil {
			return err
		}

		draw.Draw(canvas, image.Rect(x, y, x+cellSize, y+cellSize), resized, image.Point{}, draw.Over)
	}

	return nil
}
