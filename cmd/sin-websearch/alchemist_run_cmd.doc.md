CLI command to run a single alchemist autonomous research loop.

- Defines flags for budget, runtime, target file, metric, safety mode, and literature refresh.
- Builds an `alchemist.Config` and runs `alchemist.NewDaemon`.
- Handles graceful shutdown on SIGINT/SIGTERM.
- Saves the morning report to `.sin-code/alchemist-report-*.md`.

Related files: `alchemist_cmd.go`, `alchemist_swarm_cmd.go`, `internal/alchemist/daemon.go`.
