#!/bin/bash
set -euo pipefail

# Usage: wait_for_port host port [max_attempts] [sleep_seconds]
# Example: waitnet 127.0.0.1 8181 10 1

waitnet() {
   local host="${1:-localhost}"
   local port="${2:-80}"
   local max_attempts="${3:-10}"
   local sleep_seconds="${4:-1}"

   echo "[waitnet]  ⏳  Waiting for $host:$port to become available..."

   for attempt in $(seq 1 "$max_attempts"); do
      if timeout 1 bash -c ">/dev/tcp/$host/$port" 2>/dev/null; then
         echo "[waitnet] Port $port on $host is available."
         return 0
      fi
      
      sleep "$sleep_seconds"
   done

   echo "[waitnet]  ❌  Timeout: $host:$port is still not reachable after $max_attempts attempts."
   return 1
}