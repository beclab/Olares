#!/bin/bash
# wait-for-manager.sh

set -e

until test -f /healthcheck/manager-ready ; do
  >&2 echo "Manager is not ready - sleeping"
  sleep 1
done

>&2 echo "Manager is up - starting worker"

exec gosu postgres "/usr/local/bin/docker-entrypoint.sh" "postgres"
