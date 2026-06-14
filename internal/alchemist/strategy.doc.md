# strategy.go

Defines the strategy catalog used by the Swarm alchemist.

## Related files
- `swarm.go` — applies strategies to parallel workers.
- `strategy_test.go` — tests for strategy selection.
- `cmd/sin-websearch/alchemist_swarm_cmd.go` — CLI entry point.

## Important details
- Built-in strategies: conservative, aggressive, creative, minimal, literature-driven.
- Each strategy configures `MaxMutation`, `RiskAppetite`, `Model`, `Temperature`, and a prompt overlay.

## Caveats
- Model names reference Anthropic model aliases; verify availability before deployment.
- `GetStrategy` falls back to conservative for unknown names.
