#!/bin/bash
ENCLAVE=cdk-v1
DEST=/tmp/seq_sender

build_vars_file(){
    cat > $DEST/vars.json <<EOF
    {
    "dac_port": $dac_port,
    "l1_rpc_port": $l1_rpc_port,
    "zkevm_data_streamer_port": $zkevm_data_streamer_port,
    "zkevm_l2_sequencer_address": $zkevm_l2_sequencer_address,
    "zkevm_l2_keystore_password": $zkevm_l2_keystore_password
    }
EOF
}

[ ! -d ${DEST} ] && mkdir -p ${DEST}
rm $DEST/*
kurtosis files download cdk-v1 genesis $DEST
kurtosis files download cdk-v1 sequencer-keystore $DEST

export dac_port=$(kurtosis enclave inspect $ENCLAVE --full-uuids   | grep zkevm-dac | cut -f 4 -d ':' | cut -f 1 -d ' ')
export l1_rpc_port=$(kurtosis enclave inspect $ENCLAVE --full-uuids | grep -A6 el-1-geth-lighthouse | grep " rpc:" | cut -f 4 -d ':'| cut -f 1 -d ' ')
export zkevm_data_streamer_port=$(kurtosis enclave inspect $ENCLAVE | grep "data-streamer:" | cut -f 4 -d ':' | cut -f 1 -d ' ')
#zkevm_l2_keystore_password=$(cat $DEST/node-config.toml  | grep -A3 SequenceSender.PrivateKey | grep Password | cut -f2 -d '=')

kurtosis files download $ENCLAVE zkevm-sequence-sender-config-artifact $DEST
export zkevm_l2_sequencer_address=$(cat $DEST/config.toml  |grep L2Coinbase | cut -f 2 -d "="| tr -d '"'  | tr -d ' ')
export zkevm_l2_keystore_password=$(cat $DEST/config.toml  |grep -A1 L2Coinbase  | tr ',' '\n' | grep Password | cut -f 2 -d '=' | tr -d '}' | tr -d '"' | tr -d ' ')

#build_vars_file

envsubst < test/config/test.kurtosis_template.toml > $DEST/test.kurtosis.toml
#kurtosis files rendertemplate $ENCLAVE  config/test.kurtosis_template.toml $DEST/vars.json config/test.kurtosis.toml