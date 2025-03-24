#!/usr/bin/env bash

set -eo pipefail

go mod tidy

# Create temporary directories for dependencies and swagger files
mkdir -p tmp_deps tmp-swagger-gen/_all

# Copy some dependencies to use their proto files to generate swagger
declare -a deps=(
  "github.com/cosmos/cosmos-sdk"
  "github.com/CosmWasm/wasmd"
  "github.com/cosmos/admin-module/v2"
  "github.com/cosmos/interchain-security/v5"
  "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8"
  "github.com/skip-mev/feemarket"
  "github.com/skip-mev/slinky"
  "github.com/skip-mev/block-sdk/v2"
)

for dep in "${deps[@]}"; do
  path=$(go list -f '{{ .Dir }}' -m "$dep")
  cp -r "$path" tmp_deps
done

# Find proto directories and generate swagger files
proto_dirs=$(find ./proto ./tmp_deps -name '*.proto' -print0 | xargs -0 -n1 dirname | sort -u)

for dir in $proto_dirs; do
  query_file=$(find "${dir}" -maxdepth 1 \( -name 'query.proto' -o -name 'service.proto' \))
  if [[ -n "$query_file" ]]; then
    buf generate --template proto/buf.gen.swagger.yaml "$query_file"
  fi
done

# Fix circular definitions in swagger files
jq 'del(.definitions["cosmos.tx.v1beta1.ModeInfo.Multi"].properties.mode_infos.items["$ref"])' ./tmp-swagger-gen/cosmos/tx/v1beta1/service.swagger.json > ./tmp-swagger-gen/cosmos/tx/v1beta1/fixed_service.swagger.json
jq 'del(.definitions["cosmos.autocli.v1.ServiceCommandDescriptor"].properties.sub_commands)' ./tmp-swagger-gen/cosmos/autocli/v1/query.swagger.json > ./tmp-swagger-gen/cosmos/autocli/v1/fixed_query.swagger.json

rm -rf ./tmp-swagger-gen/cosmos/tx/v1beta1/service.swagger.json
rm -rf ./tmp-swagger-gen/cosmos/autocli/v1/query.swagger.json

# Remove unnecessary modules and their proto files
declare -a unnecessary_modules=(
  "cosmos/staking"
  "cosmos/distribution"
  "cosmos/gov"
  "cosmos/mint"
  "cosmos/group"
  "interchain_security/ccv/provider"
)

for module in "${unnecessary_modules[@]}"; do
  rm -rf "tmp-swagger-gen/$module"
done

# Convert all *.swagger.json files into a single folder _all
counter=0
find ./tmp-swagger-gen -name '*.swagger.json' -print0 | while IFS= read -r -d '' f; do
  echo "[+] $f"
  case "$f" in
    *router*) cp "$f" "./tmp-swagger-gen/_all/pfm-$counter.json" ;;
    *cosmwasm*) cp "$f" "./tmp-swagger-gen/_all/cosmwasm-$counter.json" ;;
    *osmosis*) cp "$f" "./tmp-swagger-gen/_all/osmosis-$counter.json" ;;
    *juno*) cp "$f" "./tmp-swagger-gen/_all/juno-$counter.json" ;;
    *cosmos*) cp "$f" "./tmp-swagger-gen/_all/cosmos-$counter.json" ;;
    *) cp "$f" "./tmp-swagger-gen/_all/other-$counter.json" ;;
  esac
  ((counter++))
done

# Merge all the above into FINAL.json
python3 ./scripts/swagger_merger.py

# Combine and extend swagger references
swagger-combine ./tmp-swagger-gen/FINAL.json -o ./tmp-swagger-gen/tmp_swagger.yaml -f yaml --continueOnConflictingPaths true --includeDefinitions true
swagger-merger --input ./tmp-swagger-gen/tmp_swagger.yaml -o ./docs/static/swagger.yaml

# Clean up temporary files
rm -rf tmp-swagger-gen tmp_deps
