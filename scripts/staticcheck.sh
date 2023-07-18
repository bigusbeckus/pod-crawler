#!/usr/bin/env bash

set -euo pipefail

echo "Checking that code complies with static analysis requirements..."

packages=$(go list ./...)

go run honnef.co/go/tools/cmd/staticcheck -checks ${packages}

echo "Done"
