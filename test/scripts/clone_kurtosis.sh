#!/bin/bash
source $(dirname $0)/env.sh

usage() {
    echo "Usage: $0 <kurtosis_version> <dest_folder>"
    echo "Clones kurtosis-cdk version <kurtosis_version> to folder <dest_folder>"
}

KURTOSIS_CDK_URL="git@github.com:0xPolygon/kurtosis-cdk.git"
KURTOSIS_VERSION=$1
if [ -z $KURTOSIS_VERSION ]; then
    echo "Missing param KURTOSIS_VERSION"
    usage
    exit 1
fi
DEST_FOLDER="$2"
if [ -z $DEST_FOLDER ]; then
    echo "Missing param Destination Folder"
    usage
    exit 1
fi
if [ -d $DEST_FOLDER ]; then
    echo "Folder $DEST_FOLDER already exists. No cloning needed"
    pushd $DEST_FOLDER > /dev/null
    echo "Pulling latest changes"
    git pull
    popd > /dev/null
    exit 0
fi
echo "Cloning kurtosis-cdk version $KURTOSIS_VERSION to folder $DEST_FOLDER"
git clone --depth=1 $KURTOSIS_CDK_URL -b "$KURTOSIS_VERSION" "$DEST_FOLDER"