package main

import (
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"os"

	"collage-it/internal/collage"
)

const (
	defaultOutputFileName = "collage.jpg"
	defaultJPEGQuality    = 95
)

func run(opts options, dir, outputPath string) error {
	images, err := findInputImages(dir, opts.namePrefix)
	if err != nil {
		return err
	}

	if err := validateInputImageCount(images); err != nil {
		return err
	}

	preparedImages, err := loadAndPrepareImages(images)
	if err != nil {
		return err
	}

	layout := collage.LayoutConfig{
		ArtifactSizePx: opts.artifactSizePx,
		SpacingPx:      opts.spacingPx,
	}

	canvas, err := collage.NewCanvasFromLayout(layout, collage.DefaultBackgroundColor())
	if err != nil {
		return err
	}

	if err := collage.DrawCollage(canvas, preparedImages, layout); err != nil {
		return err
	}

	return saveJPEG(outputPath, canvas)
}

func loadAndPrepareImages(paths []string) ([]image.Image, error) {
	prepared := make([]image.Image, 0, len(paths))
	for _, path := range paths {
		src, err := loadImage(path)
		if err != nil {
			return nil, err
		}
		prepared = append(prepared, centerCropSquare(src))
	}
	return prepared, nil
}

func loadImage(path string) (img image.Image, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer func() {
		if cerr := f.Close(); err == nil && cerr != nil {
			err = fmt.Errorf("close %s: %w", path, cerr)
		}
	}()

	img, _, err = image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}

	return img, nil
}

func saveJPEG(path string, img image.Image) (err error) {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create output %s: %w", path, err)
	}
	defer func() {
		if cerr := f.Close(); err == nil && cerr != nil {
			err = fmt.Errorf("close output %s: %w", path, cerr)
		}
	}()

	if err := jpeg.Encode(f, img, &jpeg.Options{Quality: defaultJPEGQuality}); err != nil {
		return fmt.Errorf("encode output %s: %w", path, err)
	}

	return nil
}
