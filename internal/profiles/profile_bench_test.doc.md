# profile_bench_test.go

Benchmarks for profile YAML parsing, registry creation, and lookups.

## Related files
- `profile.go` — profile implementation and registry.
- `profile_test.go` — unit tests.

## Important details
- Uses a representative sample profile YAML.
- Benchmarks loading, builtin registry creation, get/list/add operations.

## Caveats
- `BenchmarkRegistryAdd` reuses the same profile pointer; may not represent real-world add patterns.
