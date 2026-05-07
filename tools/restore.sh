#!/usr/bin/env bash
# Database restore script for IPAM Next
# Usage: ./scripts/restore.sh <backup_file>
# Env vars: DATABASE_URL (default: postgres://ipam:ipam@localhost:5432/ipam)
set -euo pipefail

if [ $# -lt 1 ]; then
  echo "Usage: $0 <backup_file>"
  echo "  backup_file: path to a .sql or .sql.gz backup created by backup.sh"
  exit 1
fi

BACKUP_FILE="$1"
DATABASE_URL="${DATABASE_URL:-postgres://ipam:ipam@localhost:5432/ipam}"

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
  gunzip -c "${BACKUP_FILE}" | psql "${DATABASE_URL}"
else
  psql "${DATABASE_URL}" < "${BACKUP_FILE}"
fi

echo "[restore] Done."
