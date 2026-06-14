# config_bench_test.go

Benchmarks for configuration loading and helper functions.

## Related files
- `config.go` — the implementation under test.
- `config_test.go` — unit tests.

## Important details
- Benchmarks the custom `itoa`, env key loading, full `Load`, and default cache path.
- Uses temporary HOME directories to avoid reading the real config.

## Caveats
- Benchmarks measure the default/no-config path, not production config with many keys.
