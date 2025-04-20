#!/bin/bash

if [ -f .env ]; then
    source .env
fi

cd sql/schema

# Use GOOSE_DBSTRING if set, otherwise fall back to DB_URL
DB_CONNECTION=${GOOSE_DBSTRING:-$DB_URL}

if [ -z "$DB_CONNECTION" ]; then
    echo "Error: No database connection string found. Set either GOOSE_DBSTRING or DB_URL"
    exit 1
fi

goose postgres "$DB_CONNECTION" up
