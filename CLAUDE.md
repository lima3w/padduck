# IPAM Next — Claude Instructions

## Milestone branch lifecycle (MANDATORY)

- Create `milestone/vX.Y.Z` from main at the start of each milestone.
- Develop and commit only on the milestone branch.
- Merge to main once all issues are closed and `make ci-local` passes.
- **Delete the milestone branch** (both remote and local) immediately after the merge is confirmed. If a milestone branch is open but its commits are already in main, delete it at the end of that session without a separate merge step.

```bash
git push origin --delete milestone/vX.Y.Z
git branch -d milestone/vX.Y.Z
```

## Commit workflow

1. Run `make ci-local` — never commit if it fails.
2. Commit with format `type(scope): summary`.
3. Push to the milestone branch immediately after every successful commit.

## Agent strategy

Spawn one backend agent and one frontend agent in parallel for each milestone. Both agents work on the same milestone branch concurrently.

## Architecture (non-negotiable)

Handler → Service → Repository → Database. All schema changes via migrations only. No business logic in handlers; no DB access outside repository.

## Testing (MANDATORY)

Tests are not optional follow-up work. They ship in the same commit as the code they cover.

### Rules

1. **New handler file → test file required in the same commit.**
   A commit adding `handlers/foo.go` must also add `handlers/foo_test.go`.
   Minimum coverage per endpoint: one happy-path test and one permission-denied test.

2. **New service method → unit test required in the same commit.**
   Any exported method added to a service file must have a corresponding test in `services/foo_test.go`.

3. **Bug fix → regression test required.**
   Every bug fix commit must include a test that fails before the fix and passes after.

4. **RBAC change → permission test required.**
   Adding or changing a permission constant must be accompanied by a test asserting the correct roles can and cannot exercise it.

### What to test

- Happy path: correct status code and response shape.
- Permission denied: a user without the required permission receives 403.
- Not found: a request for a non-existent resource returns 404.
- Validation: malformed input returns 400 with an error message.

### Pattern

Follow the style in `handlers/roles_test.go` and `services/rbac_test.go` — use the shared test DB helper, not mocks. Tests must pass under `go test -race`.
