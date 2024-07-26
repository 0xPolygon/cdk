#!/bin/bash
source $(dirname $0)/env.sh

if [ -z $DEST_KURTOSIS_PARAMS_YML ]; then
    echo "DEST_KURTOSIS_PARAMS_YML is not set. Must be set on file env.sh"
    exit 1
fi

KURTOSIS_FOLDER=$1
if [ -z $KURTOSIS_FOLDER ]; then
    echo "Missing param Kurtosis Folder"
    exit 1
fi

FORK_NAME=$2
if [ -z $FORK_NAME ]; then
    echo "Missing param Fork Name"
    exit 1
fi
DATA_AVAILABILITY_MODE=$3
if [ -z $DATA_AVAILABILITY_MODE ]; then
    echo "Missing param Data Availability Mode : [rollup, cdk-validium]"
    exit 1
fi




cp $KURTOSIS_FOLDER/cdk-erigon-sequencer-params.yml $DEST_KURTOSIS_PARAMS_YML
yq -Y --in-place ".args.data_availability_mode = \"$DATA_AVAILABILITY_MODE\"" $DEST_KURTOSIS_PARAMS_YML
yq -Y --in-place ".args.zkevm_sequence_sender_image = \"cdk:latest\"" $DEST_KURTOSIS_PARAMS_YML
yq -Y --in-place ".args.sequencer_type = \"erigon\"" $DEST_KURTOSIS_PARAMS_YML
yq -Y --in-place ".args.deploy_cdk_erigon_node = true" $DEST_KURTOSIS_PARAMS_YML
yq -Y --in-place ".args.sequencer_type = \"erigon\"" $DEST_KURTOSIS_PARAMS_YML
yq -Y --in-place ".args.sequencer_sender_type = \"cdk\"" $DEST_KURTOSIS_PARAMS_YML