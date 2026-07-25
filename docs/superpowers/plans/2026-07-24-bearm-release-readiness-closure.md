# Bearm Release-Readiness Closure Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Promote the implemented Bearm repository to a formally evidenced, release-ready state without creating a `v1.0.0` tag, GitHub Release, or production Homebrew cask.

**Architecture:** Use staged promotion gates: publish the reviewed local baseline, version the historical planning documents, codify host and artifact checks, prove the changes through pull-request CI, provision the external Homebrew tap and least-privilege secret, run a non-publishing release-readiness workflow, and only then close the historical plan checkboxes. Local scripts own repeatable checks; GitHub Actions owns cross-platform, security, OIDC signing, and remote evidence.

**Tech Stack:** Go 1.24+, POSIX shell, Git, GitHub CLI, GitHub Actions, GoReleaser v2, Syft, Cosign/Sigstore, Govulncheck, CodeQL, `jq`, Homebrew tap repository.

**Design specification:** [`../specs/2026-07-24-bearm-release-readiness-closure-design.md`](../specs/2026-07-24-bearm-release-readiness-closure-design.md)

## Global Constraints

- Use `superpowers:using-git-worktrees` before changing tracked files.
- Work on `chore/release-readiness-closure`, except for the final focused documentation branch.
- Preserve `.codex/`, `.gemini/`, and every unrelated working-tree change.
- Never create or push a tag, create a GitHub Release, or publish `Casks/bearm.rb`.
- Never obtain the Homebrew token from `gh auth token`.
- Never put the token in a command argument, repository file, temporary file, log, or status output.
- A failing product gate requires `superpowers:systematic-debugging` and `superpowers:test-driven-development` before a fix.
- Do not weaken a test, matrix entry, race count, security query, or artifact requirement.
- Use focused Conventional Commits without AI attribution.
- Check a historical box only when the planned commit subject exists, the implementation is present, and its current gate passes.
- Run `superpowers:verification-before-completion` before each completion claim.

---

### Task 1: Publish the Reviewed Baseline and Create the Isolated Worktree

**Files:**

- No tracked file changes.
- Preserve local untracked planning documents for Task 2.

- [ ] **Step 1: Verify the exact local baseline**

Run from the current repository root:

```bash
git branch --show-current
git rev-parse HEAD
git diff --quiet
git diff --cached --quiet
git remote get-url origin
git ls-remote --heads origin
```

Expected:

- branch is `main`;
- commit is `8c500f6e4ef0c62d827c2015e88d8ee457793216`;
- both diff checks exit `0`;
- origin is `git@github.com:Diaszano/bearm.git`;
- the final command prints no branch refs because the remote repository is empty.

If the commit differs, stop and review the new commits. Do not reset or overwrite local work.

- [ ] **Step 2: Confirm that no release state exists**

```bash
git tag --list
gh release list --repo Diaszano/bearm
```

Expected: neither command lists a tag or release.

- [ ] **Step 3: Publish only tracked `main`**

```bash
git push -u origin main
gh repo edit Diaszano/bearm --default-branch main
```

Expected: `main` is created on `origin`; unrelated untracked files are not pushed.

- [ ] **Step 4: Verify the published baseline**

```bash
test "$(git rev-parse main)" = "$(git ls-remote origin refs/heads/main | awk '{print $1}')"
test "$(gh repo view Diaszano/bearm --json defaultBranchRef --jq '.defaultBranchRef.name')" = main
gh api repos/Diaszano/bearm --jq '{visibility,default_branch}'
```

Expected JSON:

```json
{"default_branch":"main","visibility":"PUBLIC"}
```

- [ ] **Step 5: Create the isolated worktree**

Invoke `superpowers:using-git-worktrees` and create branch
`chore/release-readiness-closure` from `origin/main`.

After entering the worktree:

```bash
git branch --show-current
git merge-base --is-ancestor origin/main HEAD
git status --short
```

Expected: the branch is `chore/release-readiness-closure`, the ancestry check
exits `0`, and the worktree is clean.

---

### Task 2: Version the Historical Specification and Plans

**Files:**

- Add: `docs/superpowers/specs/2026-07-23-bearm-design.md`
- Add: `docs/superpowers/plans/2026-07-23-bearm-foundation.md`
- Add: `docs/superpowers/plans/2026-07-23-bearm-cli-compatibility.md`
- Add: `docs/superpowers/plans/2026-07-23-bearm-linux-trash.md`
- Add: `docs/superpowers/plans/2026-07-23-bearm-darwin-trash.md`
- Add: `docs/superpowers/plans/2026-07-23-bearm-removal-engine.md`
- Add: `docs/superpowers/plans/2026-07-23-bearm-journal-recovery.md`
- Add: `docs/superpowers/plans/2026-07-23-bearm-configuration-protection.md`
- Add: `docs/superpowers/plans/2026-07-23-bearm-compatibility-release.md`
- Add: `docs/superpowers/plans/2026-07-23-bearm-plan-index.md`
- Add: `docs/superpowers/plans/2026-07-24-bearm-release-readiness-closure.md`

- [ ] **Step 1: Resolve the source checkout without assuming its path**

The approved documents are intentionally untracked in the source checkout and
therefore are absent from a new worktree. Resolve that checkout from the shared
Git directory:

```bash
BEARM_SOURCE_ROOT="$(git -C "$(git rev-parse --git-common-dir)/.." rev-parse --show-toplevel)"
test -f "$BEARM_SOURCE_ROOT/docs/superpowers/specs/2026-07-23-bearm-design.md"
test -f "$BEARM_SOURCE_ROOT/docs/superpowers/plans/2026-07-24-bearm-release-readiness-closure.md"
```

Expected: both checks exit `0`. Do not delete or move the source copies.

- [ ] **Step 2: Copy only the approved documents**

This is a bulk mechanical copy of already reviewed documents:

