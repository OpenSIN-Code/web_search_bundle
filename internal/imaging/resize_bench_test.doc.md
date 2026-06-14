# resize_bench_test.go

Benchmarks for image resizing and batch resizing hot paths.

## Related files
- `resize.go` — implementation.
- `resize_test.go` — unit tests.

## Important details
- Generates synthetic PNGs in `b.TempDir()`.
- Benchmarks `ResizeImage` at 1024 and 512 max widths, and `ResizeBatch` with 8 images.

## Caveats
- Benchmarks repeatedly encode the same gradients; real images may differ.
