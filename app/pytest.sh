#!/bin/sh

# run pytest with coverage and check the exit code of pytest
if ! coverage run -m pytest $@;
then
    echo "Tests failed"
    exit 1
fi

coverage report
