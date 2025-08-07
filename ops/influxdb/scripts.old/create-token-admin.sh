#!/bin/bash
set -euo pipefail

source "${ARGPARSER_PATH:-$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/argparser.sh}"

define_option token-path "${INFLUXDB3_ADMIN_TOKEN_PATH:-"/var/lib/influxdb/admin.token"}"
define_option addr "http://${INFLUXDB3_HTTP_BIND_ADDR:-"0.0.0.0:8181"}"

parse_args "$@"

token_path=$(get_option token-path)
addr=$(get_option addr)

if [[ ! -f "$token_path" ]]; then
   echo "🔑  Creating admin token..."
   output=$(influxdb3 create token --admin --host "$addr" 2>&1) || {
      echo "❌  Failed to create token. Output:"
      echo "$output"
      exit 1
   }

   token=$(echo "$output" | grep -oP 'apiv3_\S+' | head -n1) || true
   if [[ -z "${token:-}" ]]; then
      echo "❌  Failed to extract token from output:"
      echo "$output"
      exit 1
   fi

   echo "$token" > "$token_path"
   echo "✅  Token created and saved to $token_path"
else
   echo "✅  Token already exists at $token_path"
fi