```bash
cp -- "$BEARM_SOURCE_ROOT/docs/superpowers/specs/2026-07-23-bearm-design.md" docs/superpowers/specs/
cp -- "$BEARM_SOURCE_ROOT/docs/superpowers/plans/2026-07-23-bearm-foundation.md" docs/superpowers/plans/
cp -- "$BEARM_SOURCE_ROOT/docs/superpowers/plans/2026-07-23-bearm-cli-compatibility.md" docs/superpowers/plans/
cp -- "$BEARM_SOURCE_ROOT/docs/superpowers/plans/2026-07-23-bearm-linux-trash.md" docs/superpowers/plans/
cp -- "$BEARM_SOURCE_ROOT/docs/superpowers/plans/2026-07-23-bearm-darwin-trash.md" docs/superpowers/plans/
cp -- "$BEARM_SOURCE_ROOT/docs/superpowers/plans/2026-07-23-bearm-removal-engine.md" docs/superpowers/plans/
cp -- "$BEARM_SOURCE_ROOT/docs/superpowers/plans/2026-07-23-bearm-journal-recovery.md" docs/superpowers/plans/
cp -- "$BEARM_SOURCE_ROOT/docs/superpowers/plans/2026-07-23-bearm-configuration-protection.md" docs/superpowers/plans/
cp -- "$BEARM_SOURCE_ROOT/docs/superpowers/plans/2026-07-23-bearm-compatibility-release.md" docs/superpowers/plans/
cp -- "$BEARM_SOURCE_ROOT/docs/superpowers/plans/2026-07-23-bearm-plan-index.md" docs/superpowers/plans/
cp -- "$BEARM_SOURCE_ROOT/docs/superpowers/plans/2026-07-24-bearm-release-readiness-closure.md" docs/superpowers/plans/
```

- [ ] **Step 3: Verify scope and keep historical boxes open**

```bash
git status --short
test "$(rg -l '^- \[ \]' docs/superpowers/plans/2026-07-23-bearm-*.md | wc -l | tr -d ' ')" -ge 8
git diff --check -- docs/superpowers
```

Expected: only the eleven paths listed above are new. Historical checkboxes
remain unchecked because remote evidence does not exist yet.

- [ ] **Step 4: Commit the planning corpus**

```bash
git add \
  docs/superpowers/specs/2026-07-23-bearm-design.md \
  docs/superpowers/plans/2026-07-23-bearm-foundation.md \
  docs/superpowers/plans/2026-07-23-bearm-cli-compatibility.md \
  docs/superpowers/plans/2026-07-23-bearm-linux-trash.md \
  docs/superpowers/plans/2026-07-23-bearm-darwin-trash.md \
  docs/superpowers/plans/2026-07-23-bearm-removal-engine.md \
  docs/superpowers/plans/2026-07-23-bearm-journal-recovery.md \
  docs/superpowers/plans/2026-07-23-bearm-configuration-protection.md \
  docs/superpowers/plans/2026-07-23-bearm-compatibility-release.md \
  docs/superpowers/plans/2026-07-23-bearm-plan-index.md \
  docs/superpowers/plans/2026-07-24-bearm-release-readiness-closure.md
git diff --cached --check
git commit -m "docs: track Bearm implementation plans"
```

---

### Task 3: Codify the Host Closure Gates

**Files:**

- Create: `scripts/closure-gates.sh`
- Modify: `.github/workflows/ci.yml`

- [ ] **Step 1: Prove that the required backend repetition is not yet encoded**

```bash
! rg -q 'go test -race ./internal/trash/linux -count=10' .github/workflows/ci.yml
! rg -q 'go test -race ./internal/trash/darwin -count=10' .github/workflows/ci.yml
test ! -e scripts/closure-gates.sh
```

Expected: all three assertions exit `0`.

- [ ] **Step 2: Create the complete closure script**

Create `scripts/closure-gates.sh` with:

```sh
#!/bin/sh
set -eu

closure_root="$(mktemp -d /tmp/bearm-closure.XXXXXX)"
case "$closure_root" in
  /tmp/bearm-closure.*) ;;
  *)
    echo "refusing unsafe temporary path: $closure_root" >&2
    exit 1
    ;;
esac

cleanup() {
  find "$closure_root" -depth -delete
}
trap cleanup EXIT HUP INT TERM

closure_binary="$closure_root/bearm"
closure_work="$closure_root/work"
closure_config="$closure_root/config"
closure_state="$closure_root/state"
closure_trash="$closure_root/trash"
mkdir -p "$closure_work" "$closure_config" "$closure_state" "$closure_trash"

case "$(go env GOOS)" in
  linux)
    closure_profile=gnu
    closure_backend=./internal/trash/linux
    ;;
  darwin)
    closure_profile=bsd
    closure_backend=./internal/trash/darwin
    ;;
  *)
    echo "unsupported closure host: $(go env GOOS)" >&2
    exit 1
    ;;
esac

run_bearm() {
  env \
    BEARM_CONFIG_HOME="$closure_config" \
    BEARM_STATE_HOME="$closure_state" \
    BEARM_TRASH="$closure_trash" \
    BEARM_COMPAT="$closure_profile" \
    "$closure_binary" "$@"
}

go mod verify
make verify
go test ./test/compatibility -v
go test ./test/integration -v
go test -race "$closure_backend" -count=10
go test -race ./internal/journal ./internal/restore
make build
./bin/bearm version | grep -F 'bearm '
BEARM_COMPAT=gnu ./bin/bearm rm -f

go build -trimpath -o "$closure_binary" ./cmd/bearm
run_bearm version | grep -F 'bearm '
run_bearm config path | grep -F "$closure_config/config.toml"
run_bearm config check | grep -F 'Configuração válida.'

restore_source="$closure_work/restore.txt"
printf 'restore evidence\n' > "$restore_source"
run_bearm rm "$restore_source"
test ! -e "$restore_source"
run_bearm list --json | grep -F "$restore_source"
run_bearm restore --last
test "$(sed -n '1p' "$restore_source")" = 'restore evidence'

purge_source="$closure_work/purge.txt"
printf 'purge evidence\n' > "$purge_source"
run_bearm rm "$purge_source"
purge_item="$(
  run_bearm list |
    awk -F '	' -v expected="$purge_source" '$3 == expected { print $1; exit }'
)"
test -n "$purge_item"
run_bearm purge "$purge_item" --yes
test ! -e "$purge_source"
if run_bearm list --json | grep -F "$purge_source"; then
  echo "purged item remains active" >&2
  exit 1
fi
run_bearm doctor | grep -F 'Nenhum problema encontrado.'

alias_source="$closure_work/alias.txt"
printf 'alias evidence\n' > "$alias_source"
env \
  BEARM_ALIAS_COMMAND="$closure_binary rm" \
  BEARM_ALIAS_TARGET="$alias_source" \
  BEARM_CONFIG_HOME="$closure_config" \
  BEARM_STATE_HOME="$closure_state" \
  BEARM_TRASH="$closure_trash" \
  BEARM_COMPAT="$closure_profile" \
  /bin/sh <<'SH'
set -eu
alias rm="$BEARM_ALIAS_COMMAND"
alias rm | grep -F "$BEARM_ALIAS_COMMAND"
eval 'rm "$BEARM_ALIAS_TARGET"'
SH
test ! -e "$alias_source"
run_bearm list --json | grep -F "$alias_source"
run_bearm restore --last
test "$(sed -n '1p' "$alias_source")" = 'alias evidence'

echo "Bearm closure gates passed on $(go env GOOS)."
```

