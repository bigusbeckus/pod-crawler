#!/usr/bin/env bash

set -euo pipefail

source scripts/lib/ui.sh

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
	ui_print_content "Usage:"
	ui_print_content "./scripts/build.sh [options]" 1

	echo
	ui_print_content "Options:"
	ui_print_content "(no arguments)            Build binaries" 1
	ui_print_content "-c, --clean, clean        Clean up build outputs" 1
	ui_print_content "-h, --help, help          Print this help message" 1
	ui_print_content "-v, --version, version    Print version information" 1
}

function build {
	ui_print_title "Build started"
	# ui_print_content "Creating binaries..." 1

	go build -o $TARGET_FILE

	ui_print_content "Done. Outputs:" 1
	ui_print_content "- $TARGET_FILE" 2
}

function clean {
	ui_print_title "Cleaning up build outputs..."
	ui_print_content "This action will permanently delete the following files:" 1
	ui_print_content "- $TARGET_FILE" 2
	ui_print_content "Are you sure? (type 'yes' to delete build outputs or anything else to cancel): " 1

	read -p "    Confirm: " -r
	# read -p "Are you sure? (y/N): " -n 1 -rs

	if [[ $REPLY == "yes" ]]; then
		if [ -d "$OUTDIR" ]; then
			ui_print_title "Deleting $OUTDIR/ and all of its contents"
			rm "$OUTDIR" -rf
			ui_print_content "Done." 1
		else
			ui_print_title "Folder $OUTDIR/ does not exist. Nothing to clean"
		fi
	else
		ui_print_title "Aborted"
	fi
}

main $@

# go build -o bin/podcrawler

# ui_print_content "Regular"
# ui_print_title "Title A"
# ui_print_content "Content A" 1
# ui_print_title "Title B" 1
# ui_print_content "Content B" 2
# ui_print_title "Title C" 2
# ui_print_content "Content C" 3
# ui_print_title "Title D" 3
# ui_print_content "Content D" 4
# ui_print_title "Title E" 4
# ui_print_content "Content E" 5
