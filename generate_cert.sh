#!/bin/bash

# Create the certificates directory if it doesn't exist
mkdir -p certs

# Generate private key
openssl genrsa -out certs/server.key 2048

# Create a configuration file for the certificate
cat > certs/san.cnf << EOF
[req]
default_bits = 2048
prompt = no
default_md = sha256
distinguished_name = dn
x509_extensions = v3_req

[dn]
C = RU
ST = Moscow
L = Moscow
O = YandexAdv
CN = localhost

[v3_req]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
IP.1 = 127.0.0.1
EOF

# Generate self-signed certificate with SAN
openssl req -new -x509 -sha256 -key certs/server.key -out certs/server.crt -days 3650 -config certs/san.cnf

# Set appropriate permissions
chmod 600 certs/server.key
chmod 644 certs/server.crt

echo "Certificates generated in certs directory"
