#!/bin/bash
openssl req -newkey rsa:2048 \
  -new -nodes -x509 \
  -out cert.pem \
  -keyout key.pem \
  -subj "/C=US/ST=Arizone/L=Phoenix/O=Your Organization/OU=Your Unit/CN=hostname" \
  -reqexts SAN \
  -extensions SAN \
  -config <(cat /etc/ssl/openssl.cnf \
      <(printf '[SAN]\nsubjectAltName=DNS:hostname,IP:127.0.0.1')) \
  -sha256 \
  -days 3650