// Purpose: Default program.md templates for alchemist init.
// Docs: alchemist_template.doc.md
package main

func getProgramTemplate(kind string) string {
	switch kind {
	case "python", "ml":
		return `# program.md — Autonomous Research Loop

## The Setup
- **Target File:** ` + "`train.py`" + ` (agent modifies this)
- **Immutable Files:** ` + "`prepare.py`" + `, ` + "`evaluate.py`" + ` (DO NOT MODIFY)
- **Metric:** ` + "`val_bpb`" + ` (lower is better)
- **Time Budget:** 5 minutes per experiment (wall clock)

## Rules
1. Formulate hypothesis from the queue below.
2. Modify ONLY the target file.
3. Run verification command.
4. If metric improves → commit. Otherwise → reset.
5. Update Learnings section after every run.

## Hypothesis Queue
- [ ] Try GELU activation instead of ReLU
- [ ] Add layer normalization before attention
- [ ] Reduce vocab_size to 4096
- [ ] Increase DEPTH to 12

## Learnings (agent updates this)
`

	default: // "go"
		return `# program.md — Autonomous Go Performance Research

## The Setup
- **Target File:** ` + "`internal/dispatcher/dispatcher.go`" + `
- **Immutable Files:** ` + "`*_test.go`" + ` files (DO NOT MODIFY)
- **Metric:** ` + "`ops_per_sec`" + ` from ` + "`go test -bench=.`" + ` (higher is better)
- **Time Budget:** 5 minutes per experiment

## Rules
1. Pick next hypothesis from queue.
2. Modify ONLY the target file.
3. Run ` + "`go test -bench=. -run=^$ -count=3`" + `
4. If ops/sec improves → commit. Otherwise → reset.
5. Append learning below.

## Hypothesis Queue
- [ ] Replace sync.Mutex with sync.RWMutex in hot path
- [ ] Use sync.Pool for request buffers
- [ ] Pre-allocate slices with known capacity
- [ ] Switch from map[string]X to map[int]X with interned keys
- [ ] Batch channel sends (10 at a time)

## Learnings (agent updates this)
`
	}
}
