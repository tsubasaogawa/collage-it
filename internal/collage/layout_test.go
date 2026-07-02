package collage

import (
	"image/color"
	"strings"
	"testing"
)

func TestLayoutConfigCellSize(t *testing.T) {
	tests := []struct {
		name           string
		config         LayoutConfig
		wantCellSize   int
		wantErr        bool
		wantErrPattern string
	}{
		{
			name: "basic layout 2048x16",
			config: LayoutConfig{
				ArtifactSizePx: 2048,
				SpacingPx:      16,
			},
			wantCellSize: 672, // (2048 - 2*16) / 3 = 2016 / 3 = 672
			wantErr:      false,
		},
		{
			name: "no spacing",
			config: LayoutConfig{
				ArtifactSizePx: 600,
				SpacingPx:      0,
			},
			wantCellSize: 200, // 600 / 3 = 200
			wantErr:      false,
		},
		{
			name: "spacing too large",
			config: LayoutConfig{
				ArtifactSizePx: 100,
				SpacingPx:      100,
			},
			wantErr:        true,
			wantErrPattern: "artifact-size-px",
		},
		{
			name: "artifact size too small",
			config: LayoutConfig{
				ArtifactSizePx: 3,
				SpacingPx:      2,
			},
			wantErr:        true,
			wantErrPattern: "artifact-size-px",
		},
		{
			name: "negative artifact size",
			config: LayoutConfig{
				ArtifactSizePx: -100,
				SpacingPx:      10,
			},
			wantErr:        true,
			wantErrPattern: "artifact-size-px must be positive",
		},
		{
			name: "negative spacing",
			config: LayoutConfig{
				ArtifactSizePx: 600,
				SpacingPx:      -10,
			},
			wantErr:        true,
			wantErrPattern: "spacing-px must be non-negative",
		},
		{
			name: "minimum size with spacing 0",
			config: LayoutConfig{
				ArtifactSizePx: 3, // 3 cells of 1px + 0 spacing
				SpacingPx:      0,
			},
			wantCellSize: 1,
			wantErr:      false,
		},
		{
			name: "just below minimum with spacing 0",
			config: LayoutConfig{
				ArtifactSizePx: 2,
				SpacingPx:      0,
			},
			wantErr:        true,
			wantErrPattern: "minimum required",
		},
		{
			name: "error message includes minimum requirement",
			config: LayoutConfig{
				ArtifactSizePx: 10,
				SpacingPx:      5, // min required: 3 + 2*5 = 13
			},
			wantErr:        true,
			wantErrPattern: "minimum required is 13",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.config.CellSize()
			if (err != nil) != tt.wantErr {
				t.Errorf("CellSize() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.wantErrPattern != "" {
				if errorMsg := err.Error(); !contains(errorMsg, tt.wantErrPattern) {
					t.Errorf("CellSize() error message = %q, want to contain %q", errorMsg, tt.wantErrPattern)
				}
			}
			if !tt.wantErr && got != tt.wantCellSize {
				t.Errorf("CellSize() = %d, want %d", got, tt.wantCellSize)
			}
		})
	}
}

func TestLayoutConfigCellPlacement(t *testing.T) {
	config := LayoutConfig{
		ArtifactSizePx: 2048,
		SpacingPx:      16,
	}
	// cellSize = (2048 - 32) / 3 = 672

	tests := []struct {
		name    string
		row     int
		col     int
		wantX   int
		wantY   int
		wantErr bool
	}{
		{
			name:    "cell (0,0)",
			row:     0,
			col:     0,
			wantX:   0,   // 0 * (672 + 16)
			wantY:   0,   // 0 * (672 + 16)
			wantErr: false,
		},
		{
			name:    "cell (0,1)",
			row:     0,
			col:     1,
			wantX:   688, // 1 * (672 + 16)
			wantY:   0,   // 0 * (672 + 16)
			wantErr: false,
		},
		{
			name:    "cell (1,1)",
			row:     1,
			col:     1,
			wantX:   688, // 1 * (672 + 16)
			wantY:   688, // 1 * (672 + 16)
			wantErr: false,
		},
		{
			name:    "cell (2,2)",
			row:     2,
			col:     2,
			wantX:   1376, // 2 * (672 + 16)
			wantY:   1376, // 2 * (672 + 16)
			wantErr: false,
		},
		{
			name:    "invalid row",
			row:     3,
			col:     0,
			wantErr: true,
		},
		{
			name:    "invalid col",
			row:     0,
			col:     -1,
			wantErr: true,
		},
		{
			name:    "row too large",
			row:     3,
			col:     0,
			wantErr: true,
		},
		{
			name:    "col too large",
			row:     0,
			col:     3,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x, y, err := config.CellPlacement(tt.row, tt.col)
			if (err != nil) != tt.wantErr {
				t.Errorf("CellPlacement() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if x != tt.wantX || y != tt.wantY {
					t.Errorf("CellPlacement() = (%d,%d), want (%d,%d)", x, y, tt.wantX, tt.wantY)
				}
			}
		})
	}
}

