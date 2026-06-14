# orchestrator_test.go

Hermetic unit tests for the multi-agent research mission orchestrator.

## Related files
- `orchestrator.go` — implementation under test.
- `orchestrator_bench_test.go` — benchmarks.
- `internal/orchestrator/orchestrator.go` — base fan-out orchestrator.
- `internal/profiles/profile.go` — profile definitions.
- `internal/engines/common.go` — `Engine` interface and `Result` type.

## Important details
- Uses deterministic `testEngine` to avoid network calls.
- Covers orchestrator creation, focus distribution, mission ID generation, librarian synthesis, explore agents, and the full `Run` flow.
- `Run` tests suppress stdout/stderr to keep the test output clean.

## Caveats
- `Run` delegates to the base orchestrator, so a working engine implementation is required for end-to-end tests.
- `generateMissionID` depends on the current time, so tests only check the prefix and non-empty value.
