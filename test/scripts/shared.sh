
function override_cdk_node_config_file(){
    local __DATA_AVAILABILITY_MODE=$1
    if [ -z $__DATA_AVAILABILITY_MODE ]; then
        echo "Missing DATA_AVAILABILITY_MODE:override_cdk_node_config_file  ['rollup', 'cdk-validium', 'pessimistic']"
        return 1
    fi
    echo "Override cdk config file"
    local _FILENAME=$BASE_FOLDER/config/kurtosis-cdk-node-config.toml.template.$__DATA_AVAILABILITY_MODE
    cp $_FILENAME $KURTOSIS_FOLDER/templates/trusted-node/cdk-node-config.toml
    if [ $? -ne 0 ]; then
        echo "... copying generic config file"
        _FILENAME=$BASE_FOLDER/config/kurtosis-cdk-node-config.toml.template
        cp $_FILENAME $KURTOSIS_FOLDER/templates/trusted-node/cdk-node-config.toml
    fi
    echo "Override cdk config using: " $_FILENAME
}