Make it executable:

```bash
chmod 0755 scripts/closure-gates.sh
```

- [ ] **Step 3: Raise both CI backend repetitions to ten**

In `.github/workflows/ci.yml`, make the two commands exactly:

```yaml
      - run: go test -race ./internal/trash/linux -count=10
```

and:

```yaml
      - run: go test -race ./internal/trash/darwin -count=10
```

- [ ] **Step 4: Run the new gate**

```bash
./scripts/closure-gates.sh
```

Expected final line on the current host:

```text
Bearm closure gates passed on linux.
```

On macOS, the final word is `darwin`.

- [ ] **Step 5: Commit**

```bash
git add scripts/closure-gates.sh .github/workflows/ci.yml
git diff --cached --check
git commit -m "test: codify Bearm closure gates"
```

---

### Task 4: Validate Every Required Snapshot Artifact

**Files:**

- Create: `scripts/verify-release-artifacts.sh`
- Create: `scripts/verify-release-artifacts_test.sh`
- Modify: `.goreleaser.yaml`

- [ ] **Step 1: Write the failing artifact-verifier test**

Create `scripts/verify-release-artifacts_test.sh` with:

```sh
#!/bin/sh
set -eu

test_root="$(mktemp -d /tmp/bearm-artifacts-test.XXXXXX)"
case "$test_root" in
  /tmp/bearm-artifacts-test.*) ;;
  *)
    echo "refusing unsafe temporary path: $test_root" >&2
    exit 1
    ;;
esac

cleanup() {
  find "$test_root" -depth -delete
}
trap cleanup EXIT HUP INT TERM

repository_root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
verifier="$repository_root/scripts/verify-release-artifacts.sh"
fixture="$test_root/fixture"
mkdir -p "$fixture/dist"

if (cd "$fixture" && "$verifier" dist >/dev/null 2>&1); then
  echo "missing manifest unexpectedly passed" >&2
  exit 1
fi

printf 'archive\n' > "$fixture/dist/bearm_linux_amd64.tar.gz"
cat > "$fixture/dist/artifacts.json" <<'JSON'
[
  {
    "name": "bearm_linux_amd64.tar.gz",
    "path": "dist/bearm_linux_amd64.tar.gz",
    "goos": "linux",
    "goarch": "amd64",
    "type": "Archive"
  }
]
JSON

if (cd "$fixture" && "$verifier" dist >/dev/null 2>&1); then
  echo "incomplete manifest unexpectedly passed" >&2
  exit 1
fi

for artifact in \
  bearm_linux_amd64.tar.gz \
  bearm_linux_arm64.tar.gz \
  bearm_darwin_amd64.tar.gz \
  bearm_darwin_arm64.tar.gz \
  bearm_source.tar.gz \
  checksums.txt \
  bearm_linux_amd64.tar.gz.sbom.json \
  bearm_linux_arm64.tar.gz.sbom.json \
  bearm_darwin_amd64.tar.gz.sbom.json \
  bearm_darwin_arm64.tar.gz.sbom.json \
  checksums.txt.sigstore.json \
  bearm.rb
do
  printf 'fixture\n' > "$fixture/dist/$artifact"
done

cat > "$fixture/dist/artifacts.json" <<'JSON'
[
  {"name":"bearm_linux_amd64.tar.gz","path":"dist/bearm_linux_amd64.tar.gz","goos":"linux","goarch":"amd64","type":"Archive"},
  {"name":"bearm_linux_arm64.tar.gz","path":"dist/bearm_linux_arm64.tar.gz","goos":"linux","goarch":"arm64","type":"Archive"},
  {"name":"bearm_darwin_amd64.tar.gz","path":"dist/bearm_darwin_amd64.tar.gz","goos":"darwin","goarch":"amd64","type":"Archive"},
  {"name":"bearm_darwin_arm64.tar.gz","path":"dist/bearm_darwin_arm64.tar.gz","goos":"darwin","goarch":"arm64","type":"Archive"},
  {"name":"bearm_source.tar.gz","path":"dist/bearm_source.tar.gz","type":"Source"},
  {"name":"checksums.txt","path":"dist/checksums.txt","type":"Checksum"},
  {"name":"bearm_linux_amd64.tar.gz.sbom.json","path":"dist/bearm_linux_amd64.tar.gz.sbom.json","type":"SBOM"},
  {"name":"bearm_linux_arm64.tar.gz.sbom.json","path":"dist/bearm_linux_arm64.tar.gz.sbom.json","type":"SBOM"},
  {"name":"bearm_darwin_amd64.tar.gz.sbom.json","path":"dist/bearm_darwin_amd64.tar.gz.sbom.json","type":"SBOM"},
  {"name":"bearm_darwin_arm64.tar.gz.sbom.json","path":"dist/bearm_darwin_arm64.tar.gz.sbom.json","type":"SBOM"},
  {"name":"checksums.txt.sigstore.json","path":"dist/checksums.txt.sigstore.json","type":"Signature"},
  {"name":"bearm.rb","path":"dist/bearm.rb","type":"Homebrew Cask"}
]
JSON

(cd "$fixture" && "$verifier" dist)

find "$fixture/dist" -name 'bearm_darwin_arm64.tar.gz' -delete
if (cd "$fixture" && "$verifier" dist >/dev/null 2>&1); then
  echo "missing referenced file unexpectedly passed" >&2
  exit 1
fi

echo "release artifact verifier tests passed"
```

