#!/usr/bin/env bash

set -eo pipefail

proto_dirs=$(find ./proto ./tmp_deps -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do

  # generate swagger files (filter query files)
  query_file=$(find "${dir}" -maxdepth 1 \( -name 'query.proto' -o -name 'service.proto' \))
  if [[ ! -z "$query_file" ]]; then
    buf generate --template proto/buf.gen.swagger.yaml $query_file
  fi
done

# Fix circular definition in cosmos b just removing them
jq 'del(.definitions["cosmos.tx.v1beta1.ModeInfo.Multi"].properties.mode_infos.items["$ref"])' ./tmp-swagger-gen/cosmos/tx/v1beta1/service.swagger.json > ./tmp-swagger-gen/cosmos/tx/v1beta1/fixed_service.swagger.json
jq 'del(.definitions["cosmos.autocli.v1.ServiceCommandDescriptor"].properties.sub_commands)' ./tmp-swagger-gen/cosmos/autocli/v1/query.swagger.json > ./tmp-swagger-gen/cosmos/autocli/v1/fixed_query.swagger.json

# merges all the above into FINAL.json
python3 ./scripts/swagger_merger.py

# Makes a swagger temp file with reference pointers
swagger-combine ./tmp-swagger-gen/FINAL.json -o ./tmp-swagger-gen/tmp_swagger.yaml -f yaml --continueOnConflictingPaths true --includeDefinitions true

# extends out the *ref instances to their full value
swagger-merger --input ./tmp-swagger-gen/tmp_swagger.yaml -o ./docs/static/swagger.yaml