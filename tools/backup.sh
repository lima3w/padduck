#!/usr/bin/env bash
# Database backup script for IPAM Next
# Usage: ./scripts/backup.sh [output_dir]
# Env vars: DATABASE_URL (default: postgres://ipam:ipam@localhost:5432/ipam)
set -euo pipefail

DATABASE_URL="${DATABASE_URL:-postgres://ipam:ipam@localhost:5432/ipam}"
OUTPUT_DIR="${1:-./backups}"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILE="${OUTPUT_DIR}/ipam_backup_${TIMESTAMP}.sql.gz"

mkdir -p "${OUTPUT_DIR}"

echo "[backup] Starting backup to ${BACKUP_FILE}"

pg_dump "${DATABASE_URL}" | gzip > "${BACKUP_FILE}"

SIZE=$(du -h "${BACKUP_FILE}" | cut -f1)
echo "[backup] Done: ${BACKUP_FILE} (${SIZE})"

# Prune backups older than 30 days
find "${OUTPUT_DIR}" -name "ipam_backup_*.sql.gz" -mtime +30 -delete && \
  echo "[backup] Pruned backups older than 30 days" || true
