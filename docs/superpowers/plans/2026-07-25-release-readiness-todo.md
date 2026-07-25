# Release-Readiness TODO

This list tracks the remaining work identified on 2026-07-25. It is a working
queue; final evidence belongs in the implementation-plan index.

## Local work

- [x] Add the original Bearm specification and eight historical implementation
  plans to version control.
- [x] Add `scripts/closure-gates.sh` and enforce the Linux/macOS backend race
  checks with `-count=10` in CI.
- [x] Validate a non-publishing snapshot and inspect archives, checksums,
  SBOMs, signature bundle, and Homebrew cask.
- [x] Add the manual release-readiness workflow.
- [x] Complete the Go 1.26 modernization plan.
- [x] Restore exact Watchman Codex/OpenCode agent parity.

## Remote work

- [x] Publish the reviewed `main` baseline to `Diaszano/bearm`.
- [x] Create `Diaszano/homebrew-tap` and configure
  `HOMEBREW_TAP_GITHUB_TOKEN` for readiness validation.
- [ ] Replace the user-authorized broad CLI token with a least-privilege token
  restricted to `Diaszano/homebrew-tap` before a production release.
- [x] Run Ubuntu/macOS CI, security, and manual release-readiness workflows.
- [x] Record run links, completion date, and commit range; then close plan
  checkboxes.

## Ordering

The local closure automation comes first because it makes every later remote
gate reproducible. Remote publication and credential provisioning are blocked
until they are explicitly authorized and available.
