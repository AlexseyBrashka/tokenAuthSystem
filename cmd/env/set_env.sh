#!/bin/sh
set -a
. /app/.env
set +a
exec "$@"