func TestLayoutConfigCellPlacementRejectsInvalidConfig(t *testing.T) {
	config := LayoutConfig{ArtifactSizePx: 0, SpacingPx: 0}

	_, _, err := config.CellPlacement(0, 0)
	if err == nil {
		t.Fatal("CellPlacement() error = nil, want non-nil for invalid LayoutConfig")
	}
	if got := err.Error(); !strings.Contains(got, "artifact-size-px must be positive") {
		t.Fatalf("error = %q, want to mention artifact-size-px must be positive", got)
	}
}

func TestLayoutMinimumSize(t *testing.T) {
	// Minimum viable: 3 cells of 1px each + 2 gaps of 0px spacing = 3px
	tests := []struct {
		name       string
		size       int
		spacing    int
		shouldWork bool
	}{
		{
			name:       "minimum with no spacing",
			size:       3,
			spacing:    0,
			shouldWork: true,
		},
		{
			name:       "below minimum with spacing",
			size:       2,
			spacing:    0,
			shouldWork: false,
		},
		{
			name:       "too small for spacing",
			size:       10,
			spacing:    5,
			shouldWork: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := LayoutConfig{
				ArtifactSizePx: tt.size,
				SpacingPx:      tt.spacing,
			}
			_, err := config.CellSize()
			if (err == nil) != tt.shouldWork {
				t.Errorf("CellSize() error = %v, shouldWork %v", err, tt.shouldWork)
			}
		})
	}
}

func TestMinRequiredSize(t *testing.T) {
	tests := []struct {
		name          string
		spacing       int
		wantMinSize   int
	}{
		{
			name:          "no spacing",
			spacing:       0,
			wantMinSize:   3, // 3 cells + 2*0
		},
		{
			name:          "small spacing",
			spacing:       1,
			wantMinSize:   5, // 3 + 2*1
		},
		{
			name:          "medium spacing",
			spacing:       5,
			wantMinSize:   13, // 3 + 2*5
		},
		{
			name:          "large spacing",
			spacing:       100,
			wantMinSize:   203, // 3 + 2*100
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := LayoutConfig{
				ArtifactSizePx: tt.wantMinSize,
				SpacingPx:      tt.spacing,
			}
			minSize := config.minRequiredSize()
			if minSize != tt.wantMinSize {
				t.Errorf("minRequiredSize() = %d, want %d", minSize, tt.wantMinSize)
			}

			// Verify that exactly this size works (cell size = 1)
			_, err := config.CellSize()
			if err != nil {
				t.Errorf("CellSize() at minimum size should work, got error: %v", err)
			}

			// Verify that size - 1 doesn't work
			configTooSmall := LayoutConfig{
				ArtifactSizePx: tt.wantMinSize - 1,
				SpacingPx:      tt.spacing,
			}
			_, err = configTooSmall.CellSize()
			if err == nil {
				t.Errorf("CellSize() below minimum should fail but didn't")
			}
		})
	}
}

func TestNewCanvas(t *testing.T) {
	tests := []struct {
		name              string
		cfg               CanvasConfig
		wantWidth         int
		wantHeight        int
		wantBackgroundPx  color.RGBA
		wantErr           bool
		wantErrPattern    string
	}{
		{
			name: "white canvas 256x256",
			cfg: CanvasConfig{
				Size:            256,
				BackgroundColor: DefaultBackgroundColor(),
			},
			wantWidth:        256,
			wantHeight:       256,
			wantBackgroundPx: color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
			wantErr:          false,
		},
		{
			name: "black canvas 512x512",
			cfg: CanvasConfig{
				Size:            512,
				BackgroundColor: color.RGBA{R: 0, G: 0, B: 0, A: 0xff},
			},
			wantWidth:        512,
			wantHeight:       512,
			wantBackgroundPx: color.RGBA{R: 0, G: 0, B: 0, A: 0xff},
			wantErr:          false,
		},
		{
			name: "invalid: zero size",
			cfg: CanvasConfig{
				Size:            0,
				BackgroundColor: DefaultBackgroundColor(),
			},
			wantErr:        true,
			wantErrPattern: "size must be positive",
		},
		{
			name: "invalid: negative size",
			cfg: CanvasConfig{
				Size:            -100,
				BackgroundColor: DefaultBackgroundColor(),
			},
			wantErr:        true,
			wantErrPattern: "size must be positive",
		},
		{
			name: "invalid: nil color",
			cfg: CanvasConfig{
				Size:            256,
				BackgroundColor: nil,
			},
			wantErr:        true,
			wantErrPattern: "background color must not be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canvas, err := NewCanvas(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCanvas() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.wantErrPattern != "" {
				if errorMsg := err.Error(); !contains(errorMsg, tt.wantErrPattern) {
					t.Errorf("NewCanvas() error message = %q, want to contain %q", errorMsg, tt.wantErrPattern)
				}
			}
			if !tt.wantErr {
				if canvas == nil {
					t.Errorf("NewCanvas() returned nil canvas")
					return
				}

				// Check dimensions
				bounds := canvas.Bounds()
				if bounds.Dx() != tt.wantWidth || bounds.Dy() != tt.wantHeight {
					t.Errorf("Canvas size = %dx%d, want %dx%d", bounds.Dx(), bounds.Dy(), tt.wantWidth, tt.wantHeight)
				}

				// Check that canvas is filled with background color (sample corners and center)
				checkPixel := func(x, y int) {
					gotColor := color.RGBAModel.Convert(canvas.At(x, y)).(color.RGBA)
					if gotColor != tt.wantBackgroundPx {
						t.Errorf("Pixel at (%d,%d) = %v, want %v", x, y, gotColor, tt.wantBackgroundPx)
					}
				}
				checkPixel(0, 0)                              // top-left
				checkPixel(tt.wantWidth - 1, tt.wantHeight - 1) // bottom-right
				checkPixel(tt.wantWidth / 2, tt.wantHeight / 2) // center
			}
		})
	}
}

