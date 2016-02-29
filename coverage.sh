#!/bin/sh

mkdir .cover
for pkg in $(go list ./... | grep -vE 'cmd|integration|ttn$')
do
    profile=".cover/$(echo $pkg | grep -oE 'ttn/.*' | sed 's/\///g').cover"
    go test -cover -coverprofile=$profile $pkg
done
echo "mode: set" > coverage.out && cat .cover/*.cover | grep -v mode: | sort -r | \
awk '{if($1 != last) {print $0;last=$1}}' >> coverage.out
rm -r .cover
