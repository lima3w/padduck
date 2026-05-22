#!/bin/sh
set -e
# Fix ownership of the data directory when it is bind-mounted from the host
# (Docker creates host bind-mount dirs as root). This runs as root and then
# drops to the unprivileged app user for the main process.
chown app:app /app/data
exec su-exec app "$@"
