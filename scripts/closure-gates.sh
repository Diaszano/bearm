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
trap cleanup EXIT
trap 'exit 129' HUP
trap 'exit 130' INT
trap 'exit 143' TERM

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
./bin/bearm version | grep -F 'Bearm '
BEARM_COMPAT=gnu ./bin/bearm rm -f

go build -trimpath -o "$closure_binary" ./cmd/bearm
run_bearm version | grep -F 'Bearm '
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
