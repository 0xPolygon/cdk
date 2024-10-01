#!/bin/bash
source $(dirname $0)/scripts/env.sh
FORK=elderberry
DATA_AVAILABILITY_MODE=$1
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

$BASE_FOLDER/scripts/kurtosis_prepare_params_yml.sh "$KURTOSIS_FOLDER" $DATA_AVAILABILITY_MODE
[ $? -ne 0 ] && echo "Error preparing params.yml" && exit 1

kurtosis clean --all
echo "Override cdk config file"
cp $BASE_FOLDER/config/kurtosis-cdk-node-config.toml.template $KURTOSIS_FOLDER/templates/trusted-node/cdk-node-config.toml
kurtosis run --enclave cdk-v1 --args-file $DEST_KURTOSIS_PARAMS_YML --image-download always $KURTOSIS_FOLDER
