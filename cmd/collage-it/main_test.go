package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestParseOptionsDefaults(t *testing.T) {
	opts, err := parseOptions(nil)
	if err != nil {
		t.Fatalf("parseOptions() error = %v", err)
	}

	if opts.namePrefix != defaultNamePrefix {
		t.Fatalf("namePrefix = %q, want %q", opts.namePrefix, defaultNamePrefix)
	}
	if opts.spacingPx != defaultSpacingPx {
		t.Fatalf("spacingPx = %d, want %d", opts.spacingPx, defaultSpacingPx)
	}
	if opts.artifactSizePx != defaultArtifactSizePx {
		t.Fatalf("artifactSizePx = %d, want %d", opts.artifactSizePx, defaultArtifactSizePx)
	}
}

func TestParseOptionsCustomValues(t *testing.T) {
	opts, err := parseOptions([]string{"-name-prefix", "trip_", "-spacing-px", "24", "-artifact-size-px", "1024"})
	if err != nil {
		t.Fatalf("parseOptions() error = %v", err)
	}

	if opts.namePrefix != "trip_" {
		t.Fatalf("namePrefix = %q, want %q", opts.namePrefix, "trip_")
	}
	if opts.spacingPx != 24 {
		t.Fatalf("spacingPx = %d, want %d", opts.spacingPx, 24)
	}
	if opts.artifactSizePx != 1024 {
		t.Fatalf("artifactSizePx = %d, want %d", opts.artifactSizePx, 1024)
	}
}

func TestParseOptionsRejectsNegativeSpacing(t *testing.T) {
	_, err := parseOptions([]string{"-spacing-px", "-1"})
	if err == nil {
		t.Fatal("parseOptions() error = nil, want non-nil")
	}
	if got := err.Error(); !strings.Contains(got, "-spacing-px") {
		t.Fatalf("error = %q, want to mention -spacing-px", got)
	}
}

func TestParseOptionsRejectsNonPositiveArtifactSize(t *testing.T) {
	_, err := parseOptions([]string{"-artifact-size-px", "0"})
	if err == nil {
		t.Fatal("parseOptions() error = nil, want non-nil")
	}
	if got := err.Error(); !strings.Contains(got, "-artifact-size-px") {
		t.Fatalf("error = %q, want to mention -artifact-size-px", got)
	}
}

func TestFindInputImagesFiltersAndSorts(t *testing.T) {
	dir := t.TempDir()

	mustWriteFile(t, filepath.Join(dir, "trip_c.png"))
	mustWriteFile(t, filepath.Join(dir, "trip_a.jpg"))
	mustWriteFile(t, filepath.Join(dir, "trip_b.jpeg"))
	mustWriteFile(t, filepath.Join(dir, "trip_d.txt"))
	mustWriteFile(t, filepath.Join(dir, "other.jpg"))
	if err := os.Mkdir(filepath.Join(dir, "nested"), 0o755); err != nil {
		t.Fatalf("os.Mkdir() error = %v", err)
	}

	got, err := findInputImages(dir, "trip_")
	if err != nil {
		t.Fatalf("findInputImages() error = %v", err)
	}

	want := []string{
		filepath.Join(dir, "trip_a.jpg"),
		filepath.Join(dir, "trip_b.jpeg"),
		filepath.Join(dir, "trip_c.png"),
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("findInputImages() = %v, want %v", got, want)
	}
}

func TestValidateInputImageCount(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		images := make([]string, 9)
		if err := validateInputImageCount(images); err != nil {
			t.Fatalf("validateInputImageCount() error = %v", err)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		err := validateInputImageCount(make([]string, 8))
		if err == nil {
			t.Fatal("validateInputImageCount() error = nil, want non-nil")
		}
		if got := err.Error(); !strings.Contains(got, "exactly 9") {
			t.Fatalf("error = %q, want to mention exactly 9", got)
		}
	})
}

