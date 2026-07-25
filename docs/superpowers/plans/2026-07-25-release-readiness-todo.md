# Release-Readiness TODO

This list tracks the remaining work identified on 2026-07-25. It is a working
queue; final evidence belongs in the implementation-plan index.

## Local work

- [ ] Add the original Bearm specification and eight historical implementation
  plans to version control.
- [ ] Add `scripts/closure-gates.sh` and enforce the Linux/macOS backend race
  checks with `-count=10` in CI.
- [ ] Validate a non-publishing snapshot and inspect archives, checksums,
  SBOMs, signature bundle, and Homebrew cask.
- [ ] Add the manual release-readiness workflow.
- [ ] Complete the Go 1.26 modernization plan.
- [ ] Restore exact Watchman Codex/OpenCode agent parity.

## Remote work

- [ ] Publish the reviewed `main` baseline to `Diaszano/bearm`.
- [ ] Create `Diaszano/homebrew-tap` and configure the least-privilege
  `HOMEBREW_TAP_GITHUB_TOKEN` secret.
- [ ] Run Ubuntu/macOS CI, security, and manual release-readiness workflows.
- [ ] Record run links, completion date, and commit range; then close plan
  checkboxes.

## Ordering

The local closure automation comes first because it makes every later remote
gate reproducible. Remote publication and credential provisioning are blocked
until they are explicitly authorized and available.
