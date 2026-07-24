# Bearm Release-Readiness Closure Design

**Status:** Approved for planning  
**Date:** 2026-07-24  
**Project:** Bearm  
**Target state:** Release-ready without creating `v1.0.0`

## 1. Purpose

This specification defines how to close Bearm's eight existing Superpowers
implementation plans with verifiable local and remote evidence.

The implementation already exists on local `main`, and all 43 commit subjects
prescribed by the plans are present. Local Linux acceptance passes. Formal
closure remains incomplete because the plans and original specification are
untracked, the GitHub repository is empty, macOS and remote security gates
have not run, release tooling is unavailable locally, the Homebrew tap does
not exist, and the plan checkboxes remain open.

## 2. Goals

The closure work must:

1. Publish the current local `main` baseline to `Diaszano/bearm`.
2. Version the approved design specification and all eight implementation
   plans.
3. Make every plan exit gate reproducible in local scripts or GitHub Actions.
4. Pass the required Go 1.24 and stable-Go matrix on Ubuntu and macOS.
5. Pass Linux and macOS backend race tests with `-count=10`.
6. Pass Govulncheck and CodeQL.
7. Provision the public `Diaszano/homebrew-tap` repository.
8. Provision a least-privilege fine-grained PAT as
   `HOMEBREW_TAP_GITHUB_TOKEN`.
9. Add and run a manual release-readiness workflow.
10. Generate and inspect snapshot archives, source archive, checksum, SBOMs,
    Sigstore signature bundle, and Homebrew cask without publishing a release.
11. Mark historical plan checkboxes only after their evidence exists.
12. Record completion status, date, commit range, and remote run links in the
    plan documents and index.

## 3. Non-Goals

This work does not:

1. Create or push a `v1.0.0` tag.
2. Publish a GitHub Release.
3. Publish a production Homebrew cask.
4. Replace or alias the user's host `/bin/rm`.
5. Change Bearm product behavior unless a failing closure gate exposes a
   defect.
6. Commit `.codex/` or `.gemini/` content unless a separately approved project
   requirement makes a specific file necessary.
7. Weaken safety, compatibility, security, race, or release checks to obtain a
   green result.

## 4. Promotion Model

Closure is a gated promotion:

```text
implemented locally
        |
        v
documentation versioned
        |
        v
main published
        |
        v
Linux and macOS CI green
        |
        v
tap and secret provisioned
        |
        v
snapshot, SBOM, signature, and cask validated
        |
        v
checkboxes and statuses closed
        |
        v
release-ready
```

Each transition requires fresh evidence. A failure leaves the project in its
current state and does not advance documentary status.

## 5. Repository and Branch Workflow

### 5.1 Baseline Publication

The local `main` branch at the reviewed implementation commit is pushed to the
empty public repository `Diaszano/bearm`, and GitHub `main` becomes the default
branch.

The initial push contains only tracked project files. Existing unrelated
untracked files remain local.

### 5.2 Closure Development

After baseline publication:

1. create an isolated worktree;
2. create branch `chore/release-readiness-closure`;
3. add the specifications, plans, exact gates, and manual readiness workflow;
4. commit focused Conventional Commits;
5. push the branch and open a pull request;
6. merge only after all pull-request gates pass.

Product defects found by a gate are fixed through a failing test, minimal
implementation, and focused Conventional Commit. Workflow-only defects are
verified with the relevant validator and rerun.

### 5.3 Documentary Closure

After the readiness workflow passes, create a focused documentation branch and
commit that:

- changes proved historical steps from `- [ ]` to `- [x]`;
- sets each plan status to `Complete`;
- records the completion date;
- records the implemented commit range;
- adds CI and readiness run links to the plan index;
- closes the release-readiness plan itself.

The documentary closure commit must pass CI before merge.

## 6. GitHub Provisioning

### 6.1 Bearm Repository

`Diaszano/bearm` is public. The closure process must:

- publish `main`;
- set `main` as the default branch;
- verify that CI and security workflows can run;
- avoid creating a release or tag.

### 6.2 Homebrew Tap

Create public repository `Diaszano/homebrew-tap` with:

```text
README.md
Casks/
```

Initialize branch `main` with commit:

```text
chore: initialize Homebrew tap
```

No production cask is committed during readiness validation.

### 6.3 Fine-Grained PAT

The cross-repository publishing token must:

- be a fine-grained personal access token;
- have access only to `Diaszano/homebrew-tap`;
- have repository permission `Contents: Read and write`;
- have an explicit expiration date;
- have no organization-wide or unrelated repository access.

The token is stored in `Diaszano/bearm` as:

```text
HOMEBREW_TAP_GITHUB_TOKEN
```

The token value must never:

- be derived from `gh auth token`;
- be passed as a command-line argument;
- be written to a repository or temporary file;
- be printed in logs or status output.

Provisioning uses an interactive or standard-input secret flow. Verification
may list the secret name and inspect repository permissions but must not read
the secret value.

## 7. Required Automation

### 7.1 Local Closure Script

Add a closure script that runs the plan exit gates available on the current
host:

