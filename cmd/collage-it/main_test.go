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
	"time"
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
	if opts.inputDirectory != defaultInputDirectory {
		t.Fatalf("inputDirectory = %q, want %q", opts.inputDirectory, defaultInputDirectory)
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
	if opts.inputDirectory != defaultInputDirectory {
		t.Fatalf("inputDirectory = %q, want %q", opts.inputDirectory, defaultInputDirectory)
	}
}

func TestParseOptionsAcceptsInputDirectory(t *testing.T) {
	opts, err := parseOptions([]string{"-name-prefix", "IMG_", "/photos/monthly-nine/"})
	if err != nil {
		t.Fatalf("parseOptions() error = %v", err)
	}

	if opts.namePrefix != "IMG_" {
		t.Fatalf("namePrefix = %q, want %q", opts.namePrefix, "IMG_")
	}
	if opts.inputDirectory != "/photos/monthly-nine/" {
		t.Fatalf("inputDirectory = %q, want %q", opts.inputDirectory, "/photos/monthly-nine/")
	}
}

func TestParseOptionsRejectsMultipleInputDirectories(t *testing.T) {
	_, err := parseOptions([]string{"/photos/one", "/photos/two"})
	if err == nil {
		t.Fatal("parseOptions() error = nil, want non-nil")
	}
	if got := err.Error(); !strings.Contains(got, "at most one input directory") {
		t.Fatalf("error = %q, want to mention input directory count", got)
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

func TestSelectLatestInputImagesUsesModificationTime(t *testing.T) {
	dir := t.TempDir()
	baseTime := time.Unix(1_000_000, 0)
	for i := 0; i < 10; i++ {
		path := filepath.Join(dir, fmt.Sprintf("trip_%02d.png", i+1))
		mustWriteFile(t, path)
		modTime := baseTime.Add(time.Duration(i) * time.Minute)
		if err := os.Chtimes(path, modTime, modTime); err != nil {
			t.Fatalf("os.Chtimes() error = %v", err)
		}
	}

	images, err := findInputImages(dir, "trip_")
	if err != nil {
		t.Fatalf("findInputImages() error = %v", err)
	}

	got, err := selectLatestInputImages(images)
	if err != nil {
		t.Fatalf("selectLatestInputImages() error = %v", err)
	}

	want := make([]string, 0, 9)
	for i := 1; i < 10; i++ {
		want = append(want, filepath.Join(dir, fmt.Sprintf("trip_%02d.png", i+1)))
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("selectLatestInputImages() = %v, want %v", got, want)
	}
}

func TestSelectLatestInputImagesBreaksEqualModificationTimesByFilename(t *testing.T) {
	dir := t.TempDir()
	modTime := time.Unix(1_000_000, 0)
	images := make([]string, 0, 10)
	for i := 0; i < 10; i++ {
		path := filepath.Join(dir, fmt.Sprintf("trip_%02d.png", i+1))
		mustWriteFile(t, path)
		if err := os.Chtimes(path, modTime, modTime); err != nil {
			t.Fatalf("os.Chtimes() error = %v", err)
		}
		images = append(images, path)
	}

	got, err := selectLatestInputImages(images)
	if err != nil {
		t.Fatalf("selectLatestInputImages() error = %v", err)
	}

	want := images[:9]
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("selectLatestInputImages() = %v, want %v", got, want)
	}
}

func TestSelectLatestInputImagesKeepsNine(t *testing.T) {
	images := []string{
		"image-01.png",
		"image-02.png",
		"image-03.png",
		"image-04.png",
		"image-05.png",
		"image-06.png",
		"image-07.png",
		"image-08.png",
		"image-09.png",
	}

	got, err := selectLatestInputImages(images)
	if err != nil {
		t.Fatalf("selectLatestInputImages() error = %v", err)
	}
	if !reflect.DeepEqual(got, images) {
		t.Fatalf("selectLatestInputImages() = %v, want %v", got, images)
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
		if got := err.Error(); !strings.Contains(got, "at least 9") {
			t.Fatalf("error = %q, want to mention at least 9", got)
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

func TestRunValidatesInputImageCount(t *testing.T) {
	t.Run("too few images", func(t *testing.T) {
		dir := t.TempDir()
		for i := 0; i < 8; i++ {
			name := filepath.Join(dir, fmt.Sprintf("trip_%02d.png", i+1))
			mustWritePNGImage(t, name, newTestImage(4, 4))
		}

		opts := options{namePrefix: "trip_", spacingPx: 1, artifactSizePx: 11}
		err := run(opts, dir, filepath.Join(dir, defaultOutputFileName))
		if err == nil {
			t.Fatal("run() error = nil, want non-nil for 8 images")
		}
		if got := err.Error(); !strings.Contains(got, "at least 9") {
			t.Fatalf("error = %q, want to mention at least 9", got)
		}
	})

	t.Run("more than nine images uses the latest nine", func(t *testing.T) {
		dir := t.TempDir()
		baseTime := time.Unix(1_000_000, 0)
		for i := 0; i < 10; i++ {
			name := filepath.Join(dir, fmt.Sprintf("trip_%02d.png", i+1))
			if i == 0 {
				// The oldest candidate must be ignored before image decoding.
				mustWriteFile(t, name)
			} else {
				mustWritePNGImage(t, name, newTestImage(4, 4))
			}
			modTime := baseTime.Add(time.Duration(i) * time.Minute)
			if err := os.Chtimes(name, modTime, modTime); err != nil {
				t.Fatalf("os.Chtimes() error = %v", err)
			}
		}

		opts := options{namePrefix: "trip_", spacingPx: 1, artifactSizePx: 11}
		if err := run(opts, dir, filepath.Join(dir, defaultOutputFileName)); err != nil {
			t.Fatalf("run() error = %v, want nil when more than 9 images are present", err)
		}
	})
}

func TestRunRejectsMissingInputDirectory(t *testing.T) {
	dir := t.TempDir()
	missingDir := filepath.Join(dir, "does-not-exist")

	opts := options{namePrefix: "trip_", spacingPx: 1, artifactSizePx: 11}
	err := run(opts, missingDir, filepath.Join(dir, defaultOutputFileName))
	if err == nil {
		t.Fatal("run() error = nil, want non-nil for missing input directory")
	}
}

func TestRunPropagatesDecodeError(t *testing.T) {
	dir := t.TempDir()
	for i := 0; i < 8; i++ {
		name := filepath.Join(dir, fmt.Sprintf("trip_%02d.png", i+1))
		mustWritePNGImage(t, name, newTestImage(4, 4))
	}
	// 9th "image" is actually invalid image data with a supported extension.
	invalidPath := filepath.Join(dir, "trip_09.png")
	mustWriteFile(t, invalidPath)

	opts := options{namePrefix: "trip_", spacingPx: 1, artifactSizePx: 11}
	err := run(opts, dir, filepath.Join(dir, defaultOutputFileName))
	if err == nil {
		t.Fatal("run() error = nil, want non-nil for invalid image data")
	}
	if got := err.Error(); !strings.Contains(got, "decode") || !strings.Contains(got, invalidPath) {
		t.Fatalf("error = %q, want to mention decode and %q", got, invalidPath)
	}
}

func TestRunSpacingPxAffectsGapColor(t *testing.T) {
	dir := t.TempDir()
	for i := 0; i < 9; i++ {
		name := filepath.Join(dir, fmt.Sprintf("trip_%02d.png", i+1))
		mustWritePNGImage(t, name, newTestImage(4, 4))
	}

	outputPath := filepath.Join(dir, defaultOutputFileName)
	opts := options{
		namePrefix:     "trip_",
		spacingPx:      3,
		artifactSizePx: 30,
	}
	if err := run(opts, dir, outputPath); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	f, err := os.Open(outputPath)
	if err != nil {
		t.Fatalf("os.Open() error = %v", err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		t.Fatalf("image.Decode() error = %v", err)
	}

	// cellSize = (30 - 2*3) / 3 = 8. The gap between cell 0 and cell 1 spans
	// x in [8, 11); it should render as the (white) background color rather
	// than photo content, regardless of spacing. Allow a small tolerance
	// since the output is JPEG-compressed (lossy).
	wantBackground := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	gap := color.RGBAModel.Convert(img.At(9, 0)).(color.RGBA)
	assertColorNear(t, gap, wantBackground, 8)
}

func TestRunArtifactSizePxDeterminesOutputSize(t *testing.T) {
	sizes := []int{9, 30, 99}

	for _, size := range sizes {
		size := size
		t.Run(fmt.Sprintf("size=%d", size), func(t *testing.T) {
			dir := t.TempDir()
			for i := 0; i < 9; i++ {
				name := filepath.Join(dir, fmt.Sprintf("trip_%02d.png", i+1))
				mustWritePNGImage(t, name, newTestImage(4, 4))
			}

			outputPath := filepath.Join(dir, defaultOutputFileName)
			opts := options{namePrefix: "trip_", spacingPx: 1, artifactSizePx: size}
			if err := run(opts, dir, outputPath); err != nil {
				t.Fatalf("run() error = %v", err)
			}

			f, err := os.Open(outputPath)
			if err != nil {
				t.Fatalf("os.Open() error = %v", err)
			}
			defer f.Close()

			img, _, err := image.Decode(f)
			if err != nil {
				t.Fatalf("image.Decode() error = %v", err)
			}
			if bounds := img.Bounds(); bounds.Dx() != size || bounds.Dy() != size {
				t.Fatalf("bounds = %v, want %dx%d", bounds, size, size)
			}
		})
	}
}

func TestLoadImageRejectsMissingFile(t *testing.T) {
	dir := t.TempDir()
	missingPath := filepath.Join(dir, "missing.png")

	_, err := loadImage(missingPath)
	if err == nil {
		t.Fatal("loadImage() error = nil, want non-nil for missing file")
	}
	if got := err.Error(); !strings.Contains(got, "open") || !strings.Contains(got, missingPath) {
		t.Fatalf("error = %q, want to mention open and %q", got, missingPath)
	}
}

func TestLoadImageRejectsInvalidImageData(t *testing.T) {
	dir := t.TempDir()
	invalidPath := filepath.Join(dir, "invalid.png")
	mustWriteFile(t, invalidPath)

	_, err := loadImage(invalidPath)
	if err == nil {
		t.Fatal("loadImage() error = nil, want non-nil for invalid image data")
	}
	if got := err.Error(); !strings.Contains(got, "decode") || !strings.Contains(got, invalidPath) {
		t.Fatalf("error = %q, want to mention decode and %q", got, invalidPath)
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

func assertColorNear(t *testing.T, got, want color.RGBA, tolerance int) {
	t.Helper()

	diff := func(a, b uint8) int {
		if int(a) > int(b) {
			return int(a) - int(b)
		}
		return int(b) - int(a)
	}
	if diff(got.R, want.R) > tolerance || diff(got.G, want.G) > tolerance || diff(got.B, want.B) > tolerance {
		t.Fatalf("color = %#v, want %#v within tolerance %d", got, want, tolerance)
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
