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
trap cleanup EXIT
trap 'exit 129' HUP
trap 'exit 130' INT
trap 'exit 143' TERM

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
  checksums.txt.sigstore.json
do
  printf 'fixture\n' > "$fixture/dist/$artifact"
done

mkdir -p "$fixture/dist/homebrew/Casks" "$fixture/dist/bearm_linux_amd64_v1"
printf 'fixture\n' > "$fixture/dist/homebrew/Casks/bearm.rb"
printf 'fixture\n' > "$fixture/dist/bearm_linux_amd64_v1/bearm"

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
  {"name":"bearm","path":"dist/bearm_linux_amd64_v1/bearm","goos":"linux","goarch":"amd64","type":"Binary"},
  {"name":"bearm.rb","path":"dist/homebrew/Casks/bearm.rb","type":"Homebrew Cask"}
]
JSON

(cd "$fixture" && "$verifier" dist)

manifest="$fixture/dist/artifacts.json"
valid_manifest="$fixture/valid-artifacts.json"
cp "$manifest" "$valid_manifest"

printf 'fixture\n' > "$fixture/dist/bearm_linux_riscv64.tar.gz"
jq '. + [{"name":"bearm_linux_riscv64.tar.gz","path":"dist/bearm_linux_riscv64.tar.gz","goos":"linux","goarch":"riscv64","type":"Archive"}]' \
  "$valid_manifest" > "$manifest"
if (cd "$fixture" && "$verifier" dist >/dev/null 2>&1); then
  echo "extra archive unexpectedly passed" >&2
  exit 1
fi

cp "$valid_manifest" "$manifest"
printf 'fixture\n' > "$fixture/dist/unrelated.sbom.json"
jq '. + [{"name":"unrelated.sbom.json","path":"dist/unrelated.sbom.json","type":"SBOM"}]' \
  "$valid_manifest" > "$manifest"
if (cd "$fixture" && "$verifier" dist >/dev/null 2>&1); then
  echo "extra SBOM unexpectedly passed" >&2
  exit 1
fi

cp "$valid_manifest" "$manifest"
printf 'fixture\n' > "$fixture/outside.txt"
jq 'map(if .name == "bearm_linux_amd64.tar.gz" then .path = "dist/../outside.txt" else . end)' \
  "$valid_manifest" > "$manifest"
if (cd "$fixture" && "$verifier" dist >/dev/null 2>&1); then
  echo "traversal artifact path unexpectedly passed" >&2
  exit 1
fi

cp "$valid_manifest" "$manifest"
mkdir -p "$fixture/dist/unapproved/Casks"
cp "$fixture/dist/homebrew/Casks/bearm.rb" "$fixture/dist/unapproved/Casks/bearm.rb"
jq 'map(if .type == "Homebrew Cask" then .path = "dist/unapproved/Casks/bearm.rb" else . end)' \
  "$valid_manifest" > "$manifest"
if (cd "$fixture" && "$verifier" dist >/dev/null 2>&1); then
  echo "unexpected Homebrew cask path passed" >&2
  exit 1
fi

cp "$valid_manifest" "$manifest"
find "$fixture/dist" -name 'bearm_darwin_arm64.tar.gz' -delete
if (cd "$fixture" && "$verifier" dist >/dev/null 2>&1); then
  echo "missing referenced file unexpectedly passed" >&2
  exit 1
fi

echo "release artifact verifier tests passed"
