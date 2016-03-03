#!/bin/bash

set -e

TTNROOT=${PWD}
RELEASEPATH=${RELEASEPATH:-$TTNROOT/release}

mkdir -p $RELEASEPATH

git_commit=$(git rev-parse HEAD)

echo "Preparing build..."
echo "CI_COMMIT:  $CI_COMMIT"
echo "CI_BRANCH:  $CI_BRANCH"
echo "CI_TAG:     $CI_TAG"
echo "git_commit: $git_commit"
echo ""

build_release()
{
    export CGO_ENABLED=0
    export GOOS=$1
    export GOARCH=$2

    release_name=ttn-$GOOS-$GOARCH

    if [ "$GOOS" == "windows" ]
    then
        ext=".exe"
    else
        ext=""
    fi

    binary_name=$release_name$ext

    build_date=$(date -u +%Y-%m-%dT%H:%M:%SZ)

    echo "$build_date - Building ttn for $GOOS/$GOARCH..."

    # Build
    cd $TTNROOT

    go build -a -installsuffix cgo -ldflags "-w -X main.gitCommit=$git_commit -X main.buildDate=$build_date" -o $RELEASEPATH/$binary_name ./main.go

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

    # Delete Binary in CI build
    if [ "$CI" != "" ]
    then
      rm $binary_name
    fi
}

if [[ "$1" != "" ]] && [[ "$2" != "" ]]
then
  build_release $1 $2
else
  build_release darwin 386
  build_release darwin amd64
  build_release linux 386
  build_release linux amd64
fi

# Prepare Releases in CI build
if [ "$CI" != "" ]
then

  cd $RELEASEPATH

  # Commit Release
  if [ "$CI_COMMIT" != "" ]
  then
    echo "Copying files for commit $CI_COMMIT"
    mkdir -p commit/$CI_COMMIT
    cp ./ttn* commit/$CI_COMMIT/
  fi

  # Branch Release
  if [ "$CI_BRANCH" != "" ]
  then
    echo "Copying files for branch $CI_BRANCH"
    mkdir -p branch/$CI_BRANCH
    cp ./ttn* branch/$CI_BRANCH/
  fi

  # Tag Release
  if [ "$CI_TAG" != "" ]
  then
    echo "Copying files for tag $CI_TAG"
    mkdir -p tag/$CI_TAG
    cp ./ttn* tag/$CI_TAG/
  fi

  # Remove Build Files
  rm -f ./ttn*
fi
