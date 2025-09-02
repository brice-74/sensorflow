#!/bin/sh
set -e

CERT_DIR=/etc/nginx/certs

mkdir -p $CERT_DIR

USER_ID="${HOST_UID:-1000}"
GROUP_ID="${HOST_GID:-1000}"

# Authority
if [ ! -f "$CERT_DIR/ca.key" ]; then
   echo "[certs] Generate CA..."
   openssl genrsa -out $CERT_DIR/ca.key 4096
   openssl req -x509 -new -nodes -key $CERT_DIR/ca.key -sha256 -days 365 \
      -subj "/CN=SensorflowCA" -out $CERT_DIR/ca.crt
   chown $USER_ID:$GROUP_ID "$CERT_DIR/ca.key" "$CERT_DIR/ca.crt"
fi

# Server
if [ ! -f "$CERT_DIR/server.key" ]; then
   echo "[certs] Generate serveur cert..."
   openssl genrsa -out $CERT_DIR/server.key 2048
   openssl req -new -key $CERT_DIR/server.key -subj "/CN=nginx.local" -out $CERT_DIR/server.csr
   openssl x509 -req -in $CERT_DIR/server.csr -CA $CERT_DIR/ca.crt -CAkey $CERT_DIR/ca.key \
      -CAcreateserial -out $CERT_DIR/server.crt -days 365 -sha256
   chown $USER_ID:$GROUP_ID "$CERT_DIR/server.key" "$CERT_DIR/server.crt" "$CERT_DIR/server.csr"
fi

echo "[entrypoint] Launching Nginx..."
exec nginx -g 'daemon off;'
