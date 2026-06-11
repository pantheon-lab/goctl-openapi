#!/bin/bash
set -euo pipefail

# Generate user.openapi.yaml and web.openapi.yaml first
cd "$(dirname "$0")"
go generate ./...

for l in go javascript php; do
  echo "=== Generating $l client ==="
  rm -rf "clients/$l"
  docker run --rm -v "$(pwd):/local" openapitools/openapi-generator-cli generate \
    -i "/local/user.openapi.yaml" \
    -g "$l" \
    -o "/local/clients/$l"
done

echo "Done. See https://github.com/OpenAPITools/openapi-generator for more info"