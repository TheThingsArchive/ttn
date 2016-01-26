#!/bin/bash

TTNROOT=${PWD}
RELEASEPATH=${RELEASEPATH:-$TTNROOT/release}

mkdir -p $RELEASEPATH

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

    echo "Building $component for $GOOS/$GOARCH"

    # Build
    cd $TTNROOT
    go build -a -installsuffix cgo -ldflags '-w' -o $RELEASEPATH/$binary_name ./integration/$component/main.go

    # Compress
    cd $RELEASEPATH
    tar -cvzf $release_name.tar.gz $binary_name
    tar -cvJf $release_name.tar.xz $binary_name
    zip $release_name.zip $binary_name

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
  mkdir $CI_COMMIT
  cp ./router* $CI_COMMIT/
  cp ./broker* $CI_COMMIT/
fi

# Branch Release
if [ "$CI_BRANCH" != "" ]
then
  echo "Copying files for branch $CI_BRANCH"
  mkdir $CI_BRANCH
  cp ./router* $CI_BRANCH/
  cp ./broker* $CI_BRANCH/
fi

# Tag Release
if [ "$CI_TAG" != "" ]
then
  echo "Copying files for tag $CI_TAG"
  mkdir $CI_TAG
  cp ./router* $CI_TAG/
  cp ./broker* $CI_TAG/
fi
