# Contributing to Bearm

## Prerequisites

- Go 1.27.3 or newer;
- Git;
- `make`;
- a Linux or macOS test environment.

Go tools are pinned in `tools/go.mod` and run through the project Makefile and
acceptance script; a global installation is not required. The `go` command
automatically selects the required toolchain when needed.

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

### Conventional Commits Specification

All commit messages and Pull Request titles must follow the [Conventional Commits v1.0.0](https://www.conventionalcommits.org/) specification:

```text
<type>(<optional scope>): <description>
```

#### Allowed Types:
- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation only changes
- `style`: Changes that do not affect code logic (formatting, white-space)
- `refactor`: Code changes that neither fix a bug nor add a feature
- `perf`: A code change that improves performance
- `test`: Adding missing tests or correcting existing tests
- `build`: Changes that affect the build system or external dependencies
- `ci`: Changes to CI configuration files and scripts
- `chore`: Other changes that don't modify src or test files
- `revert`: Reverts a previous commit

#### Breaking Changes:
Append `!` before the colon (e.g., `feat(api)!: change default restore policy`) or include `BREAKING CHANGE:` in the commit footer.

Commit examples:

```text
feat(cli): add native help command
fix(trash): support filesystems without renameat2
docs(recovery): explain journal repair procedure
test: cover GNU option precedence
chore: update dependencies
```

Do not add AI attribution trailers. Source code, identifiers, and comments are
written in English. English is the source language for public documentation;
provide a matching Brazilian Portuguese (`.pt-BR.md`) version when updating it.

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