func TestCenterCropSquare(t *testing.T) {
	t.Run("landscape", func(t *testing.T) {
		src := newTestImage(5, 3)
		got := centerCropSquare(src)

		assertCropBounds(t, got, 3, 3)
		assertCropSourcePixel(t, got, 0, 0, 1, 0)
		assertCropSourcePixel(t, got, 2, 2, 3, 2)
	})

	t.Run("portrait", func(t *testing.T) {
		src := newTestImage(3, 5)
		got := centerCropSquare(src)

		assertCropBounds(t, got, 3, 3)
		assertCropSourcePixel(t, got, 0, 0, 0, 1)
		assertCropSourcePixel(t, got, 2, 2, 2, 3)
	})

	t.Run("square", func(t *testing.T) {
		src := newTestImage(4, 4)
		got := centerCropSquare(src)

		assertCropBounds(t, got, 4, 4)
		assertCropSourcePixel(t, got, 0, 0, 0, 0)
		assertCropSourcePixel(t, got, 3, 3, 3, 3)
	})

	t.Run("odd difference", func(t *testing.T) {
		src := newTestImage(6, 3)
		got := centerCropSquare(src)

		assertCropBounds(t, got, 3, 3)
		assertCropSourcePixel(t, got, 0, 0, 1, 0)
		assertCropSourcePixel(t, got, 2, 2, 3, 2)
	})
}

func TestRunGeneratesJPEGOutput(t *testing.T) {
	dir := t.TempDir()
	for i := 0; i < 9; i++ {
		name := filepath.Join(dir, fmt.Sprintf("trip_%02d.png", i+1))
		mustWritePNGImage(t, name, newTestImage(4+i%3, 5+i%4))
	}

	outputPath := filepath.Join(dir, defaultOutputFileName)
	opts := options{
		namePrefix:     "trip_",
		spacingPx:      1,
		artifactSizePx: 11,
	}
	if err := run(opts, dir, outputPath); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	f, err := os.Open(outputPath)
	if err != nil {
		t.Fatalf("os.Open() error = %v", err)
	}
	defer f.Close()

	img, format, err := image.Decode(f)
	if err != nil {
		t.Fatalf("image.Decode() error = %v", err)
	}
	if format != "jpeg" {
		t.Fatalf("format = %q, want jpeg", format)
	}
	if bounds := img.Bounds(); bounds.Dx() != opts.artifactSizePx || bounds.Dy() != opts.artifactSizePx {
		t.Fatalf("bounds = %v, want %dx%d", bounds, opts.artifactSizePx, opts.artifactSizePx)
	}
}

func TestSaveJPEGRejectsDirectoryPath(t *testing.T) {
	dir := t.TempDir()
	err := saveJPEG(dir, newTestImage(1, 1))
	if err == nil {
		t.Fatal("saveJPEG() error = nil, want non-nil")
	}
	if got := err.Error(); !strings.Contains(got, "create output") || !strings.Contains(got, dir) {
		t.Fatalf("error = %q, want to mention output path", got)
	}
}

func mustWriteFile(t *testing.T, path string) {
	t.Helper()

	if err := os.WriteFile(path, []byte("test"), 0o644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}
}

func mustWritePNGImage(t *testing.T, path string, img image.Image) {
	t.Helper()

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("os.Create() error = %v", err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		t.Fatalf("png.Encode() error = %v", err)
	}
}

func newTestImage(width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetRGBA(x, y, color.RGBA{
				R: uint8(x * 32),
				G: uint8(y * 32),
				B: uint8((x + y) * 16),
				A: 0xff,
			})
		}
	}
	return img
}

func assertCropBounds(t *testing.T, got image.Image, wantWidth, wantHeight int) {
	t.Helper()

	bounds := got.Bounds()
	if bounds.Dx() != wantWidth || bounds.Dy() != wantHeight {
		t.Fatalf("bounds = %v, want %dx%d", bounds, wantWidth, wantHeight)
	}
}

func assertCropSourcePixel(t *testing.T, got image.Image, gx, gy, sx, sy int) {
	t.Helper()

	want := color.RGBA{
		R: uint8(sx * 32),
		G: uint8(sy * 32),
		B: uint8((sx + sy) * 16),
		A: 0xff,
	}
	if gotColor := color.RGBAModel.Convert(got.At(gx, gy)).(color.RGBA); gotColor != want {
		t.Fatalf("pixel(%d,%d) = %#v, want %#v", gx, gy, gotColor, want)
	}
}
