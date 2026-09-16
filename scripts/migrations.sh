#!/usr/bin/env bash
set -e

if [ "$#" -lt 3 ]; then
  echo "Usage: $0 <migration-path> <database-url> <up|down|create|force> [argument]"
  exit 1
fi

migrations_path="$1"
database_url="$2"
command="$3"
shift 3

case "$command" in
  up)
    migrate -path "$migrations_path" -database "$database_url" up
    ;;
  down)
    count="${1:-1}"
    echo "Rolling back ${count} migration(s). Continue? [y/N]"
    read -r confirm
    if [ "$confirm" = "y" ]; then
      migrate -path "$migrations_path" -database "$database_url" down "$count"
    fi
    ;;
  create)
    migrate create -ext sql -dir "$migrations_path" -seq "${1:-}"
    ;;
  force)
    migrate -path "$migrations_path" -database "$database_url" force "${1:-}"
    ;;
  *)
    echo "Unknown command: $command"
    exit 1
    ;;
esac
