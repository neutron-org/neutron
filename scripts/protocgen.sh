#!/usr/bin/env bash

set -eo pipefail

echo "Generating gogo proto code"
cd proto

buf generate --template buf.gen.gogo.yml $file

cd ..

# move proto files to the right places
cp -r github.com/neutron-org/neutron/x/* x/
rm -rf github.com