#!/bin/bash
source $(dirname $0)/scripts/env.sh

FORK=$1
if [ -z $FORK ]; then
    echo "Missing FORK: ['fork9', 'fork12']"
    exit 1
fi

DATA_AVAILABILITY_MODE=$2
if [ -z $DATA_AVAILABILITY_MODE ]; then
    echo "Missing DATA_AVAILABILITY_MODE: ['rollup', 'cdk-validium', 'pessimistic']"
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
kurtosis_config_file="combinations/$FORK-$DATA_AVAILABILITY_MODE.yml"
if [ $DATA_AVAILABILITY_MODE == "pessimistic" ]; then
    if [ ! -z $agglayer_prover_sp1_key ]; then
        local_config_file=$(mktemp --suffix ".yml")
        cp "$kurtosis_config_file" $local_config_file
        echo "Setting agglayer_prover_sp1_key"
        echo "  agglayer_prover_sp1_key: $agglayer_prover_sp1_key" >> $local_config_file
        kurtosis_config_file=$local_config_file 
    fi
fi
kurtosis run --enclave cdk --args-file "$kurtosis_config_file" --image-download always $KURTOSIS_FOLDER
[ !-z $local_config_file ] && rm $local_config_file