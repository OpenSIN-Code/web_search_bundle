# config_test.go

Hermetic unit tests for configuration loading and helper functions.

## Related files
- `config.go` — configuration implementation.
- `config_bench_test.go` — benchmarks.

## Important details
- Tests default/fallback cache paths, env overrides, file loading, and backwards-compatible env vars.
- Clears relevant env vars to avoid cross-test pollution.

## Caveats
- Some tests depend on `HOME` being set to a temporary directory.
- `MustLoad` is exercised with a clean HOME; it will panic if the real config is broken.
