#!/bin/sh
# scripts/verify-commits.sh
# Verifies a range of commits against Conventional Commits v1.0.0
# Usage: ./scripts/verify-commits.sh [base_ref] [head_ref]

set -e

BASE_REF="${1:-origin/main}"
HEAD_REF="${2:-HEAD}"

# Resolve commit-msg validator path
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
VALIDATOR="$REPO_ROOT/.githooks/commit-msg"

if [ ! -f "$VALIDATOR" ]; then
    echo "Error: commit validator not found at $VALIDATOR" >&2
    exit 1
fi

# Check if base commit exists
if ! git rev-parse --verify "$BASE_REF" >/dev/null 2>&1; then
    # In shallow clones or single branch checkouts, verify the HEAD commit only
    echo "Base ref '$BASE_REF' not found. Verifying HEAD commit only..."
    COMMITS=$(git rev-parse HEAD)
else
    COMMITS=$(git rev-list "$BASE_REF..$HEAD_REF")
fi

if [ -z "$COMMITS" ]; then
    echo "No commits to verify."
    exit 0
fi

FAILED=0
TMP_MSG="$(mktemp)"
trap 'rm -f "$TMP_MSG"' EXIT INT TERM

echo "Checking commit messages against Conventional Commits..."

for COMMIT in $COMMITS; do
    git log -1 --format=%B "$COMMIT" > "$TMP_MSG"
    SUBJECT=$(git log -1 --format=%s "$COMMIT")
    
    if ! sh "$VALIDATOR" "$TMP_MSG" >/dev/null 2>&1; then
        echo "❌ Invalid commit: $COMMIT"
        echo "   Message: \"$SUBJECT\""
        FAILED=1
    else
        echo "✅ Valid: $COMMIT - \"$SUBJECT\""
    fi
done

if [ $FAILED -ne 0 ]; then
    echo "" >&2
    echo "❌ Error: One or more commits do not follow Conventional Commits." >&2
    exit 1
fi

echo "✨ All commit messages follow Conventional Commits!"
exit 0
