package main

import (
	"flag"
	"fmt"
	"image"
	"image/draw"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	defaultNamePrefix     = ""
	defaultSpacingPx      = 16
	defaultArtifactSizePx = 2048
)

type options struct {
	namePrefix     string
	spacingPx      int
	artifactSizePx int
}

var supportedImageExtensions = map[string]struct{}{
	".jpg":  {},
	".jpeg": {},
	".png":  {},
}

func parseOptions(args []string) (options, error) {
	fs := flag.NewFlagSet("collage-it", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var opts options
	fs.StringVar(&opts.namePrefix, "name-prefix", defaultNamePrefix, "prefix for input image filenames")
	fs.IntVar(&opts.spacingPx, "spacing-px", defaultSpacingPx, "spacing between photos in pixels")
	fs.IntVar(&opts.artifactSizePx, "artifact-size-px", defaultArtifactSizePx, "output image side length in pixels")

	if err := fs.Parse(args); err != nil {
		return options{}, err
	}

	if opts.spacingPx < 0 {
		return options{}, fmt.Errorf("invalid value for -spacing-px: must be >= 0")
	}
	if opts.artifactSizePx <= 0 {
		return options{}, fmt.Errorf("invalid value for -artifact-size-px: must be > 0")
	}

	return opts, nil
}

func findInputImages(dir, prefix string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var images []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if prefix != "" && !strings.HasPrefix(name, prefix) {
			continue
		}
		if !isSupportedImageFile(name) {
			continue
		}

		images = append(images, filepath.Join(dir, name))
	}

	sort.Strings(images)
	return images, nil
}

func isSupportedImageFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	_, ok := supportedImageExtensions[ext]
	return ok
}

func validateInputImageCount(images []string) error {
	if len(images) != 9 {
		return fmt.Errorf("found %d image files, want exactly 9", len(images))
	}
	return nil
}

func centerCropSquare(src image.Image) image.Image {
	bounds := src.Bounds()
	size := bounds.Dx()
	if bounds.Dy() < size {
		size = bounds.Dy()
	}

	x0 := bounds.Min.X + (bounds.Dx()-size)/2
	y0 := bounds.Min.Y + (bounds.Dy()-size)/2
	cropBounds := image.Rect(0, 0, size, size)
	dst := image.NewRGBA(cropBounds)
	draw.Draw(dst, cropBounds, src, image.Pt(x0, y0), draw.Src)
	return dst
}

func main() {
	opts, err := parseOptions(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	images, err := findInputImages(".", opts.namePrefix)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := validateInputImageCount(images); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
