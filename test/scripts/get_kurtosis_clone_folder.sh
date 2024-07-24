#!/bin/bash
source $(dirname $0)/env.sh

KURTOSIS_VERSION=$1
if [ -z $KURTOSIS_VERSION ]; then
    echo "KURTOSIS_VERSION is not set. Must be set on file env.sh"
    exit 1
fi
KURTOSIS_VERSION_FOLDER_NAME=$(echo $KURTOSIS_VERSION |  tr -c '[:alnum:]._-' '_')
DEST_FOLDER=$TMP_CDK_FOLDER/kurtosis-cdk-$KURTOSIS_VERSION_FOLDER_NAME
echo $DEST_FOLDER