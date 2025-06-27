#!/bin/bash
VERSION=$1
if [ -z "$VERSION" ]; then
    echo "Version not provided"
    exit 1
fi

cp terraform-registry-manifest.json "terraform-provider-budgeteer_${VERSION}_manifest.json"
