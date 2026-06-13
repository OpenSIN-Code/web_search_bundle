Multi-strategy parallel Alchemist coordinator.

Defines the `Swarm` type that runs several independent Alchemist daemons with different strategies (conservative, aggressive, creative, minimal, literature-driven). Each worker gets its own isolated Git branch and strategy-specific `program.md`. Results are aggregated from the shared SQLite history store.

Dependencies: standard library (`context`, `fmt`, `log/slog`, `os`, `path/filepath`, `sync`, `time`) and other `internal/alchemist` packages.
