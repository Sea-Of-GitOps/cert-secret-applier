#!/bin/bash

# Base directory structure
BASE_DIR="./etc/letsencrypt"
LIVE_DIR="$BASE_DIR/live"
ARCHIVE_DIR="$BASE_DIR/archive"
CERTS_DIR="$BASE_DIR/certs"

# Initialize directories
initialize_directories() {
    mkdir -p "$LIVE_DIR" "$ARCHIVE_DIR" "$CERTS_DIR"
    echo "Directory structure initialized."
}

# Generate key and certificate
generate_certificate() {
    DOMAIN=${DUCKDNS_DOMAIN}
    if [ -z "$DOMAIN" ]; then
        echo "Error: DUCKDNS_DOMAIN environment variable is not set."
        exit 1
    fi

    DOMAIN_DIR="$LIVE_DIR/$DOMAIN"
    mkdir -p "$DOMAIN_DIR"

    # Generate private key
    openssl genrsa -out "$DOMAIN_DIR/privkey.pem" 2048 2>/dev/null

    # Generate a self-signed certificate
    openssl req -new -x509 -key "$DOMAIN_DIR/privkey.pem" -out "$DOMAIN_DIR/fullchain.pem" \
        -days 365 -subj "/CN=$DOMAIN" 2>/dev/null

    echo "Certificate for $DOMAIN generated and saved in $DOMAIN_DIR."
}

# Run the program
initialize_directories
generate_certificate
