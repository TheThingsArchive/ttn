#!/bin/bash

set -e

TTNROOT=${PWD}
RELEASEPATH=${RELEASEPATH:-$TTNROOT/release}

mkdir -p $RELEASEPATH

git_commit=$(git rev-parse HEAD)

build_release()
{
    component=$1

    export CGO_ENABLED=0
    export GOOS=$2
    export GOARCH=$3

    release_name=$component-$GOOS-$GOARCH

    if [ "$GOOS" == "windows" ]
    then
        ext=".exe"
    else
        ext=""
    fi

    binary_name=$release_name$ext

    build_date=$(date -u +%Y-%m-%dT%H:%M:%SZ)

    echo "$build_date - Building $component for $GOOS/$GOARCH..."

    # Build
    cd $TTNROOT

    go build -a -installsuffix cgo -ldflags "-w -X main.gitCommit=$git_commit -X main.buildDate=$build_date" -o $RELEASEPATH/$binary_name ./integration/$component/main.go

    # Compress
    cd $RELEASEPATH

    echo -n "                     - Compressing tar.gz...    "
    tar -czf $release_name.tar.gz $binary_name
    echo " Done"

    echo -n "                     - Compressing tar.xz...    "
    tar -cJf $release_name.tar.xz $binary_name
    echo " Done"

    echo -n "                     - Compressing zip...       "
    zip -q $release_name.zip $binary_name > /dev/null
    echo " Done"

    # Delete Binary
    rm $binary_name
}

build_release router darwin 386
build_release router darwin amd64
build_release router linux 386
build_release router linux amd64
build_release router linux arm
build_release router windows 386
build_release router windows amd64

build_release broker darwin 386
build_release broker darwin amd64
build_release broker linux 386
build_release broker linux amd64
build_release broker linux arm
build_release broker windows 386
build_release broker windows amd64

# Prepare Releases
cd $RELEASEPATH

# Commit Release
if [ "$CI_COMMIT" != "" ]
then
  echo "Copying files for commit $CI_COMMIT"
  mkdir -p commit/$CI_COMMIT
  cp ./router* commit/$CI_COMMIT/
  cp ./broker* commit/$CI_COMMIT/
fi

# Branch Release
if [ "$CI_BRANCH" != "" ]
then
  echo "Copying files for branch $CI_BRANCH"
  mkdir -p branch/$CI_BRANCH
  cp ./router* branch/$CI_BRANCH/
  cp ./broker* branch/$CI_BRANCH/
fi

# Tag Release
if [ "$CI_TAG" != "" ]
then
  echo "Copying files for tag $CI_TAG"
  mkdir -p tag/$CI_TAG
  cp ./router* tag/$CI_TAG/
  cp ./broker* tag/$CI_TAG/
fi

# Remove Build Files
rm -f ./router*
rm -f ./broker*
