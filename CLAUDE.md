MILESTONE BRANCHES: Create milestone/vX.Y.Z from main. Merge to main only when all issues closed and make ci-local passes. Delete remote+local branch immediately after merge.

COMMITS: make ci-local must pass before every commit. Format: type(scope): summary. Push to milestone branch immediately after every commit.

AGENTS: Spawn backend+frontend agents in parallel for each milestone.

ARCH (non-negotiable): handler->service->repository->database. Schema via migrations only. No business logic in handlers; no DB access outside repository.

TESTING (mandatory, same commit as the code):
new handler file: add handlers/foo_test.go; min 1 happy-path + 1 403 per endpoint
new service method: add unit test in services/foo_test.go
bug fix: regression test (fails before fix, passes after)
rbac change: test correct roles can/cannot exercise permission
cover: 200 happy-path, 403 no-permission, 404 not-found, 400 bad-input
style: follow handlers/roles_test.go + services/rbac_test.go; shared test DB helper, not mocks; go test -race
