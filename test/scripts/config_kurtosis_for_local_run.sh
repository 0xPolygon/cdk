#!/bin/bash
#Include common varaibles
source $(dirname $0)/env.sh




###############################################################################
# MAIN
###############################################################################
set -o pipefail # enable strict command pipe error detection

if [ -z $TMP_CDK_FOLDER -o -z $ENCLAVE ]; then
    echo "TMP_CDK_FOLDER or ENCLAVE is not set. Must be set on file env.sh"
    exit 1
fi
DEST=${TMP_CDK_FOLDER}/local_config

[ ! -d ${DEST} ] && mkdir -p ${DEST}
rm $DEST/*
kurtosis files download cdk-v1 genesis $DEST
[ $? -ne 0 ] && echo "Error downloading genesis" && exit 1
export genesis_file=$DEST/genesis.json
kurtosis files download cdk-v1 sequencer-keystore $DEST
[ $? -ne 0 ] && echo "Error downloading sequencer-keystore" && exit 1
export sequencer_keystore_file=$DEST/sequencer.keystore

l1_rpc_port=$(kurtosis port print $ENCLAVE el-1-geth-lighthouse rpc | cut -f 3 -d ":")
[ $? -ne 0 ] && echo "Error getting l1_rpc_port" && exit 1 || export l1_rpc_port && echo "l1_rpc_port=$l1_rpc_port"

zkevm_data_streamer_port=$(kurtosis port print $ENCLAVE cdk-erigon-sequencer-001  data-streamer | cut -f 3 -d ":")
[ $? -ne 0 ] && echo "Error getting zkevm_data_streamer_port" && exit 1 || export zkevm_data_streamer_port && echo "zkevm_data_streamer_port=$zkevm_data_streamer_port"

kurtosis files download $ENCLAVE zkevm-sequence-sender-config-artifact $DEST
export zkevm_l2_sequencer_address=$(cat $DEST/config.toml  |grep L2Coinbase | cut -f 2 -d "="| tr -d '"'  | tr -d ' ')
export zkevm_l2_keystore_password=$(cat $DEST/config.toml  |grep -A1 L2Coinbase  | tr ',' '\n' | grep Password | cut -f 2 -d '=' | tr -d '}' | tr -d '"' | tr -d ' ')
export l1_chain_id=$(cat $DEST/config.toml  |grep  L1ChainID  |  cut -f 2 -d '=')
export zkevm_is_validium=$(cat $DEST/config.toml  |grep  IsValidiumMode  |  cut -f 2 -d '=')

if [ "$zkevm_is_validium" == "true" ]; then
    dac_port=$(kurtosis port print $ENCLAVE zkevm-dac http-rpc | cut -f 3 -d ":")
    [ $? -ne 0 ] && echo "Error getting dac_port" && exit 1 || export dac_port  && echo "dac_port=$dac_port"
fi

envsubst < test/config/test.kurtosis_template.toml > $DEST/test.kurtosis.toml

echo "- start kurtosis"
echo "    kurtosis clean --all; kurtosis run --enclave cdk-v1 --args-file cdk-erigon-sequencer-params.yml --image-download always ."
echo " "
echo "- Stop sequence-sender"
echo "    kurtosis service stop cdk-v1 zkevm-node-sequence-sender-001"
echo " "
echo "- Add next configuration to vscode launch.json"
cat << EOF
         {
            "name": "run local_docker",
                "type": "go",
                "request": "launch",
                "mode": "auto",
                "program": "cmd/",
                "args":["run","-cfg","$DEST/test.kurtosis.toml",
                "--components", "sequence-sender",
                "--custom-network-file", "$DEST/local_config/genesis.json"
                ]
            },
EOF

