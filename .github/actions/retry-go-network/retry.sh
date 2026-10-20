#!/usr/bin/env bash

set -euo pipefail

: "${RETRY_COMMAND:?Set the command to run}"
: "${RETRY_ATTEMPTS:=3}"
: "${RETRY_DELAY:=10}"

if ! [[ "$RETRY_ATTEMPTS" =~ ^[1-9][0-9]*$ && "$RETRY_DELAY" =~ ^(0|[1-9][0-9]*)$ ]]; then
  echo "::error::attempts must be positive and delay must be nonnegative integers"
  exit 2
fi

log_dir=$(mktemp -d)
trap 'rm -rf "$log_dir"' EXIT

# Match the Go proxy URL as well as a transient transport error. Compilation,
# test, checksum and missing-module failures should not be retried.
transient_error='https?://(proxy\.golang\.org|sum\.golang\.org)/.*(INTERNAL_ERROR|unexpected EOF|i/o timeout|TLS handshake timeout|connection reset by peer|502 Bad Gateway|503 Service Unavailable|504 Gateway Timeout)'

for ((attempt = 1; attempt <= RETRY_ATTEMPTS; attempt++)); do
  echo "Attempt $attempt/$RETRY_ATTEMPTS"
  if bash --noprofile --norc -eo pipefail -c "$RETRY_COMMAND" 2>&1 | tee "$log_dir/output"; then
    exit 0
  else
    statuses=("${PIPESTATUS[@]}")
  fi

  status=${statuses[0]}
  if ((statuses[1] != 0)); then
    exit "${statuses[1]}"
  fi
  if ((status == 130 || status == 143)); then
    exit "$status"
  fi
  if ! grep -Eq "$transient_error" "$log_dir/output"; then
    exit "$status"
  fi
  if ((attempt == RETRY_ATTEMPTS)); then
    echo "::error::Go download still failing after $attempt attempts"
    exit "$status"
  fi

  echo "::warning::Transient Go download failure; retrying in $RETRY_DELAY seconds"
  sleep "$RETRY_DELAY"
done
