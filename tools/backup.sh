#!/usr/bin/env bash
# Database backup script for Padduck
# Usage: ./scripts/backup.sh [output_dir]
# Env vars: DATABASE_URL (required), e.g. postgres://user:pass@host:5432/padduck
set -euo pipefail

# Dumps contain password hashes and encrypted credentials — keep them
# owner-only (files 600, directories 700).
umask 077

: "${DATABASE_URL:?DATABASE_URL must be set (e.g. postgres://user:pass@host:5432/padduck)}"
OUTPUT_DIR="${1:-./backups}"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILE="${OUTPUT_DIR}/padduck_backup_${TIMESTAMP}.sql.gz"

# shellcheck source=tools/lib/db_url.sh
source "$(dirname "${BASH_SOURCE[0]}")/lib/db_url.sh"
strip_url_password # sets SAFE_DATABASE_URL and exports PGPASSWORD

mkdir -p "${OUTPUT_DIR}"
chmod 700 "${OUTPUT_DIR}"

echo "[backup] Starting backup to ${BACKUP_FILE}"

pg_dump "${SAFE_DATABASE_URL}" | gzip > "${BACKUP_FILE}"

SIZE=$(du -h "${BACKUP_FILE}" | cut -f1)
echo "[backup] Done: ${BACKUP_FILE} (${SIZE})"

# Prune backups older than 30 days
find "${OUTPUT_DIR}" -name "padduck_backup_*.sql.gz" -mtime +30 -delete && \
  echo "[backup] Pruned backups older than 30 days" || true
