package main

import (
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
