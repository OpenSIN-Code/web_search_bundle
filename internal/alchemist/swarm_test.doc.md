# swarm_test.go

Smoke tests for the multi-strategy alchemist swarm.

- Imports: `alchemist` (internal), `context`, `os`, `os/exec`, `path/filepath`, `strings`, `testing`, `time`
- Creates temporary git repos for each test to avoid mutating the real project.
- Tests: swarm construction, strategy program generation, markdown report rendering, and a full headless/auto-commit run.
