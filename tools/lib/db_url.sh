# Shared helper for backup.sh / restore.sh.
#
# strip_url_password splits the password out of $DATABASE_URL so it is passed
# to libpq via the PGPASSWORD environment variable instead of appearing in
# the pg_dump/psql argument list (argv is world-readable via ps).
#
# Sets SAFE_DATABASE_URL (the URL without the password) and exports
# PGPASSWORD when the URL contains one. URLs without a password (peer auth,
# .pgpass) pass through unchanged.

urldecode() {
  local encoded="${1//+/ }"
  printf '%b' "${encoded//%/\\x}"
}

strip_url_password() {
  if [[ "${DATABASE_URL}" =~ ^(postgres(ql)?://[^:@/]+):([^@/]*)@(.*)$ ]]; then
    PGPASSWORD="$(urldecode "${BASH_REMATCH[3]}")"
    export PGPASSWORD
    SAFE_DATABASE_URL="${BASH_REMATCH[1]}@${BASH_REMATCH[4]}"
  else
    SAFE_DATABASE_URL="${DATABASE_URL}"
  fi
}
