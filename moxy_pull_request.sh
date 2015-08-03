#!/bin/sh

## Fail if any command fails
set -e

## Export parameters
### FIXME(pht) use variables for reporting
export TAG=$(echo $BRANCH | sed 's#.*/##') #Keep the last part of the branch name for docker tags
export TAG="$(echo -n "$BRANCH" | md5sum | awk '{ print $1 }')"
# JIRA isssue (extracted from branch name)
export ISSUE=$(echo -n "$BRANCH" | sed 's:.*/\([A-Z]*-[0-9]*\).*:\1:')
# This is here to define the variable ; in practice we will use the dummy auth
export BE_URL=http://localhost:8080
# URL for the mosquitto container
export MQTT_URL=localhost:1884

# Go variables
export JENKINS=/home/jenkins
export GOROOT=$JENKINS/go
export GOPATH=$JENKINS/gocode
export PATH=$PATH:$GOPATH/bin:$GOROOT/bin

## Kill and remove existing docker containers
echo "Killing docker images"
docker kill $(docker ps -a -q) || true mosquitto
echo "Removing docker images"
docker rm $(docker ps -a -q) || true

## Run docker containers
echo "Starting mosquitto container"
docker run -d -p 1884:1883 --user=mosquitto --name=mosquitto airvantage/mosquitto

## Build moxy
echo "Building moxy"
gb build all

## Run moxy in a subshell
(
    echo "Starting moxy"
    killall moxy || true
    killall moxy-dummyauth || true
    rm -rf /tmp/auth.sock
    ./bin/moxy -auth=bin/moxy-dummyauth -t -v > /dev/null
) & moxy_pid=$!
echo "Moxy running with pid " $moxy_pid

## Run tests
echo "Running tests"
python3 ./test/interoperability/client_test.py --hostname localhost --port 1883 -z -d > tests.log
test_results=$?

## Kill moxy
kill $moxy_pid || true

## TODO(pht) format tests results
echo "Return $test_results"
exit $test_results
