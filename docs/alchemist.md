# Alchemist Autoresearch

`sin-websearch alchemist` is a Karpathy-style autonomous research system. It runs a loop of hypothesis → experiment → verify → commit-or-discard, backed by git and SQLite.

## Quick start

```bash
# Create a program.md research plan
sin-websearch alchemist init --template go

# Run a single headless loop (no git mutations)
sin-websearch alchemist run \
  --target train.py \
  --metric val_bpb \
  --regex 'val_bpb:\s*([0-9\.]+)' \
  --cmd "python train.py --eval"

# Run overnight with auto-commit (still local, never pushed)
sin-websearch alchemist run \
  --safety auto-commit \
  --runtime 8h \
  --cmd "python train.py --eval"

# Multi-strategy swarm
sin-websearch alchemist swarm \
  --strategies conservative,aggressive,creative,minimal \
  --runtime 2h \
  --cmd "go test -bench=."
```

## Safety modes

- `headless` (default): logs results, no git changes (M4 safe).
- `auto-commit`: commits locally on improvement, never pushes.
- `interactive`: asks before every commit/push.

## program.md format

```markdown
# Research Program

## Hypothesis Queue

- [ ] Increase batch size from 32 to 64
- [ ] Add dropout 0.1
- [ ] Tune learning rate

## Learnings

- Baseline established at 0.42
```

The agent updates the queue and learnings as it runs.

## Swarm strategies

| Strategy | Risk | Max mutation | Description |
|---|---|---|---|
| conservative | 0.1 | 20 lines | Minimal, surgical changes |
| aggressive | 0.8 | 200 lines | Structural refactors |
| creative | 0.7 | 150 lines | Cross-domain ideas |
| minimal | 0.05 | 5 lines | Control group |
| literature-driven | 0.5 | 100 lines | Hypotheses from sin-websearch |

## Literature-Loader

The daemon can refresh hypotheses periodically by running a `sin-websearch mission` on your target topic:

```bash
sin-websearch alchemist run \
  --literature-refresh 10 \
  --literature-profile technical-deep-dive \
  --cmd "python train.py --eval"
```

Set `--literature-refresh 0` to disable.

## MCP tool

When `sin-websearch serve` is running, agents can call:

```json
{
  "tool": "websearch_alchemist",
  "arguments": {
    "run_cmd": "python train.py --eval",
    "target": "train.py",
    "metric": "val_bpb",
    "regex": "val_bpb:\\s*([0-9\\.]+)",
    "max_experiments": 3,
    "safety": "headless"
  }
}
```

Add `strategies` to run in swarm mode.

## Notes

- The alchemist creates a local work branch `alchemist/<timestamp>`.
- It never pushes automatically; review and push manually.
- History is stored in `.sin-code/alchemist.db`.
