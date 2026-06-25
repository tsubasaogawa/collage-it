package collage

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
)

// CanvasConfig specifies how to create a canvas for a collage.
type CanvasConfig struct {
	// Width and height in pixels (must be equal for a square canvas).
	Size int
	// BackgroundColor is the fill color for the canvas.
	BackgroundColor color.Color
}

// Validate checks that the CanvasConfig is valid.
func (c CanvasConfig) Validate() error {
	if c.Size <= 0 {
		return fmt.Errorf("size must be positive, got %d", c.Size)
	}
	if c.BackgroundColor == nil {
		return fmt.Errorf("background color must not be nil")
	}
	return nil
}

// NewCanvas creates a new square RGBA canvas with the specified size and background color.
// The canvas size is guaranteed to be exactly CanvasConfig.Size x CanvasConfig.Size.
// Returns an error if the configuration is invalid.
func NewCanvas(cfg CanvasConfig) (*image.RGBA, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	canvas := image.NewRGBA(image.Rect(0, 0, cfg.Size, cfg.Size))

	// Fill the entire canvas with the background color.
	draw.Draw(
		canvas,
		canvas.Bounds(),
		&image.Uniform{C: cfg.BackgroundColor},
		image.Point{},
		draw.Src,
	)

	return canvas, nil
}

// NewCanvasFromLayout creates a canvas that matches the artifact size from a LayoutConfig.
// This is a convenience function that bridges layout and canvas creation.
func NewCanvasFromLayout(layout LayoutConfig, bgColor color.Color) (*image.RGBA, error) {
	if err := layout.validate(); err != nil {
		return nil, err
	}
	cfg := CanvasConfig{
		Size:            layout.ArtifactSizePx,
		BackgroundColor: bgColor,
	}
	return NewCanvas(cfg)
}

// DefaultBackgroundColor returns a sensible default background color (white).
func DefaultBackgroundColor() color.Color {
	return color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
}
