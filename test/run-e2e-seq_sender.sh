#!/bin/bash
source $(dirname $0)/scripts/env.sh
FORK=elderberry
DATA_AVAILABILITY_MODE=$1
if [ -z $DATA_AVAILABILITY_MODE ]; then
    echo "Missing DATA_AVAILABILITY_MODE: ['rollup', 'cdk-validium']"
    exit 1
fi
KURTOSIS_VERSION=jesteban/cdk-seq_sender
BASE_FOLDER=$(dirname $0)
KURTOSIS_FOLDER=$($BASE_FOLDER/scripts/get_kurtosis_clone_folder.sh $KURTOSIS_VERSION)
[ $? -ne 0 ] && echo "Error getting kurtosis folder" && exit 1

$BASE_FOLDER/scripts/clone_kurtosis.sh $KURTOSIS_VERSION "$KURTOSIS_FOLDER"
[ $? -ne 0 ] && echo "Error cloning kurtosis " && exit 1

docker images -q cdk:latest > /dev/null
if [ $? -ne 0 ] ; then
    echo "Building cdk:latest"
    pushd $BASE_FOLDER/..
    make build-docker
    popd
else
    echo "docker cdk:latest already exists"
fi

$BASE_FOLDER/scripts/kurtosis_prepare_params_yml.sh "$KURTOSIS_FOLDER" "elderberry" "cdk-validium"
[ $? -ne 0 ] && echo "Error preparing params.yml" && exit 1

kurtosis clean --all
kurtosis run --enclave cdk-v1 --args-file $DEST_KURTOSIS_PARAMS_YML --image-download always $KURTOSIS_FOLDER
#[ $? -ne 0 ] && echo "Error running kurtosis" && exit 1
echo "Waiting 10 minutes to get some verified batch...."
$KURTOSIS_FOLDER/.github/actions/monitor-cdk-verified-batches/batch_verification_monitor.sh 1 600