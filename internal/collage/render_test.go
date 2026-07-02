package collage

import (
	"image"
	"image/color"
	"strings"
	"testing"
)

func TestResizeToSquareNearestNeighbor(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 2, 2))
	src.SetRGBA(0, 0, color.RGBA{R: 255, A: 255})
	src.SetRGBA(1, 0, color.RGBA{G: 255, A: 255})
	src.SetRGBA(0, 1, color.RGBA{B: 255, A: 255})
	src.SetRGBA(1, 1, color.RGBA{R: 255, G: 255, A: 255})

	got, err := ResizeToSquare(src, 4)
	if err != nil {
		t.Fatalf("ResizeToSquare() error = %v", err)
	}

	assertColor(t, got, 0, 0, color.RGBA{R: 255, A: 255})
	assertColor(t, got, 1, 1, color.RGBA{R: 255, A: 255})
	assertColor(t, got, 2, 0, color.RGBA{G: 255, A: 255})
	assertColor(t, got, 3, 3, color.RGBA{R: 255, G: 255, A: 255})
	assertColor(t, got, 0, 2, color.RGBA{B: 255, A: 255})
	assertColor(t, got, 2, 2, color.RGBA{R: 255, G: 255, A: 255})
}

func TestResizeToSquareRejectsInvalidInput(t *testing.T) {
	t.Run("non-positive size", func(t *testing.T) {
		src := image.NewRGBA(image.Rect(0, 0, 2, 2))
		if _, err := ResizeToSquare(src, 0); err == nil {
			t.Fatal("ResizeToSquare() error = nil, want non-nil for size <= 0")
		}
	})

	t.Run("nil source", func(t *testing.T) {
		if _, err := ResizeToSquare(nil, 4); err == nil {
			t.Fatal("ResizeToSquare() error = nil, want non-nil for nil source")
		}
	})

	t.Run("zero bounds source", func(t *testing.T) {
		src := image.NewRGBA(image.Rect(0, 0, 0, 0))
		if _, err := ResizeToSquare(src, 4); err == nil {
			t.Fatal("ResizeToSquare() error = nil, want non-nil for zero-bounds source")
		}
	})
}

func TestDrawCollage(t *testing.T) {
	layout := LayoutConfig{
		ArtifactSizePx: 11,
		SpacingPx:      1,
	}

	canvas, err := NewCanvas(CanvasConfig{
		Size:            11,
		BackgroundColor: color.RGBA{R: 255, G: 255, B: 255, A: 255},
	})
	if err != nil {
		t.Fatalf("NewCanvas() error = %v", err)
	}

	images := make([]image.Image, 0, 9)
	want := make([]color.RGBA, 0, 9)
	for i := 0; i < 9; i++ {
		c := color.RGBA{R: uint8(20 * i), G: uint8(255 - 20*i), B: uint8(10 * i), A: 255}
		images = append(images, solidImage(c))
		want = append(want, c)
	}

	if err := DrawCollage(canvas, images, layout); err != nil {
		t.Fatalf("DrawCollage() error = %v", err)
	}

	cellSize, err := layout.CellSize()
	if err != nil {
		t.Fatalf("CellSize() error = %v", err)
	}

	for idx, wantColor := range want {
		row := idx / 3
		col := idx % 3
		x := col * (cellSize + layout.SpacingPx)
		y := row * (cellSize + layout.SpacingPx)
		assertColor(t, canvas, x, y, wantColor)
		assertColor(t, canvas, x+cellSize-1, y+cellSize-1, wantColor)
	}

	assertColor(t, canvas, 3, 0, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	assertColor(t, canvas, 0, 3, color.RGBA{R: 255, G: 255, B: 255, A: 255})
}

func TestDrawCollageRejectsInvalidInput(t *testing.T) {
	layout := LayoutConfig{ArtifactSizePx: 11, SpacingPx: 1}
	canvas, err := NewCanvas(CanvasConfig{
		Size:            11,
		BackgroundColor: DefaultBackgroundColor(),
	})
	if err != nil {
		t.Fatalf("NewCanvas() error = %v", err)
	}

	if err := DrawCollage(nil, nil, layout); err == nil {
		t.Fatal("DrawCollage(nil, ...) error = nil, want non-nil")
	}

	if err := DrawCollage(canvas, []image.Image{}, layout); err == nil {
		t.Fatal("DrawCollage() with wrong image count error = nil, want non-nil")
	}
}

func TestDrawCollageRejectsCanvasSizeMismatch(t *testing.T) {
	layout := LayoutConfig{ArtifactSizePx: 11, SpacingPx: 1}
	canvas, err := NewCanvas(CanvasConfig{
		Size:            9, // mismatched with layout.ArtifactSizePx
		BackgroundColor: DefaultBackgroundColor(),
	})
	if err != nil {
		t.Fatalf("NewCanvas() error = %v", err)
	}

	images := make([]image.Image, 0, 9)
	for i := 0; i < 9; i++ {
		images = append(images, solidImage(color.RGBA{A: 255}))
	}

	err = DrawCollage(canvas, images, layout)
	if err == nil {
		t.Fatal("DrawCollage() error = nil, want non-nil for canvas size mismatch")
	}
	if got := err.Error(); !strings.Contains(got, "canvas size") {
		t.Fatalf("error = %q, want to mention canvas size", got)
	}
}

func solidImage(c color.RGBA) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.SetRGBA(0, 0, c)
	return img
}

func assertColor(t *testing.T, img image.Image, x, y int, want color.RGBA) {
	t.Helper()

	got := color.RGBAModel.Convert(img.At(x, y)).(color.RGBA)
	if got != want {
		t.Fatalf("pixel(%d,%d) = %#v, want %#v", x, y, got, want)
	}
}
