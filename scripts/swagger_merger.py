#!/usr/bin/env python3

# MODIFIED SCRIPT FROM JUNO: https://github.com/CosmosContracts/juno/blob/main/scripts/merge_protoc.py

import json
import os

current_dir = os.path.dirname(os.path.realpath(__file__))
project_root = os.path.dirname(current_dir)

all_dir = os.path.join(project_root, "tmp-swagger-gen", "_all")

# get the go.mod file Version
version = ""
with open(os.path.join(project_root, "go.mod"), "r") as f:
    for line in f.readlines():
        if line.startswith("module"):
            version = line.split("/")[-1].strip()
            break

if not version:
    print("Could not find version in go.mod")
    exit(1)

# What we will save when all combined
output: dict
output = {
    "swagger": "2.0",
    "info": {"title": "Neutron", "version": version},
    "consumes": ["application/json"],
    "produces": ["application/json"],
    "paths": {},
    "definitions": {},
}

# Combine all individual files calls into 1 massive file.
files = os.listdir(all_dir)
files.sort()

for file in files:
    if not file.endswith(".json"):
        continue

    # read file all_dir / file
    with open(os.path.join(all_dir, file), "r") as f:
        data = json.load(f)

    for key in data["paths"]:
        output["paths"][key] = data["paths"][key]

    for key in data["definitions"]:
        output["definitions"][key] = data["definitions"][key]


# loop through all paths, then alter any keys which are "operationId" to be a random string of 20 characters
# this is done to avoid duplicate keys in the final output (which opens 2 tabs in swagger-ui)
# current-random
for path in output["paths"]:
    for method in output["paths"][path]:
        if "operationId" in output["paths"][path][method]:
            output["paths"][path][method][
                "operationId"
            ] = f'{output["paths"][path][method]["operationId"]}'


# save output into 1 big json file
with open(
        os.path.join(project_root, "tmp-swagger-gen", "FINAL.json"), "w"
) as f:
    json.dump(output, f, indent=2, sort_keys=True)