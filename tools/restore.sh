#!/usr/bin/env bash
# Database restore script for Padduck
# Usage: ./scripts/restore.sh <backup_file>
# Env vars: DATABASE_URL (required), e.g. postgres://user:pass@host:5432/padduck
set -euo pipefail

if [ $# -lt 1 ]; then
  echo "Usage: $0 <backup_file>"
  echo "  backup_file: path to a .sql or .sql.gz backup created by backup.sh"
  exit 1
fi

BACKUP_FILE="$1"
: "${DATABASE_URL:?DATABASE_URL must be set (e.g. postgres://user:pass@host:5432/padduck)}"

# shellcheck source=tools/lib/db_url.sh
source "$(dirname "${BASH_SOURCE[0]}")/lib/db_url.sh"
strip_url_password # sets SAFE_DATABASE_URL and exports PGPASSWORD

if [ ! -f "${BACKUP_FILE}" ]; then
  echo "[restore] Error: file not found: ${BACKUP_FILE}"
  exit 1
fi

echo "[restore] WARNING: This will overwrite the current database!"
echo "[restore] Source: ${BACKUP_FILE}"
read -r -p "Type 'yes' to continue: " confirm
if [ "${confirm}" != "yes" ]; then
  echo "[restore] Aborted."
  exit 1
fi

echo "[restore] Restoring from ${BACKUP_FILE}..."

if [[ "${BACKUP_FILE}" == *.gz ]]; then
  gunzip -c "${BACKUP_FILE}" | psql "${SAFE_DATABASE_URL}"
else
  psql "${SAFE_DATABASE_URL}" < "${BACKUP_FILE}"
fi

echo "[restore] Done."
