CLI command to run a multi-strategy parallel alchemist swarm.

- Validates requested strategies against `alchemist.StrategyNames`.
- Configures a `alchemist.SwarmConfig` and runs `alchemist.NewSwarm`.
- Supports `--first-win` to cancel remaining workers once a verified winner is found.
- Saves the swarm report to `.sin-code/swarm-report-*.md`.

Related files: `alchemist_cmd.go`, `alchemist_run_cmd.go`, `internal/alchemist/swarm.go`.