func TestCanvasAlwaysSquare(t *testing.T) {
	sizes := []int{1, 10, 256, 1024, 2048}
	for _, size := range sizes {
		t.Run("size="+string(rune(size)), func(t *testing.T) {
			cfg := CanvasConfig{
				Size:            size,
				BackgroundColor: DefaultBackgroundColor(),
			}
			canvas, err := NewCanvas(cfg)
			if err != nil {
				t.Errorf("NewCanvas() returned error: %v", err)
				return
			}
			bounds := canvas.Bounds()
			if bounds.Dx() != size || bounds.Dy() != size {
				t.Errorf("Canvas for size %d is %dx%d, want %dx%d", size, bounds.Dx(), bounds.Dy(), size, size)
			}
		})
	}
}

func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestCanvasConfigValidate(t *testing.T) {
	tests := []struct {
		name           string
		cfg            CanvasConfig
		wantErr        bool
		wantErrPattern string
	}{
		{
			name: "valid config",
			cfg: CanvasConfig{
				Size:            256,
				BackgroundColor: DefaultBackgroundColor(),
			},
			wantErr: false,
		},
		{
			name: "zero size",
			cfg: CanvasConfig{
				Size:            0,
				BackgroundColor: DefaultBackgroundColor(),
			},
			wantErr:        true,
			wantErrPattern: "size must be positive",
		},
		{
			name: "negative size",
			cfg: CanvasConfig{
				Size:            -1,
				BackgroundColor: DefaultBackgroundColor(),
			},
			wantErr:        true,
			wantErrPattern: "size must be positive",
		},
		{
			name: "nil background color",
			cfg: CanvasConfig{
				Size:            256,
				BackgroundColor: nil,
			},
			wantErr:        true,
			wantErrPattern: "background color must not be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.wantErrPattern != "" {
				if errorMsg := err.Error(); !contains(errorMsg, tt.wantErrPattern) {
					t.Errorf("Validate() error message = %q, want to contain %q", errorMsg, tt.wantErrPattern)
				}
			}
		})
	}
}

func TestNewCanvasFromLayout(t *testing.T) {
	tests := []struct {
		name        string
		layout      LayoutConfig
		bgColor     color.Color
		wantSize    int
		wantErr     bool
		wantErrMsg  string
	}{
		{
			name: "valid layout with white background",
			layout: LayoutConfig{
				ArtifactSizePx: 2048,
				SpacingPx:      16,
			},
			bgColor:  DefaultBackgroundColor(),
			wantSize: 2048,
			wantErr:  false,
		},
		{
			name: "valid layout with black background",
			layout: LayoutConfig{
				ArtifactSizePx: 1024,
				SpacingPx:      8,
			},
			bgColor:  color.RGBA{R: 0, G: 0, B: 0, A: 0xff},
			wantSize: 1024,
			wantErr:  false,
		},
		{
			name: "invalid layout - negative size",
			layout: LayoutConfig{
				ArtifactSizePx: -1,
				SpacingPx:      10,
			},
			bgColor:    DefaultBackgroundColor(),
			wantErr:    true,
			wantErrMsg: "artifact-size-px must be positive",
		},
		{
			name: "invalid layout - negative spacing",
			layout: LayoutConfig{
				ArtifactSizePx: 1024,
				SpacingPx:      -5,
			},
			bgColor:    DefaultBackgroundColor(),
			wantErr:    true,
			wantErrMsg: "spacing-px must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canvas, err := NewCanvasFromLayout(tt.layout, tt.bgColor)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCanvasFromLayout() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.wantErrMsg != "" {
				if errorMsg := err.Error(); !contains(errorMsg, tt.wantErrMsg) {
					t.Errorf("NewCanvasFromLayout() error message = %q, want to contain %q", errorMsg, tt.wantErrMsg)
				}
			}
			if !tt.wantErr {
				if canvas == nil {
					t.Errorf("NewCanvasFromLayout() returned nil canvas")
					return
				}
				bounds := canvas.Bounds()
				if bounds.Dx() != tt.wantSize || bounds.Dy() != tt.wantSize {
					t.Errorf("Canvas size = %dx%d, want %dx%d", bounds.Dx(), bounds.Dy(), tt.wantSize, tt.wantSize)
				}
			}
		})
	}
}