Make it executable and run it:

```bash
chmod 0755 scripts/verify-release-artifacts_test.sh
./scripts/verify-release-artifacts_test.sh
```

Expected failure:

```text
scripts/verify-release-artifacts_test.sh: .../scripts/verify-release-artifacts.sh: not found
```

- [ ] **Step 2: Implement the artifact verifier**

Create `scripts/verify-release-artifacts.sh` with:

```sh
#!/bin/sh
set -eu

artifact_root="${1:-dist}"
case "$artifact_root" in
  /*|*..*)
    echo "artifact root must be a safe relative path: $artifact_root" >&2
    exit 1
    ;;
esac

manifest="$artifact_root/artifacts.json"
test -f "$manifest" || {
  echo "missing artifact manifest: $manifest" >&2
  exit 1
}
command -v jq >/dev/null 2>&1 || {
  echo "jq is required" >&2
  exit 1
}
jq -e 'type == "array"' "$manifest" >/dev/null

require_type() {
  artifact_type="$1"
  jq -e --arg artifact_type "$artifact_type" \
    'map(select(.type == $artifact_type)) | length > 0' \
    "$manifest" >/dev/null || {
      echo "missing artifact type: $artifact_type" >&2
      exit 1
    }
}

for artifact_type in Archive Source Checksum SBOM Signature "Homebrew Cask"
do
  require_type "$artifact_type"
done

test "$(jq '[.[] | select(.type == "SBOM")] | length' "$manifest")" -ge 4

for target in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64
do
  target_os="${target%/*}"
  target_arch="${target#*/}"
  jq -e \
    --arg target_os "$target_os" \
    --arg target_arch "$target_arch" \
    'map(select(
      .type == "Archive" and
      .goos == $target_os and
      .goarch == $target_arch
    )) | length == 1' \
    "$manifest" >/dev/null || {
      echo "missing unique archive for $target" >&2
      exit 1
    }
  archive_name="$(jq -er \
    --arg target_os "$target_os" \
    --arg target_arch "$target_arch" \
    '.[] | select(.type == "Archive" and .goos == $target_os and .goarch == $target_arch) | .name' \
    "$manifest")"
  jq -e --arg sbom_name "$archive_name.sbom.json" \
    'map(select(.type == "SBOM" and .name == $sbom_name)) | length == 1' \
    "$manifest" >/dev/null || {
      echo "missing archive SBOM for $archive_name" >&2
      exit 1
    }
done

jq -er '.[].path' "$manifest" |
while IFS= read -r artifact_path
do
  case "$artifact_path" in
    "$artifact_root"/*) ;;
    *)
      echo "artifact path escapes expected root: $artifact_path" >&2
      exit 1
      ;;
  esac
  test -f "$artifact_path" || {
    echo "manifest references missing file: $artifact_path" >&2
    exit 1
  }
done

checksum_count="$(
  jq '[.[] | select(.type == "Checksum")] | length' "$manifest"
)"
signature_count="$(
  jq '[.[] | select(.type == "Signature")] | length' "$manifest"
)"
source_count="$(
  jq '[.[] | select(.type == "Source")] | length' "$manifest"
)"
cask_count="$(
  jq '[.[] | select(.type == "Homebrew Cask")] | length' "$manifest"
)"
test "$checksum_count" -eq 1
test "$signature_count" -eq 1
test "$source_count" -eq 1
test "$cask_count" -eq 1

echo "release artifacts verified"
```

Make it executable:

```bash
chmod 0755 scripts/verify-release-artifacts.sh
```

- [ ] **Step 3: Run the verifier tests**

```bash
./scripts/verify-release-artifacts_test.sh
```

Expected:

```text
release artifacts verified
release artifact verifier tests passed
```

- [ ] **Step 4: Make snapshot cask generation explicitly non-publishing**

Add this field to the existing `.goreleaser.yaml` `homebrew_casks` entry,
immediately after `token`:

```yaml
    skip_upload: "{{ .IsSnapshot }}"
```

This produces the cask in `dist` for a snapshot while leaving a real tagged
release able to publish it later.

- [ ] **Step 5: Validate the GoReleaser configuration**

If GoReleaser is installed locally:

```bash
goreleaser check
```

Otherwise, validate in a container:

```bash
docker run --rm -v "$PWD:/workspace" -w /workspace goreleaser/goreleaser:v2.17.0 check
```

Expected: configuration is valid. If neither tool is available, stop; do not
commit unvalidated release configuration.

- [ ] **Step 6: Commit**

```bash
git add \
  scripts/verify-release-artifacts.sh \
  scripts/verify-release-artifacts_test.sh \
  .goreleaser.yaml
git diff --cached --check
git commit -m "test: verify Bearm release artifacts"
```

---

### Task 5: Add the Non-Publishing Release-Readiness Workflow

**Files:**

- Create: `.github/workflows/release-readiness.yml`

- [ ] **Step 1: Create the workflow**

Create `.github/workflows/release-readiness.yml` with:

