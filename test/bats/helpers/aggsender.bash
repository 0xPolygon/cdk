#!/usr/bin/env bash

function wait_to_settled_certificate_containing_global_index(){
    local _l2_pp1_cdk_node_url=$1
    local _global_index=$2
    local _check_frequency=${3:-30}
    local _timeout=${4:-300}
    echo "... waiting for certificate with global index $_global_index" >&3
    run_with_timeout "settle cert for $_global_index" $_check_frequency  $_timeout $aggsender_find_imported_bridge $_l2_pp1_cdk_node_url $_global_index
}