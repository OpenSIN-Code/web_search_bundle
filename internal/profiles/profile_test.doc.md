# profile_test.go

Hermetic unit tests for research profile loading and the built-in registry.

## Related files
- `profile.go` — profile implementation and registry.
- `mission/orchestrator.go` — consumes profiles.

## Important details
- Tests YAML loading, defaults, built-in profiles, custom directories, and registry add/get.
- Built-in profiles include competitive-analysis, person-dossier, market-landscape, crisis-monitoring, product-launch, technical-deep-dive.

## Caveats
- `NewRegistry("")` loads embedded/built-in profiles; tests depend on them.
