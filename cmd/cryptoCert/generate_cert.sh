#!/bin/bash

# Exit on error
set -e

echo "Generating SSL certificates..."

# Create output directory
mkdir -p certs
cd certs

# Generate private key
openssl genpkey -algorithm RSA -out server.key -pkeyopt rsa_keygen_bits:4096

# Create config file for localhost
cat > san.cnf <<EOL
[req]
default_bits = 4096
prompt = no
default_md = sha256
distinguished_name = dn
req_extensions = req_ext

[dn]
CN = localhost

[req_ext]
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
IP.1 = 127.0.0.1
EOL

# Generate CSR and certificate
openssl req -new -key server.key -out server.csr -config san.cnf
openssl x509 -req -in server.csr -signkey server.key -out server.crt -days 365 -extfile san.cnf -extensions req_ext

# Create combined PEM file
cat server.crt server.key > server.pem

# Set secure permissions
chmod 600 server.key server.pem
chmod 644 server.crt

# Clean up
rm -f server.csr san.cnf

echo "Done! Generated files in ./certs:"
echo "  - server.key (private key)"
echo "  - server.crt (certificate)"
echo "  - server.pem (combined file)"