#!/usr/bin/env bash

set -euo pipefail

VERSION=0.0.1
OUTDIR="bin"
TARGET_FILE="$OUTDIR/podcrawler"

function main {
	local args_count=$#

	if [[ $args_count == 0 ]]; then
		build
		exit 0
	fi

	while [[ $args_count -gt 0 ]]; do
		case $1 in
		-c | --clean)
			clean
			exit 0
			;;
		-h | --help | help)
			show_help
			exit 0
			;;
		-v | --version | version)
			show_version
			exit 0
			;;
		-o | --output)
			TARGET_FILE=$2
			build
			exit 0
			;;
		-* | --* | *)
			echo "Unknown option $1"
			echo
			show_help
			exit 0
			;;
		esac
	done
}

function show_version {
	echo "Podcast Crawler Build Script"
	echo "Version $VERSION"
}

function show_help {
	echo "Usage:"
	echo "./scripts/build.sh [options]"

	echo
	echo "Options:"
	echo "  (no arguments)            Build binaries"
	echo "  -c, --clean, clean        Clean up build outputs"
	echo "  -h, --help, help          Print this help message"
	echo "  -v, --version, version    Print version information"
}

function build {
	echo "Build started"

	go build -o $TARGET_FILE

	echo "Done. Outputs:"
	echo "    - $TARGET_FILE"
}

function clean {
	echo "Cleaning up build outputs..."
	echo "    This action will permanently delete the following files:"
	echo "        - $TARGET_FILE"
	echo "    Are you sure? (type 'yes' to delete build outputs or anything else to cancel): "

	read -p "    Confirm: " -r

	if [[ $REPLY == "yes" ]]; then
		if [ -d "$OUTDIR" ]; then
			echo "Deleting $OUTDIR/ and all of its contents"
			rm "$OUTDIR" -rf
			echo "Done."
		else
			echo "Folder $OUTDIR/ does not exist. Nothing to clean"
		fi
	else
		echo "Aborted"
	fi
}

main $@
