# literature_test.go

Smoke tests for the `LiteratureLoader` sin-websearch refresh.

- Imports: `alchemist` (internal), `context`, `os`, `path/filepath`, `testing`
- Creates a fake `sin-websearch` shell script that returns a valid mission JSON.
- Tests: refresh cadence, successful refresh, missing binary, and program.md injection.
