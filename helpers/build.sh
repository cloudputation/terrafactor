#!/bin/bash

set -e

mkdir -p build
rm -f build/terrafactor
go build -o build/terrafactor .
chmod +x build/terrafactor
echo "✓ Binary built: build/terrafactor"
