# history_test.go

Smoke tests for the SQLite-backed experiment history store.

- Imports: `alchemist` (internal), `testing`, `time`
- Tests: insert, summary, recent queries, and the safe-rate helper.
- Uses `t.TempDir()` for isolated per-test databases.
