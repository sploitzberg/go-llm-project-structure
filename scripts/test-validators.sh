#!/usr/bin/env bash
# Run regression tests for repository validation scripts.

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

failures=0
tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT INT TERM

pass() {
    echo "PASS: $*"
}

fail() {
    echo "FAIL: $*" >&2
    failures=$((failures + 1))
}

expect_failure_containing() {
    local description="$1"
    local expected="$2"
    shift 2

    local output
    local status
    set +e
    output=$("$@" 2>&1)
    status=$?
    set -e

    if [[ $status -ne 0 && "$output" == *"$expected"* ]]; then
        pass "$description"
    else
        fail "$description (status=$status, expected output containing: $expected)"
        printf '%s\n' "$output" >&2
    fi
}

echo "==> Testing architecture validator"
if go test ./scripts/ci/hexarch; then
    pass "architecture dependency matrix"
else
    fail "architecture dependency matrix"
fi

echo
echo "==> Testing file-quality validator"
mkdir -p "$tmpdir/scripts/ci/pre-commit"
cp scripts/ci/pre-commit/12-file-quality.sh "$tmpdir/scripts/ci/pre-commit/12-file-quality.sh"
chmod +x "$tmpdir/scripts/ci/pre-commit/12-file-quality.sh"

printf 'package fixture\n\nvar value = 1 \n' > "$tmpdir/trailing.go"
expect_failure_containing \
    "trailing whitespace is rejected" \
    "trailing whitespace" \
    "$tmpdir/scripts/ci/pre-commit/12-file-quality.sh"
rm "$tmpdir/trailing.go"

cat > "$tmpdir/conflict.go" <<'EOF'
package fixture

<<<<<<< HEAD
var value = 1
=======
var value = 2
>>>>>>> branch
EOF
expect_failure_containing \
    "merge conflict markers are rejected" \
    "merge conflict markers" \
    "$tmpdir/scripts/ci/pre-commit/12-file-quality.sh"
rm "$tmpdir/conflict.go"

printf 'package fixture\n\nvar value = 1\n' > "$tmpdir/valid.go"
if output=$("$tmpdir/scripts/ci/pre-commit/12-file-quality.sh" 2>&1) && [[ "$output" == *"File quality: OK"* ]]; then
    pass "valid files are accepted"
else
    fail "valid files are accepted"
    printf '%s\n' "${output:-}" >&2
fi

echo
echo "==> Testing staged-snapshot pre-commit hook"
hook_fixture="$tmpdir/hook-fixture"
mkdir -p "$hook_fixture/.githooks" "$hook_fixture/scripts/ci/pre-commit"
cp .githooks/pre-commit "$hook_fixture/.githooks/pre-commit"

cat > "$hook_fixture/validator.sh" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
grep -qx "$EXPECTED_CONTENT" state.txt
EOF
chmod +x "$hook_fixture/validator.sh"

for script in \
    02-golangci-lint.sh \
    03-tests.sh \
    04-hex-arch-guardrail.sh \
    07-secrets.sh \
    12-file-quality.sh \
    13-interface-impl.sh \
    14-exported-symbols.sh \
    17-struct-fields.sh; do
    cp "$hook_fixture/validator.sh" "$hook_fixture/scripts/ci/pre-commit/$script"
done

printf 'staged\n' > "$hook_fixture/state.txt"
git -C "$hook_fixture" init --quiet
git -C "$hook_fixture" add .
printf 'working tree\n' > "$hook_fixture/state.txt"

if EXPECTED_CONTENT=staged bash "$hook_fixture/.githooks/pre-commit" && grep -qx 'working tree' "$hook_fixture/state.txt"; then
    pass "pre-commit validates the index without modifying the working tree"
else
    fail "pre-commit did not validate the staged snapshot"
fi

echo
if ((failures > 0)); then
    echo "error: validator regression tests failed with $failures failure(s)" >&2
    exit 1
fi

echo "success: all validator regression tests passed"
