#!/bin/sh
set -e
  
# host="$1"
# shift
cmd="$@"

# TODO TIMEOUT!!

until $(curl --output /dev/null --silent --head --fail http://registry-test:8081/config); do
  >&2 echo "Registry is unavailable - sleeping"
  sleep 1
done
  
>&2 echo "Registry is up - executing command"
exec $cmd