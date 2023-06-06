#!/usr/bin/env bash

set -euo pipefail

source ./scripts/lib/ui.sh

# Check go fmt
ui_print_title "Checking that code complies with go fmt requirements..."
gofmt_files=$(go fmt ./...)
if [[ -n ${gofmt_files} ]]; then
	ui_print_title "Go format check failed"
	ui_print_content 'gofmt needs running on the following files:' 1
	ui_print_content "${gofmt_files}" 2
	ui_print_content "You can use the command: \`go fmt\` to reformat code." 1
	exit 1
fi

ui_print_content "Done" 1
exit 0
