#!/bin/sh
set -e

key_dir="$(dirname "${JWT_PRIVATE_KEY_PATH:-/app/keys/jwt_rsa.pem}")"
mkdir -p "$key_dir"

if [ ! -f "${JWT_PRIVATE_KEY_PATH}" ]; then
	echo "Generating JWT RSA keypair in ${key_dir}..."
	openssl genrsa -out "${JWT_PRIVATE_KEY_PATH}" 2048
	openssl rsa -in "${JWT_PRIVATE_KEY_PATH}" -pubout -out "${JWT_PUBLIC_KEY_PATH}"
fi

echo "Running database migrations..."
/app/migrate up

echo "Starting server..."
exec /app/server
