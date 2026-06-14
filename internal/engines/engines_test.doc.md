# engines_test.go

Hermetic unit tests for engine helper functions and constructors.

## Related files
- All engine files (the private helpers tested here are spread across the package).
- `sidecar/manager.go` — used for video/YouTube engine creation.

## Important details
- Tests shared helpers: `truncate`, `containsInsensitive`, `parseTime`, `parseVideoTime`, `parseVTT`, `stripHTML`, `dedupeLines`, `detectVideoSource`, `detectWhisperPref`.
- Verifies engine `Name()` methods.
- Uses a temp sidecar manager for hermetic YouTube/video tests.

## Caveats
- Many helpers are private and defined in different engine files.
- Test file is large because it covers helpers across the package.
