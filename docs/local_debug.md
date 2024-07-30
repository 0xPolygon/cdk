# Setup environment to local debug on VSCode

## Clone kurtosis
Set KURTOSIS_FOLDER to the folder where do you want to clone `kurtosis_cdk`
```
KURTOSIS_FOLDER="/tmp/kurtosis/" ./test/scripts/clone_kurtosis.sh develop $KURTOSIS_FOLDER
```

## Run kurtosis
Set KURTOSIS_FOLDER to the folder where do you clone `kurtosis_cdk`
```
KURTOSIS_FOLDER="/tmp/kurtosis/" kurtosis run --enclave cdk-v1 --args-file  $KURTOSIS_FOLDER/cdk-erigon-sequencer-params.yml --image-download always $KURTOSIS_FOLDER
```


## Create configuration for this kurtosis environment
./test/scripts/config_kurtosis_for_local_run.sh


## Stop sequence-sender started by Kurtosis
```
kurtosis service stop cdk-v1 zkevm-node-sequence-sender-001
```

## Add to vscode launch.json
After execution `config_kurtosis_for_local_run.sh` it suggest a entry for `launch.json` configurations

