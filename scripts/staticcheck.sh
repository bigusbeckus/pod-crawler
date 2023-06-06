#!/usr/bin/env bash

set -euo pipefail

source ./scripts/lib/ui.sh

ui_print_title "Checking that code complies with static analysis requirements..."

packages=$(go list ./...)

go run honnef.co/go/tools/cmd/staticcheck -checks ${packages}

ui_print_content "Done" 1
