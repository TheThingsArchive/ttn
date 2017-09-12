#!/bin/bash

env=$(dirname $0)

for service in discovery router broker networkserver handler
do
  ttn $service gen-cert --config "$env/$service/dev.yml" --key-dir "$env/$service" "localhost" "127.0.0.1" "::1"
done
