package collage

import "fmt"

// LayoutConfig holds the configuration for a 3x3 collage layout.
type LayoutConfig struct {
	// ArtifactSizePx is the desired side length of the output square canvas in pixels.
	ArtifactSizePx int
	// SpacingPx is the number of pixels between photos.
	SpacingPx int
}

// CellSize computes the size of each individual cell in the 3x3 grid.
// Returns the cell width/height in pixels and any error if the layout is invalid.
func (c LayoutConfig) CellSize() (int, error) {
	if err := c.validate(); err != nil {
		return 0, err
	}

	// Total available space after removing spacing.
	// For a 3x3 grid: 2 gaps of spacing between 3 cells.
	spacingTotal := 2 * c.SpacingPx
	availableSize := c.ArtifactSizePx - spacingTotal
	if availableSize <= 0 {
		minRequired := c.minRequiredSize()
		return 0, fmt.Errorf("artifact-size-px (%d) is too small for spacing-px (%d); minimum required is %d pixels", c.ArtifactSizePx, c.SpacingPx, minRequired)
	}

	// Divide available space equally among 3 cells.
	cellSize := availableSize / 3
	if cellSize <= 0 {
		minRequired := c.minRequiredSize()
		return 0, fmt.Errorf("artifact-size-px (%d) is too small for spacing-px (%d); minimum required is %d pixels", c.ArtifactSizePx, c.SpacingPx, minRequired)
	}

	return cellSize, nil
}

// CellPlacement returns the x and y coordinates for the top-left corner of the cell
// at grid position (row, col) in the 3x3 grid (0-indexed).
func (c LayoutConfig) CellPlacement(row, col int) (x, y int, err error) {
	if err := c.validate(); err != nil {
		return 0, 0, err
	}

	if row < 0 || row > 2 || col < 0 || col > 2 {
		return 0, 0, fmt.Errorf("invalid grid position: row=%d, col=%d (must be 0-2)", row, col)
	}

	cellSize, err := c.CellSize()
	if err != nil {
		return 0, 0, err
	}

	// Each cell position: start + (cellIndex * (cellSize + spacing))
	x = col * (cellSize + c.SpacingPx)
	y = row * (cellSize + c.SpacingPx)

	return x, y, nil
}

// minRequiredSize returns the minimum artifact size required for the configured spacing.
// For a 3x3 grid with minimum cell size of 1px: 3 cells + 2 spacing gaps = 3 + 2*spacing
func (c LayoutConfig) minRequiredSize() int {
	return 3 + 2*c.SpacingPx
}

// validate checks that the configuration is valid.
// Returns a descriptive error if the configuration is invalid.
func (c LayoutConfig) validate() error {
	if c.ArtifactSizePx <= 0 {
		return fmt.Errorf("artifact-size-px must be positive, got %d", c.ArtifactSizePx)
	}
	if c.SpacingPx < 0 {
		return fmt.Errorf("spacing-px must be non-negative, got %d", c.SpacingPx)
	}
	return nil
}
