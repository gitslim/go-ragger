#!/bin/bash
set -e

echo "Creating database: $DB_NAME"

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
  CREATE DATABASE "$DB_NAME";
EOSQL
