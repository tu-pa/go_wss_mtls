#!/bin/bash

while getopts n: flag
do
    case "${flag}" in
        n) ip=${OPTARG};;
    esac
done

if [ -z "$ip" ]
then
      ip="127.0.0.1"
fi

echo "IP Address: $ip";

openssl req -newkey rsa:2048 \
  -new -nodes -x509 \
  -out cert.pem \
  -keyout key.pem \
  -subj "/C=US/ST=Arizona/L=Phoenix/O=Your Orgnaization/OU=Your Unit/CN=hostname" \
  -reqexts SAN \
  -extensions SAN \
  -config <(cat /etc/ssl/openssl.cnf \
      <(printf "[SAN]\nsubjectAltName=DNS:hostname,IP:${ip}")) \
  -sha256 \
  -days 3650