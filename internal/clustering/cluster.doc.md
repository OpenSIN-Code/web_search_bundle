# cluster.go

Groups duplicate or overlapping stories across search sources using title similarity.

## Related files
- `cluster_test.go` / `cluster_bench_test.go` — tests.
- `orchestrator/orchestrator.go` — calls the clusterer on search results.

## Important details
- Uses a Jaccard similarity threshold of 0.6 over normalized word sets.
- Normalization keeps letters, numbers, and spaces; drops punctuation.
- First item in a cluster determines the cluster title and ID.

## Caveats
- O(n²) pairwise comparison; fine for small result sets but can be slow for large ones.
- English-centric normalization.