```yaml
name: Release Readiness

on:
  workflow_dispatch:

permissions:
  contents: read
  id-token: write

jobs:
  readiness:
    runs-on: ubuntu-latest
    env:
      HOMEBREW_TAP_GITHUB_TOKEN: ${{ secrets.HOMEBREW_TAP_GITHUB_TOKEN }}
    steps:
      - uses: actions/checkout@v6
        with:
          fetch-depth: 0
      - uses: actions/setup-go@v6
        with:
          go-version: stable
          cache: true

      - name: Require and verify Homebrew tap credentials
        env:
          GH_TOKEN: ${{ secrets.HOMEBREW_TAP_GITHUB_TOKEN }}
        run: |
          if [ -z "$GH_TOKEN" ]; then
            echo "::error::HOMEBREW_TAP_GITHUB_TOKEN is required"
            exit 1
          fi
          permission="$(gh api repos/Diaszano/homebrew-tap --jq '.permissions.push')"
          if [ "$permission" != true ]; then
            echo "::error::Homebrew tap token does not have push permission"
            exit 1
          fi
          if gh api repos/Diaszano/homebrew-tap/contents/Casks/bearm.rb >/dev/null 2>&1; then
            echo "::error::production Homebrew cask already exists"
            exit 1
          fi

      - name: Run closure gates
        run: ./scripts/closure-gates.sh

      - name: Install and run Govulncheck
        run: |
          go install golang.org/x/vuln/cmd/govulncheck@latest
          govulncheck ./...

      - uses: anchore/sbom-action/download-syft@v0
      - uses: sigstore/cosign-installer@v4

      - name: Check GoReleaser configuration
        uses: goreleaser/goreleaser-action@v7
        with:
          distribution: goreleaser
          version: "~> v2"
          args: check

      - name: Build signed non-publishing snapshot
        uses: goreleaser/goreleaser-action@v7
        with:
          distribution: goreleaser
          version: "~> v2"
          args: release --snapshot --clean
        env:
          GITHUB_TOKEN: ${{ github.token }}
          HOMEBREW_TAP_GITHUB_TOKEN: ${{ secrets.HOMEBREW_TAP_GITHUB_TOKEN }}

      - name: Verify artifact inventory
        run: ./scripts/verify-release-artifacts.sh dist

      - name: Verify checksum signature
        run: |
          checksum_path="$(
            jq -er '.[] | select(.type == "Checksum") | .path' dist/artifacts.json
          )"
          signature_path="$(
            jq -er '.[] | select(.type == "Signature") | .path' dist/artifacts.json
          )"
          cosign verify-blob "$checksum_path" \
            --bundle "$signature_path" \
            --certificate-identity \
              "https://github.com/Diaszano/bearm/.github/workflows/release-readiness.yml@refs/heads/main" \
            --certificate-oidc-issuer \
              "https://token.actions.githubusercontent.com"

      - name: Prove that readiness did not publish
        env:
          GH_TOKEN: ${{ github.token }}
        run: |
          test "$(gh api repos/Diaszano/bearm/tags --jq 'length')" -eq 0
          test "$(gh api repos/Diaszano/bearm/releases --jq 'length')" -eq 0
          if gh api repos/Diaszano/homebrew-tap/contents/Casks/bearm.rb >/dev/null 2>&1; then
            echo "::error::readiness published a production cask"
            exit 1
          fi

      - name: Upload readiness evidence
        uses: actions/upload-artifact@v4
        with:
          name: bearm-release-readiness
          path: dist
          if-no-files-found: error
          retention-days: 7
```

- [ ] **Step 2: Validate workflow structure locally**

```bash
python3 -c 'import yaml; yaml.safe_load(open(".github/workflows/release-readiness.yml"))'
rg -n 'workflow_dispatch|id-token: write|release --snapshot --clean|retention-days: 7' .github/workflows/release-readiness.yml
! rg -n 'continue-on-error|release --clean$|gh release create|git tag|git push.*tag' .github/workflows/release-readiness.yml
```

Expected: YAML parses, all four required lines are found, and the forbidden
scan exits `0`. If PyYAML is unavailable, use the repository's configured YAML
linter; do not silently skip syntax validation.

- [ ] **Step 3: Run all local checks affected by the workflow**

```bash
./scripts/verify-release-artifacts_test.sh
./scripts/closure-gates.sh
git diff --check
```

Expected: all commands exit `0`.

- [ ] **Step 4: Commit**

```bash
git add .github/workflows/release-readiness.yml
git diff --cached --check
git commit -m "ci: add Bearm release-readiness workflow"
```

---

### Task 6: Prove and Merge the Closure Automation

**Files:**

- Modify only if a gate exposes a defect.

- [ ] **Step 1: Perform implementation and specification review**

Invoke `superpowers:requesting-code-review`. Confirm:

- the workflow cannot publish;
- secrets are never printed;
- all four archive targets and their SBOMs are required;
- Cosign verification pins the workflow and `main`;
- CI backend repetitions are exactly ten;
- unrelated untracked files are absent.

Address every accepted finding with a focused commit and rerun the affected
gate.

- [ ] **Step 2: Push and open the automation pull request**

```bash
git status --short
git push -u origin chore/release-readiness-closure
gh pr create \
  --repo Diaszano/bearm \
  --base main \
  --head chore/release-readiness-closure \
  --title "chore: close Bearm release-readiness gaps" \
  --body "Versions the approved plans, codifies closure gates, raises backend race repetition to ten, and adds a non-publishing release-readiness workflow."
```

Expected: `git status --short` is empty and `gh` prints the pull-request URL.

- [ ] **Step 3: Watch every pull-request gate**

```bash
gh pr checks --repo Diaszano/bearm --watch --fail-fast
```

Required successful checks include:

- CI test matrix on Ubuntu and macOS with Go `1.24.x` and stable;
- Linux and macOS backend race jobs with `-count=10`;
- acceptance on Ubuntu and macOS;
- snapshot, fuzz smoke, and benchmarks;
- Govulncheck;
- CodeQL with `security-extended`.

If a check fails, preserve its URL and logs, diagnose it, add the smallest
tested fix, push, and restart this step.

- [ ] **Step 4: Merge only the green pull request**

```bash
AUTOMATION_PR="$(gh pr view --repo Diaszano/bearm --json number --jq '.number')"
AUTOMATION_HEAD="$(git rev-parse HEAD)"
gh pr merge "$AUTOMATION_PR" \
  --repo Diaszano/bearm \
  --merge \
  --delete-branch \
  --match-head-commit "$AUTOMATION_HEAD"
git fetch origin main
git merge-base --is-ancestor "$AUTOMATION_HEAD" origin/main
```

Expected: merge succeeds and the ancestry check exits `0`.

---

### Task 7: Provision the Homebrew Tap and Least-Privilege Secret

**Files in `Diaszano/homebrew-tap`:**

