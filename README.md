# collage-it

`collage-it` is a Go command-line tool that combines 9 photos into a single
square 3x3 collage image. Each input photo is center-cropped to a square
before being placed on the grid.

## Requirements

- Go 1.22 or later

## Build

```bash
go build -o collage-it ./cmd/collage-it
```

## Usage

Place at least 9 target image files in a directory, then run the tool from
that directory (or build the binary and run it there):

```bash
go run ./cmd/collage-it -name-prefix trip_ -spacing-px 16 -artifact-size-px 2048
```

- The tool scans the current directory for image files.
- At least 9 matching images must be found, or the command fails with an
  error such as `found 5 image files, want at least 9`.
- If more than 9 matching images are found, the 9 files with the newest
  modification times are selected. Files with equal modification times are
  resolved by filename.
- Supported image extensions: `.jpg`, `.jpeg`, `.png`.
- The selected files are sorted by filename for a stable, reproducible layout.

### Flags

| Flag | Default | Description |
| --- | --- | --- |
| `-name-prefix` | `""` (no filter) | Only files whose name starts with this prefix are considered as input images. |
| `-spacing-px` | `16` | Spacing between photos, in pixels. Must be `>= 0`. |
| `-artifact-size-px` | `2048` | Side length of the output square image, in pixels. Must be `> 0`. |

Invalid flag values (e.g. a negative `-spacing-px`) cause the command to exit
with a non-zero status and a descriptive error message on stderr.

## Output

The collage is written to a fixed file named `collage.jpg` in the current
directory, encoded as JPEG (quality 95).

## License

Distributed under the [MIT License](LICENSE).