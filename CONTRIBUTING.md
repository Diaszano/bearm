# Contributing to Bearm

## Prerequisites

- Go 1.24 or newer;
- Git;
- `make`;
- a Linux or macOS test environment.

GoReleaser, `golangci-lint`, and `govulncheck` are recommended for release and
security validation.

## Workflow

1. Create a conventional English branch name such as
   `feat/linux-trash-backend`, `fix/restore-collision`, or
   `docs/configuration`.
2. Write a failing test for each behavior change and confirm the failure is
   caused by the missing behavior.
3. Implement the smallest change that passes the test.
4. Run:

   ```bash
   make verify
   ```

5. Keep each task in one focused Conventional Commit.

Commit examples:

```text
feat: add FreeDesktop trash metadata
fix: retain dangling symlink targets
test: cover GNU option precedence
docs: explain restore collisions
```

Do not add AI attribution trailers. Source code, identifiers, comments,
technical documentation, branches, and commits are written in English.

## Compatibility Changes

A compatibility behavior change must include a fixture that runs both the host
`rm` and the test-built Bearm in an isolated temporary directory. Compare the
observable contract that is stable for that case:

- exit status;
- standard output;
- standard error or prompt;
- normalized source-tree outcome;
- traversal order when relevant.

Never weaken Bearm's hard safety invariants merely to match host `rm`.

## Tests

Before submitting a change:

```bash
test -z "$(gofmt -l .)"
go mod verify
go vet ./...
go test ./...
go test -race ./...
```

Platform backend changes must run on their target operating system. Tests may
operate only in isolated temporary directories and must never invoke removal
against user data.

## Public Go APIs

Document every exported Go declaration with a standard Go documentation
comment. Prefer focused files and interfaces that keep parsing, policy,
filesystem mutation, journaling, and rendering separate.

## Security

Do not open a public issue for a suspected unpatched vulnerability. Follow
[SECURITY.md](SECURITY.md).
