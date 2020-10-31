#!/usr/bin/env bash

# test filename example: tests/00-http-server-ok.sh

function set_up {
	docker-compose up -d
}

function clean_up {
	docker-compose down
}

function run_test {
	test_name=$(echo $1 | cut -d'.' -f 1)
	echo "# testing $test_name..."
	if tests/$1 ; then
		echo "# $test_name success"
	else
		code=$?
		echo "# $test_name failed with code $code"
		return $code
	fi
}

set_up
for test in $(ls tests)
do
	if ! run_test $test; then
		clean_up
		exit -1
	fi
done
clean_up