- Create: `README.md`
- Create: `Casks/.gitkeep`

**GitHub state:**

- Create public repository: `Diaszano/homebrew-tap`
- Create Actions secret in `Diaszano/bearm`: `HOMEBREW_TAP_GITHUB_TOKEN`

- [ ] **Step 1: Confirm that the tap is still absent**

```bash
if gh repo view Diaszano/homebrew-tap >/dev/null 2>&1; then
  echo "Diaszano/homebrew-tap already exists; inspect it before continuing" >&2
  exit 1
fi
test ! -e /tmp/bearm-homebrew-tap-bootstrap
```

If the repository exists, stop rather than overwriting it.

- [ ] **Step 2: Create the bootstrap repository**

```bash
mkdir -p /tmp/bearm-homebrew-tap-bootstrap/Casks
git -C /tmp/bearm-homebrew-tap-bootstrap init -b main
```

Use `apply_patch` to add `/tmp/bearm-homebrew-tap-bootstrap/README.md`:

```markdown
# Diaszano Homebrew Tap

This repository contains Homebrew casks published by Diaszano projects.

Bearm release-readiness validation does not publish a production cask. The
first `Casks/bearm.rb` will be created only by an explicitly approved tagged
release.
```

Use `apply_patch` to add an empty
`/tmp/bearm-homebrew-tap-bootstrap/Casks/.gitkeep`.

- [ ] **Step 3: Commit and create the public tap**

```bash
git -C /tmp/bearm-homebrew-tap-bootstrap add README.md Casks/.gitkeep
git -C /tmp/bearm-homebrew-tap-bootstrap commit -m "chore: initialize Homebrew tap"
gh repo create Diaszano/homebrew-tap \
  --public \
  --description "Homebrew casks for Diaszano projects" \
  --source /tmp/bearm-homebrew-tap-bootstrap \
  --remote origin \
  --push
gh repo edit Diaszano/homebrew-tap --default-branch main
```

- [ ] **Step 4: Verify the tap before handling credentials**

```bash
test "$(gh repo view Diaszano/homebrew-tap --json visibility --jq '.visibility')" = PUBLIC
test "$(gh repo view Diaszano/homebrew-tap --json defaultBranchRef --jq '.defaultBranchRef.name')" = main
test "$(gh api repos/Diaszano/homebrew-tap/commits/main --jq '.commit.message')" = "chore: initialize Homebrew tap"
gh api repos/Diaszano/homebrew-tap/contents/Casks --jq '.[].name'
```

Expected: the final command lists only `.gitkeep`.

- [ ] **Step 5: Create the fine-grained PAT in GitHub's secure UI**

Create a new fine-grained personal access token with:

- resource owner `Diaszano`;
- repository access restricted to `Diaszano/homebrew-tap`;
- repository permission `Contents: Read and write`;
- no additional repository or organization permissions;
- an explicit expiration date.

Do not use the existing GitHub CLI token and do not paste the new value into
chat, shell history, a file, or a command argument.

- [ ] **Step 6: Store the token through the interactive encrypted flow**

Run:

```bash
gh secret set HOMEBREW_TAP_GITHUB_TOKEN --repo Diaszano/bearm
```

Paste the fine-grained PAT only into the hidden prompt, then submit it.

Verify only the secret metadata:

```bash
gh secret list --repo Diaszano/bearm |
  awk '$1 == "HOMEBREW_TAP_GITHUB_TOKEN" { found = 1 } END { exit !found }'
```

Expected: the check exits `0`; no secret value is shown.

- [ ] **Step 7: Remove only the validated bootstrap directory**

```bash
test "$(git -C /tmp/bearm-homebrew-tap-bootstrap remote get-url origin)" = "git@github.com:Diaszano/homebrew-tap.git"
find /tmp/bearm-homebrew-tap-bootstrap -depth -delete
test ! -e /tmp/bearm-homebrew-tap-bootstrap
```

The local bootstrap copy is removed; the GitHub repository remains recoverable
from its remote history.

---

### Task 8: Run and Inspect the Manual Release-Readiness Gate

**Files:**

- Read: workflow artifact `bearm-release-readiness`
- No tracked changes.

- [ ] **Step 1: Dispatch from merged `main`**

```bash
gh workflow run release-readiness.yml \
  --repo Diaszano/bearm \
  --ref main
```

Expected: `gh` prints the new workflow run URL.

- [ ] **Step 2: Resolve and watch the exact run**

```bash
READINESS_RUN="$(
  gh run list \
    --repo Diaszano/bearm \
    --workflow release-readiness.yml \
    --event workflow_dispatch \
    --branch main \
    --limit 1 \
    --json databaseId \
    --jq '.[0].databaseId'
)"
test -n "$READINESS_RUN"
gh run watch "$READINESS_RUN" \
  --repo Diaszano/bearm \
  --compact \
  --exit-status
```

Expected: the run concludes `success`. A transient Sigstore service failure
may be rerun, but an artifact, permission, code, or configuration failure must
be diagnosed and fixed through a new pull request.

- [ ] **Step 3: Download and independently inspect the evidence**

```bash
READINESS_EVIDENCE="$(mktemp -d /tmp/bearm-readiness-evidence.XXXXXX)"
case "$READINESS_EVIDENCE" in
  /tmp/bearm-readiness-evidence.*) ;;
  *) exit 1 ;;
esac
mkdir -p "$READINESS_EVIDENCE/dist"
gh run download "$READINESS_RUN" \
  --repo Diaszano/bearm \
  --name bearm-release-readiness \
  --dir "$READINESS_EVIDENCE/dist"
test -f "$READINESS_EVIDENCE/dist/artifacts.json"
BEARM_REPOSITORY_ROOT="$(git rev-parse --show-toplevel)"
(
  cd "$READINESS_EVIDENCE"
  "$BEARM_REPOSITORY_ROOT/scripts/verify-release-artifacts.sh" dist
)
```

Expected: `release artifacts verified`.

The verifier path must resolve from a checkout of merged `main`.

- [ ] **Step 4: Re-verify the signed checksum when Cosign is available**