- formatting;
- module verification;
- vet;
- complete tests;
- race tests;
- compatibility tests;
- integration tests;
- platform backend repetition;
- configuration check;
- build and version smoke test;
- disposable remove/list/restore/purge/doctor/config scenario.

The script detects the host and executes Linux- or macOS-specific backend
checks. It exits nonzero on the first failed gate.

### 7.2 Pull-Request CI

CI must include:

- Ubuntu and macOS;
- Go `1.24.x` and stable;
- `go test ./...`;
- `go test -race ./...`;
- plan-specific Linux and macOS backend tests with `-count=10`;
- compatibility and integration tests;
- snapshot configuration validation;
- formatting and vet;
- fuzz smoke and benchmarks as already configured.

No Linux result substitutes for a macOS gate.

### 7.3 Security

Required green jobs:

- Govulncheck on the complete module;
- CodeQL Go analysis with `security-extended` queries.

The scheduled security workflow remains enabled after closure.

### 7.4 Manual Release-Readiness Workflow

Add a `workflow_dispatch` workflow that runs on `main` with:

```yaml
permissions:
  contents: read
  id-token: write
```

The workflow must:

1. verify that `HOMEBREW_TAP_GITHUB_TOKEN` is present;
2. query `Diaszano/homebrew-tap` using that token and require push permission;
3. run formatting, module verification, vet, tests, race tests, and
   Govulncheck;
4. install Syft;
5. install Cosign;
6. validate `.goreleaser.yaml`;
7. run a non-publishing snapshot;
8. generate archives for Linux and macOS on `amd64` and `arm64`;
9. generate a source archive;
10. generate `checksums.txt`;
11. generate archive SBOMs;
12. sign the checksum through GitHub OIDC into a `.sigstore.json` bundle;
13. generate or validate the Homebrew cask definition;
14. inspect `dist/artifacts.json` for every required artifact class;
15. upload the readiness artifacts with a bounded retention period.

The workflow must not:

- create a Git tag;
- create a GitHub Release;
- push a cask to the tap;
- expose a secret.

## 8. Evidence Model

### 8.1 Historical Tasks

Historical implementation steps are closed retroactively only when:

- their exact planned commit subject exists in history;
- the expected files or behavior exist in `main`;
- the corresponding current test or exit gate passes.

The red phase of historical TDD is evidenced by the task commit sequence and
reviewed tests. The closure process does not rewrite history or temporarily
revert production code merely to recreate a historical failure.

### 8.2 External Gates

Remote steps require links to successful GitHub Actions runs:

- pull-request CI;
- Ubuntu acceptance;
- macOS acceptance;
- security;
- manual release readiness;
- final documentary CI.

Run links are recorded in the implementation plan index. No separate evidence
ledger is created.

### 8.3 Checkboxes

Checkboxes are changed only after evidence exists. If one step cannot be
proved, it remains open and its parent plan does not receive `Status:
Complete`.

## 9. Failure Handling

### 9.1 Code or Compatibility Failure

1. preserve the failing output;
2. reproduce on the failing operating system;
3. add a regression test;
4. verify the test fails for the intended reason;
5. implement the minimal fix;
6. rerun the focused and complete gates;
7. commit separately.

### 9.2 Workflow Failure

Validate workflow syntax and action versions against official primary
documentation. Do not bypass a failed check with `continue-on-error`, a broad
exclusion, or reduced matrix coverage.

### 9.3 Credential Failure

Stop the readiness workflow before artifact generation when the secret is
missing or cannot write to the tap. Rotate or reprovision the fine-grained PAT
without printing it.

### 9.4 Artifact Failure

Missing archives, SBOMs, checksums, signatures, source archive, or cask
metadata fail the readiness workflow. Partial artifact sets are not accepted.

## 10. Acceptance Criteria

Bearm is release-ready when all are true:

1. `main` is published to `Diaszano/bearm`.
2. `main` is the GitHub default branch.
3. The original design and all implementation plans are tracked.
4. All 43 planned commit subjects are present.
5. Local closure gates pass on the current host.
6. Ubuntu Go 1.24 and stable jobs pass.
7. macOS Go 1.24 and stable jobs pass.
8. Linux backend race repetition passes with `-count=10`.
9. macOS backend race repetition passes with `-count=10`.
10. Compatibility and integration suites pass on Ubuntu and macOS.
11. Govulncheck and CodeQL pass.
12. `Diaszano/homebrew-tap` exists and has `main`.
13. `HOMEBREW_TAP_GITHUB_TOKEN` exists with verified tap write access.
14. The manual readiness run produces every required artifact.
15. The checksum signature verifies with Cosign.
16. No tag, GitHub Release, or production cask was created.
17. Each historical plan is marked `Status: Complete`.
18. Every proved step is marked `[x]`.
19. The plan index contains the successful remote run links.
20. Final documentary CI passes on Ubuntu and macOS.

## 11. Completion Boundary

The terminal state is:

```text
release-ready
```

Creating `v1.0.0`, publishing a GitHub Release, and publishing the Homebrew
cask require a separate explicitly approved release plan.
