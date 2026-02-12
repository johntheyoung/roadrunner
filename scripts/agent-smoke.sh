#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

rr_bin="$tmpdir/rr"
go build -o "$rr_bin" ./cmd/rr

run_stdout=""
run_stderr=""
run_exit=0

run_case() {
  local name="$1"
  local expected_exit="$2"
  shift 2

  local out_file err_file
  out_file="$(mktemp)"
  err_file="$(mktemp)"

  set +e
  "$@" >"$out_file" 2>"$err_file"
  run_exit=$?
  set -e

  run_stdout="$(cat "$out_file")"
  run_stderr="$(cat "$err_file")"
  rm -f "$out_file" "$err_file"

  if [[ "$run_exit" -ne "$expected_exit" ]]; then
    echo "FAIL [$name]: exit code $run_exit, expected $expected_exit" >&2
    echo "--- stdout ---" >&2
    echo "$run_stdout" >&2
    echo "--- stderr ---" >&2
    echo "$run_stderr" >&2
    exit 1
  fi
  echo "PASS [$name]"
}

assert_stdout_contains() {
  local name="$1"
  local needle="$2"
  if ! grep -Fq -- "$needle" <<<"$run_stdout"; then
    echo "FAIL [$name]: stdout missing: $needle" >&2
    echo "--- stdout ---" >&2
    echo "$run_stdout" >&2
    exit 1
  fi
}

# 1) Agent success path: version command in envelope mode.
run_case "agent-version-success" 0 "$rr_bin" --agent --request-id=req-smoke-ok --enable-commands=version version
assert_stdout_contains "agent-version-success" '"success": true'
assert_stdout_contains "agent-version-success" '"command": "version"'
assert_stdout_contains "agent-version-success" '"error-hints"'
assert_stdout_contains "agent-version-success" '"request_id": "req-smoke-ok"'

# 2) Agent mode requires allowlist.
run_case "agent-missing-allowlist" 2 "$rr_bin" --agent version
assert_stdout_contains "agent-missing-allowlist" '"success": false'
assert_stdout_contains "agent-missing-allowlist" '"code": "VALIDATION_ERROR"'
assert_stdout_contains "agent-missing-allowlist" '"hint":'

# 3) Enable-commands restriction should be deterministic and actionable.
run_case "enable-commands-restriction" 2 "$rr_bin" --json --envelope --enable-commands=chats messages list '!room:beeper.local'
assert_stdout_contains "enable-commands-restriction" '"code": "VALIDATION_ERROR"'
assert_stdout_contains "enable-commands-restriction" '"hint":'
assert_stdout_contains "enable-commands-restriction" '--enable-commands'

# 4) Readonly restriction should include a machine hint.
run_case "readonly-restriction" 2 "$rr_bin" --json --envelope --request-id=req-smoke-ro --readonly messages send '!room:beeper.local' 'hello'
assert_stdout_contains "readonly-restriction" '"code": "VALIDATION_ERROR"'
assert_stdout_contains "readonly-restriction" '"hint":'
assert_stdout_contains "readonly-restriction" '--readonly'
assert_stdout_contains "readonly-restriction" '"request_id": "req-smoke-ro"'

# 5) Connectivity errors in agent mode should expose code + hint.
run_case "connection-error-hint" 1 env BEEPER_TOKEN=test-token BEEPER_URL=http://127.0.0.1:9 "$rr_bin" --agent --enable-commands=messages messages list '!room:beeper.local'
assert_stdout_contains "connection-error-hint" '"code": "CONNECTION_ERROR"'
assert_stdout_contains "connection-error-hint" '"hint": "Run `rr doctor` to verify Desktop API connectivity and token validity."'

echo "Agent smoke checks passed."
