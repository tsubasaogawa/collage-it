package main

import (
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

func mustWriteFile(t *testing.T, path string) {
	t.Helper()

	if err := os.WriteFile(path, []byte("test"), 0o644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}
}
