# github.go

GitHub REST search engine for repositories.

## Related files
- `common.go` — shared types.
- `search_engines_test.go` — mock tests.

## Important details
- Searches repositories via `api.github.com/search/repositories`.
- Optional `GITHUB_TOKEN` increases rate limits.
- Engagement is mapped from stargazers.

## Caveats
- Only repository search is implemented; issues and users are not.
- Returns `fmt.Errorf("github: %s")` for non-200 status.
