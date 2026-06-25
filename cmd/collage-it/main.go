package main

import (
	"flag"
	"fmt"
	"io"
	"os"
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

func main() {
	if _, err := parseOptions(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
