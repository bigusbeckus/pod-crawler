#!/usr/bin/env bash

is_number() {
	# if [[ $# == 0 ]]; then
	# 	return false
	# fi

	local re='^[0-9]+$'

	if [[ $1 =~ $re ]]; then
		# return 0
		true
	else
		false
	fi
}
