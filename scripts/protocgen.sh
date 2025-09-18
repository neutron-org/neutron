#!/usr/bin/env bash

set -eo pipefail

protoc_install_proto_gen_doc() {
  echo "Installing protobuf protoc-gen-doc plugin"
  go install github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@latest 2>/dev/null
}

echo "Generating gogo proto code"
cd proto || exit 1
find . -name '*.proto' -print0 | while IFS= read -r -d '' file; do
  echo "$file"
  if grep -q "option go_package" "$file"; then
    buf generate --template buf.gen.gogo.yml "$file"
  fi
done

#protoc_install_proto_gen_doc
#
#echo "Generating proto docs"
#buf generate --template buf.gen.doc.yml

cd ..

# move proto files to the right places
mkdir -p x
cp -r github.com/neutron-org/neutron/v8/x/* x/
rm -rf github.com/neutron-org
