#!/bin/bash

# SPDX-FileCopyrightText: 2025 Josef Andersson
#
# SPDX-License-Identifier: CC0-1.0

set -euo pipefail

# Create testdata directory if it doesn't exist
TESTDATA_DIR="testdata"

# Create a temporary GPG home directory
GNUPGHOME=$(mktemp -d)
export GNUPGHOME

echo "Generating test keys in $TESTDATA_DIR..."

# Generate test key pair
cat >key.config <<EOF
%echo Generating test key
Key-Type: RSA
Key-Length: 2048
Name-Real: Test User
Name-Email: test@example.com
Expire-Date: 0
%no-protection
%commit
%echo done
EOF

# Generate the key pair
gpg --batch --generate-key key.config

# Export public and private keys
gpg --armor --export test@example.com >"$TESTDATA_DIR/valid.pub"
gpg --armor --export-secret-key test@example.com >"$TESTDATA_DIR/valid.priv"
echo "Generated public key: $TESTDATA_DIR/valid.pub"
echo "Generated private key: $TESTDATA_DIR/valid.priv"

# Set correct permissions
chmod 644 "$TESTDATA_DIR/valid.pub"
chmod 600 "$TESTDATA_DIR/valid.priv"

# Cleanup
rm key.config
echo "Cleanup completed"
