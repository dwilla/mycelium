#!/bin/bash

# Create certs directory if it doesn't exist
mkdir -p certs

# Generate a self-signed certificate
openssl req -x509 -newkey rsa:4096 -keyout certs/key.pem -out certs/cert.pem -days 365 -nodes -subj "/CN=localhost"

echo "Self-signed certificate generated in certs directory" 