# OpenCode Agent Definitions Design

## Goal

Create project-local OpenCode agent definitions for Bearm, Bob Chat, BufBear,
DentxDias, ModBear, Pratoo, and Watchman, using the existing `.codex/agents/`
sets as the source of required roles and `awesome-opencode-subagents` as the
source of OpenCode-compatible definitions.

## Scope

For each repository, create `.opencode/agents/` when it does not exist and add
one Markdown definition per configured Codex agent. Existing `.opencode/`
files, dependencies, locks, and memory are out of scope.

Each target file retains the Codex agent's filename stem so that the intended
project-local role remains recognizable. Its contents are copied verbatim from
the selected `awesome-opencode-subagents` definition.

## Mapping Rules

Use the catalog file with the same role name where it exists. Use these
semantic substitutions where the catalog has no identically named definition:

| Codex role | Catalog source |
| --- | --- |
| `reviewer` | `code-reviewer.md` |
| `codebase-orchestrator` | `workflow-orchestrator.md` |
| `docker-expert` | `devops-engineer.md` |
| `browser-debugger` | `debugger.md` |
| `ui-ux-tester` | `qa-expert.md` |

## Repository Sets

| Repository | Roles derived from `.codex/agents/` |
| --- | --- |
| Bearm | cli-developer, golang-pro, reviewer, security-auditor, test-automator |
| Bob Chat | backend-developer, nextjs-developer, reviewer, security-auditor, sql-pro, test-automator, typescript-pro |
| BufBear | build-engineer, dependency-manager, electron-pro, reviewer, security-auditor, test-automator, tooling-engineer, typescript-pro |
| DentxDias | backend-developer, codebase-orchestrator, deployment-engineer, docker-expert, fullstack-developer, golang-pro, nextjs-developer, reviewer, security-auditor, sql-pro, test-automator, typescript-pro |
| ModBear | build-engineer, dependency-manager, golang-pro, reviewer, security-auditor, test-automator, tooling-engineer, typescript-pro |
| Pratoo | backend-developer, cloud-architect, codebase-orchestrator, deployment-engineer, documentation-engineer, kotlin-specialist, mobile-developer, reviewer, security-auditor, test-automator, typescript-pro |
| Watchman | browser-debugger, build-engineer, dependency-manager, deployment-engineer, devops-engineer, docker-expert, frontend-developer, react-specialist, reviewer, security-auditor, test-automator, typescript-pro, ui-ux-tester |

## Verification

Verify that every role in each `.codex/agents/` directory has a corresponding
`.opencode/agents/<role>.md` file, that every target file matches its selected
catalog source byte-for-byte, and that no existing `.opencode/` configuration
file is modified.
