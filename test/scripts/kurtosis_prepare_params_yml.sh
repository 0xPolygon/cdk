#!/bin/bash
source $(dirname $0)/env.sh

if [ -z $DEST_KURTOSIS_PARAMS_YML ]; then
    echo "DEST_KURTOSIS_PARAMS_YML is not set. Must be set on file env.sh"
    exit 1
fi

# Check if the destination params file exists and don't do nothing
if [ -f $DEST_KURTOSIS_PARAMS_YML ]; then
    echo "Destination params file already exists"
    exit 0
fi

KURTOSIS_FOLDER=$1
if [ -z $KURTOSIS_FOLDER ]; then
    echo "Missing param Kurtosis Folder"
    exit 1
fi

DATA_AVAILABILITY_MODE=$2
if [ -z $DATA_AVAILABILITY_MODE ]; then
    echo "Missing param Data Availability Mode : [rollup, cdk-validium]"
    exit 1
fi

mkdir -p $(dirname $DEST_KURTOSIS_PARAMS_YML)
cp $KURTOSIS_FOLDER/params.yml $DEST_KURTOSIS_PARAMS_YML
yq -Y --in-place ".args.cdk_node_image = \"cdk\"" $DEST_KURTOSIS_PARAMS_YML
yq -Y --in-place ".args.data_availability_mode = \"$DATA_AVAILABILITY_MODE\"" $DEST_KURTOSIS_PARAMS_YML
