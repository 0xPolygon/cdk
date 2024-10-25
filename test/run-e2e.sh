#!/bin/bash
source $(dirname $0)/scripts/env.sh

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
echo "Override cdk config file"
cp $BASE_FOLDER/config/kurtosis-cdk-node-config.toml.template $KURTOSIS_FOLDER/templates/trusted-node/cdk-node-config.toml
kurtosis run --enclave cdk --args-file "combinations/$FORK-$DATA_AVAILABILITY_MODE.yml" --image-download always $KURTOSIS_FOLDER