```bash
CHECKSUM_PATH="$(
  jq -er '.[] | select(.type == "Checksum") | .path' \
    "$READINESS_EVIDENCE/dist/artifacts.json"
)"
SIGNATURE_PATH="$(
  jq -er '.[] | select(.type == "Signature") | .path' \
    "$READINESS_EVIDENCE/dist/artifacts.json"
)"
cosign verify-blob "$READINESS_EVIDENCE/$CHECKSUM_PATH" \
  --bundle "$READINESS_EVIDENCE/$SIGNATURE_PATH" \
  --certificate-identity \
    "https://github.com/Diaszano/bearm/.github/workflows/release-readiness.yml@refs/heads/main" \
  --certificate-oidc-issuer \
    "https://token.actions.githubusercontent.com"
```

Expected: Cosign reports successful verification. If Cosign is unavailable
locally, install it from the official Sigstore distribution before continuing;
the workflow result alone is not the independent inspection.

- [ ] **Step 5: Prove that no release was published**

```bash
test -z "$(git ls-remote --tags origin)"
test "$(gh release list --repo Diaszano/bearm --json tagName --jq 'length')" -eq 0
if gh api repos/Diaszano/homebrew-tap/contents/Casks/bearm.rb >/dev/null 2>&1; then
  echo "unexpected production cask" >&2
  exit 1
fi
```

Expected: all checks exit `0`.

- [ ] **Step 6: Remove only downloaded evidence after recording its run URL**

```bash
gh run view "$READINESS_RUN" --repo Diaszano/bearm --json url,conclusion
find "$READINESS_EVIDENCE" -depth -delete
test ! -e "$READINESS_EVIDENCE"
```

Copy the exact successful run URL for Task 9 before deleting the local
artifact copy.

---

### Task 9: Close the Historical Plans from Remote Evidence

**Files:**

- Modify: `docs/superpowers/plans/2026-07-23-bearm-foundation.md`
- Modify: `docs/superpowers/plans/2026-07-23-bearm-cli-compatibility.md`
- Modify: `docs/superpowers/plans/2026-07-23-bearm-linux-trash.md`
- Modify: `docs/superpowers/plans/2026-07-23-bearm-darwin-trash.md`
- Modify: `docs/superpowers/plans/2026-07-23-bearm-removal-engine.md`
- Modify: `docs/superpowers/plans/2026-07-23-bearm-journal-recovery.md`
- Modify: `docs/superpowers/plans/2026-07-23-bearm-configuration-protection.md`
- Modify: `docs/superpowers/plans/2026-07-23-bearm-compatibility-release.md`
- Modify: `docs/superpowers/plans/2026-07-23-bearm-plan-index.md`
- Modify: `docs/superpowers/plans/2026-07-24-bearm-release-readiness-closure.md`

- [ ] **Step 1: Create the focused documentation branch**

From a clean worktree:

```bash
git fetch origin main
git switch -c docs/close-bearm-plans origin/main
git status --short
```

Expected: the branch is clean and based on merged automation.

- [ ] **Step 2: Prove all 43 planned commit subjects**

```bash
PLANNED_SUBJECTS="$(mktemp /tmp/bearm-planned-subjects.XXXXXX)"
HISTORY_SUBJECTS="$(mktemp /tmp/bearm-history-subjects.XXXXXX)"
rg -o 'git commit -m "[^"]+"' docs/superpowers/plans/2026-07-23-bearm-*.md |
  sed 's/.*git commit -m "//; s/"$//' |
  sort -u > "$PLANNED_SUBJECTS"
git log origin/main --format=%s | sort -u > "$HISTORY_SUBJECTS"
test "$(wc -l < "$PLANNED_SUBJECTS" | tr -d ' ')" -eq 43
test -z "$(comm -23 "$PLANNED_SUBJECTS" "$HISTORY_SUBJECTS")"
find "$PLANNED_SUBJECTS" "$HISTORY_SUBJECTS" -delete
```

Expected: 43 unique planned subjects and no missing subject.

- [ ] **Step 3: Collect exact successful run URLs**

```bash
CLOSURE_SHA="$(git rev-parse origin/main)"
READINESS_RUN="$(
  gh run list \
    --repo Diaszano/bearm \
    --workflow release-readiness.yml \
    --event workflow_dispatch \
    --branch main \
    --status success \
    --limit 1 \
    --json databaseId \
    --jq '.[0].databaseId'
)"
test -n "$READINESS_RUN"
gh run list \
  --repo Diaszano/bearm \
  --workflow CI \
  --commit "$CLOSURE_SHA" \
  --status success \
  --limit 5 \
  --json url,workflowName,displayTitle,headSha
gh run list \
  --repo Diaszano/bearm \
  --workflow Security \
  --commit "$CLOSURE_SHA" \
  --status success \
  --limit 5 \
  --json url,workflowName,displayTitle,headSha
gh run view "$READINESS_RUN" \
  --repo Diaszano/bearm \
  --json url,conclusion,headSha
```

Use the exact URLs printed by these commands. Do not write a guessed URL or
an unevaluated shell expression into Markdown.

- [ ] **Step 4: Add the evidence section to the plan index**

Use `apply_patch` to add a `Release-Readiness Closure` section containing:

- closure date `2026-07-24`;
- the exact root-to-merged-closure commit range returned by Git;
- an evidence table with rows for pull-request CI, Security, and Release
  Readiness, populated with the exact successful URLs from Step 3;
- a statement that the run proved all four archives and archive SBOMs, the
  source archive, checksum, verified Sigstore bundle, and locally generated
  cask without publishing a tag, release, or production cask.

Also add
`2026-07-24-bearm-release-readiness-closure.md` to the index's plan list. Do
not save a guessed value or an unevaluated command in the document.

- [ ] **Step 5: Create a documentary evidence candidate**

```bash
git add docs/superpowers/plans/2026-07-23-bearm-plan-index.md
git diff --cached --check
git commit -m "docs: prepare Bearm closure evidence"
git push -u origin docs/close-bearm-plans
gh pr create \
  --repo Diaszano/bearm \
  --base main \
  --head docs/close-bearm-plans \
  --title "docs: close Bearm implementation plans" \
  --body "Records release-readiness evidence and closes only the plan steps proven by merged code and successful remote gates."
gh pr checks --repo Diaszano/bearm --watch --fail-fast
```

