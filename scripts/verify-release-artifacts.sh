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

archive_count="$(
  jq '[.[] | select(.type == "Archive")] | length' "$manifest"
)"
sbom_count="$(
  jq '[.[] | select(.type == "SBOM")] | length' "$manifest"
)"
test "$archive_count" -eq 4
test "$sbom_count" -eq 4

jq -e --arg artifact_root "$artifact_root" '
  def is_safe_path:
    (.path | type == "string") and
    (.path | startswith($artifact_root + "/")) and
    (.path | split("/") | all(. != "." and . != ".."));
  def is_root_file:
    .path == ($artifact_root + "/" + .name);
  all(
    .[];
    (.name | type == "string") and
    (.name | test("^[^/]+$")) and
    (.name != "." and .name != "..") and
    is_safe_path
  ) and
  ([.[].path] | unique | length == length) and
  all(
    .[] | select(
      .type == "Archive" or
      .type == "Source" or
      .type == "Checksum" or
      .type == "SBOM" or
      .type == "Signature"
    );
    is_root_file
  ) and
  all(
    .[] | select(.type == "Homebrew Cask");
    .path == ($artifact_root + "/homebrew/Casks/" + .name)
  )
' "$manifest" >/dev/null || {
  echo "artifact paths must be unique, safe, and match their release artifact type" >&2
  exit 1
}

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
