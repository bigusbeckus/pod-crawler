#!/usr/bin/env bash

set -euo pipefail

# Check go fmt
echo "Checking that code complies with go fmt requirements..."
gofmt_files=$(go fmt ./...)
if [[ -n ${gofmt_files} ]]; then
	echo "Go format check failed"
	echo '    gofmt needs running on the following files:'
	echo "    ${gofmt_files}"
	echo "    You can use the command: \`go fmt\` to reformat code."
	exit 1
fi

echo "Done"
exit 0