Expected: all candidate documentation checks pass. Record the exact
successful CI run URL; this is the non-self-referential documentary CI evidence
that will be written in the closing commit.

- [ ] **Step 6: Mark only now-proven plan steps complete**

Set task-specific metadata:

```bash
PLAN_COMPLETION_DATE=2026-07-24
IMPLEMENTATION_ROOT="$(git rev-list --max-parents=0 origin/main | tail -n 1)"
IMPLEMENTATION_TIP="$(git rev-parse origin/main)"
IMPLEMENTATION_RANGE="$IMPLEMENTATION_ROOT..$IMPLEMENTATION_TIP"
export PLAN_COMPLETION_DATE IMPLEMENTATION_RANGE
```

Apply the checkbox change as one bounded mechanical rewrite:

```bash
perl -pi -e 's/^- \[ \]/- [x]/' \
  docs/superpowers/plans/2026-07-23-bearm-foundation.md \
  docs/superpowers/plans/2026-07-23-bearm-cli-compatibility.md \
  docs/superpowers/plans/2026-07-23-bearm-linux-trash.md \
  docs/superpowers/plans/2026-07-23-bearm-darwin-trash.md \
  docs/superpowers/plans/2026-07-23-bearm-removal-engine.md \
  docs/superpowers/plans/2026-07-23-bearm-journal-recovery.md \
  docs/superpowers/plans/2026-07-23-bearm-configuration-protection.md \
  docs/superpowers/plans/2026-07-23-bearm-compatibility-release.md \
  docs/superpowers/plans/2026-07-24-bearm-release-readiness-closure.md
```

Use `apply_patch` to add `Status: Complete`, completion date `2026-07-24`, and
the exact value of `$IMPLEMENTATION_RANGE` immediately below each plan title.
Render these as bold Markdown field labels consistent with the existing
documents.

Add `**Status:** Complete` below the plan-index title. Add a fourth evidence
row containing the exact successful candidate documentation CI URL.

- [ ] **Step 7: Run every final documentary invariant**

```bash
! rg -n '^- \[ \]' \
  docs/superpowers/plans/2026-07-23-bearm-foundation.md \
  docs/superpowers/plans/2026-07-23-bearm-cli-compatibility.md \
  docs/superpowers/plans/2026-07-23-bearm-linux-trash.md \
  docs/superpowers/plans/2026-07-23-bearm-darwin-trash.md \
  docs/superpowers/plans/2026-07-23-bearm-removal-engine.md \
  docs/superpowers/plans/2026-07-23-bearm-journal-recovery.md \
  docs/superpowers/plans/2026-07-23-bearm-configuration-protection.md \
  docs/superpowers/plans/2026-07-23-bearm-compatibility-release.md \
  docs/superpowers/plans/2026-07-24-bearm-release-readiness-closure.md
test "$(rg -l '^\*\*Status:\*\* Complete$' docs/superpowers/plans/2026-07-*.md | wc -l | tr -d ' ')" -ge 10
! rg -n 'T[B]D|T[O]DO|example[.]com|<owner>|<repo>' docs/superpowers/plans
./scripts/closure-gates.sh
git diff --check
```

Expected: no unchecked boxes, no unresolved evidence instructions, all ten
documents report complete, closure gates pass, and the diff is clean.

- [ ] **Step 8: Commit the formal closure**

```bash
git add \
  docs/superpowers/plans/2026-07-23-bearm-foundation.md \
  docs/superpowers/plans/2026-07-23-bearm-cli-compatibility.md \
  docs/superpowers/plans/2026-07-23-bearm-linux-trash.md \
  docs/superpowers/plans/2026-07-23-bearm-darwin-trash.md \
  docs/superpowers/plans/2026-07-23-bearm-removal-engine.md \
  docs/superpowers/plans/2026-07-23-bearm-journal-recovery.md \
  docs/superpowers/plans/2026-07-23-bearm-configuration-protection.md \
  docs/superpowers/plans/2026-07-23-bearm-compatibility-release.md \
  docs/superpowers/plans/2026-07-23-bearm-plan-index.md \
  docs/superpowers/plans/2026-07-24-bearm-release-readiness-closure.md
git diff --cached --check
git commit -m "docs: close Bearm implementation plans"
git push
```

- [ ] **Step 9: Require green final checks and merge**

```bash
gh pr checks --repo Diaszano/bearm --watch --fail-fast
DOCUMENTATION_PR="$(
  gh pr view \
    --repo Diaszano/bearm \
    --json number \
    --jq '.number'
)"
DOCUMENTATION_HEAD="$(git rev-parse HEAD)"
gh pr merge "$DOCUMENTATION_PR" \
  --repo Diaszano/bearm \
  --merge \
  --delete-branch \
  --match-head-commit "$DOCUMENTATION_HEAD"
git fetch origin main
git merge-base --is-ancestor "$DOCUMENTATION_HEAD" origin/main
```

Expected: every final CI and security check is green and the documentary
closure is present on `origin/main`.

- [ ] **Step 10: Verify the release-ready boundary**

```bash
git show origin/main:docs/superpowers/plans/2026-07-23-bearm-plan-index.md |
  rg '^\*\*Status:\*\* Complete$'
test -z "$(git ls-remote --tags origin)"
test "$(gh release list --repo Diaszano/bearm --json tagName --jq 'length')" -eq 0
test "$(gh repo view Diaszano/homebrew-tap --json defaultBranchRef --jq '.defaultBranchRef.name')" = main
gh secret list --repo Diaszano/bearm |
  awk '$1 == "HOMEBREW_TAP_GITHUB_TOKEN" { found = 1 } END { exit !found }'
if gh api repos/Diaszano/homebrew-tap/contents/Casks/bearm.rb >/dev/null 2>&1; then
  echo "unexpected production cask" >&2
  exit 1
fi
```

Expected: all checks exit `0`. The terminal state is `release-ready`; creating
`v1.0.0`, a GitHub Release, or the production cask requires a separate,
explicitly approved release plan.
