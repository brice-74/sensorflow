#!/bin/sh
set -e

CERT_DIR=/etc/nginx/certs

wait_for_file() {
   target="$1"
   max_tries="${2:-60}"
   attempt=1

   while [ ! -f "$target" ]; do
      if [ "$attempt" -gt "$max_tries" ]; then
         echo "[entrypoint] missing required file after timeout: $target"
         exit 1
      fi

      echo "[entrypoint] waiting for signer output: $target ($attempt/$max_tries)"
      attempt=$((attempt + 1))
      sleep 1
   done
}

wait_for_file "$CERT_DIR/ca.crt"
wait_for_file "$CERT_DIR/server.crt"
wait_for_file "$CERT_DIR/server.key"

echo "[entrypoint] Launching Nginx..."
exec nginx -g 'daemon off;'
