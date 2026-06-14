# resize_test.go

Unit tests for image resizing helpers.

## Related files
- `resize.go` — implementation under test.
- `resize_bench_test.go` — benchmarks.

## Important details
- Creates synthetic PNG/JPEG images and verifies aspect ratio preservation.
- Covers `ResizeImage`, `ResizeBatch`, missing input, and PNG/JPEG input/output.

## Caveats
- Uses `image.Decode` to verify output dimensions; not a visual quality check.
