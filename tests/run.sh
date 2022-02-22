#!/bin/sh

# tests are supposed to be located in the same directory as this file

DIR=$(readlink -f $(dirname $0))

export TESTING_HOST=${TESTING_HOST:="mender-reporting:8080"}
export ELASTICSEARCH_URL=${ELASTICSEARCH_URL:-"http://mender-elasticsearch:9200"}

export PYTHONDONTWRITEBYTECODE=1

pip3 install --quiet --force-reinstall -r /testing/requirements.txt

# Wait for elastic search to accept traffic
env python3 $DIR/wait_for_es.py

py.test -vv -s --tb=short --verbose \
        --junitxml=$DIR/results.xml \
        $DIR/tests "$@"
