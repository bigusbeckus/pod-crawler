#!/usr/bin/env bash

set -euo pipefail

source scripts/lib/num.sh

TITLE_PREFIX="==>"
DEFAULT_INDENT_SPACES_COUNT=${#TITLE_PREFIX}
TITLE_PREFIX_PADDING_CHAR="="

ui_print_content() {
	if [[ $# > 4 ]]; then
		echo "ERROR: Too many arguments. Max allowed: 4, Provided: $#"
		return 1
	fi

	if [[ $# == 0 ]]; then
		echo
		return 0
	fi

	local __content=$1
	local __indentlevel=0         # $2
	local __prefix=""             # $3
	local __prefixpaddingchar=" " # $4

	# Get indentlevel argument
	if [[ $# -ge 2 ]]; then
		if ! is_number $2; then
			echo "ERROR: Indent level must be a number"
			return 1
		else
			__indentlevel=$2
		fi
	fi

	# Get prefix argument
	if [[ $# -ge 3 ]]; then
		if ! [[ -z $3 ]]; then
			__prefix="$3"
		fi
	fi

	# Get prefix padding character argument
	if [[ $# -ge 4 ]]; then
		if ! [[ -z $4 ]]; then
			__prefixpaddingchar="$4"
		fi
	fi

	# Create line through black magic
	local __spacescount=0

	if [[ ${#__prefix} == 0 ]]; then
		__spacescount=$(($__indentlevel * $DEFAULT_INDENT_SPACES_COUNT + $__indentlevel - 1))
		if [[ $__spacescount -lt 0 ]]; then
			__spacescount=0
		fi
	else
		__spacescount=$(($__indentlevel * ${#__prefix}))
		if [[ $__spacescount -gt 0 ]]; then
			__spacescount=$(($__spacescount + $__indentlevel + ${#__prefix}))
		fi
	fi

	local __linelength=$(($__spacescount + ${#__content}))

	__prefix="$(printf "%${__spacescount}s" "$__prefix")"
	__prefix=${__prefix// /$__prefixpaddingchar}
	if ! [[ -z $__prefix ]]; then
		__content="$__prefix $__content"
	fi

	local __padded_content=$(printf "%${__linelength}s" "$__content")

	echo "$__padded_content"
	return 0
}

ui_print_title() {
	if [[ $# -gt 0 ]]; then
		if [[ $# -ge 2 ]]; then
			ui_print_content "$1" $2 "$TITLE_PREFIX" "$TITLE_PREFIX_PADDING_CHAR"
		else
			ui_print_content "$1" 0 "$TITLE_PREFIX" "$TITLE_PREFIX_PADDING_CHAR"
		fi
	else
		ui_print_content "" 0 "$TITLE_PREFIX" "$TITLE_PREFIX_PADDING_CHAR"
	fi

	return 0
}
