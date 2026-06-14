# orchestrator.go

Multi-agent research mission orchestrator with explore and librarian agents.

## Related files
- `profiles/profile.go` — mission profile definitions.
- `orchestrator/orchestrator.go` — base search orchestrator.
- `verify/engine.go` — verification used on the combined results.
- `cmd/sin-websearch/mission_cmd.go` — CLI command.

## Important details
- Spawns `profile.Agents.Explore.Count` parallel explore agents.
- Each agent runs a base search with a focus (technical, community, market, creator).
- Synthesizes results with a librarian report and verification.
- Mission ID is based on nanosecond timestamp.

## Caveats
- `runLibrarianAgents` is a simple heuristic; not a real LLM agent.
- Fails the mission if no explore agent returns results.
