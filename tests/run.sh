#!/bin/sh

# tests are supposed to be located in the same directory as this file

DIR=$(readlink -f $(dirname $0))

export TEST_HOST=${TESTING_HOST:="mender-iot-manager:8080"}

export PYTHONDONTWRITEBYTECODE=1

pip3 install --quiet --force-reinstall -r /testing/requirements.txt

py.test -vv -s --tb=short --verbose \
        --junitxml=$DIR/results.xml \
        $DIR/tests "$@"
