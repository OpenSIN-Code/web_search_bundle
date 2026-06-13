# program_md_test.go

Smoke tests for `ProgramMD` parsing and persistence.

- Imports: `alchemist` (internal), `os`, `path/filepath`, `strings`, `testing`
- Tests: loading, hypothesis queue, adding/appending, and missing-file handling.
- Verifies that `AddHypothesis` and `AppendLearning` persist changes to disk.
