#!/bin/bash
source $(dirname $0)/scripts/env.sh
source $(dirname $0)/scripts/shared.sh

FORK=$1
if [ -z $FORK ]; then
    echo "Missing FORK: ['fork9', 'fork12']"
    exit 1
fi

DATA_AVAILABILITY_MODE=$2
if [ -z $DATA_AVAILABILITY_MODE ]; then
    echo "Missing DATA_AVAILABILITY_MODE: ['rollup', 'cdk-validium']"
    exit 1
fi

BASE_FOLDER=$(dirname $0)
docker images -q cdk:latest > /dev/null
if [ $? -ne 0 ] ; then
    echo "Building cdk:latest"
    pushd $BASE_FOLDER/..
    make build-docker
    popd
else
    echo "docker cdk:latest already exists"
fi

kurtosis clean --all
override_cdk_node_config_file $DATA_AVAILABILITY_MODE

KURTOSIS_CONFIG_FILE="combinations/$FORK-$DATA_AVAILABILITY_MODE.yml"
TEMP_CONFIG_FILE=$(mktemp --suffix ".yml")
echo "rendering $KURTOSIS_CONFIG_FILE to temp file $TEMP_CONFIG_FILE"
go run ../scripts/run_template.go $KURTOSIS_CONFIG_FILE > $TEMP_CONFIG_FILE
kurtosis run --enclave cdk --args-file "$TEMP_CONFIG_FILE" --image-download always $KURTOSIS_FOLDER
rm $TEMP_CONFIG_FILE