#! /usr/bin/env sh

docker build -t broker -f Broker_Dockerfile .
docker build -t router -f Router_Dockerfile .

echo "\n=== STARTING BROKER ===\n"
docker run broker &
sleep 2 && BROKER=$(docker inspect $(docker ps -q) | grep '"IPAddress"' | head -n 1 | grep -oE "([0-9]+\.?){4}")

echo "\n=== STARTING ROUTER ===\n"
docker run -e BROKERS=$BROKER router &
unset BROKER
