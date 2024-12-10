#!/bin/bash
source $(dirname $0)/scripts/env.sh

function ok_or_fatal(){
    if [ $? -ne 0 ]; then
        log_fatal $*
    fi
}

function build_docker_if_required(){
    docker images -q cdk:latest > /dev/null
    if [ $? -ne 0 ] ; then
        echo "Building cdk:latest"
        pushd $BASE_FOLDER/..
        make build-docker
        ok_or_fatal "Failed to build docker image"
        popd
    else
        echo "docker cdk:latest already exists"
    fi
}

function resolve_template(){
    local _TEMPLATE_FILE="$1"
    local _RESULT_VARNAME="$2"
    local _TEMP_FILE=$(mktemp --suffix ".yml")
    echo "rendering $_TEMPLATE_FILE to temp file $_TEMP_FILE"
    go run ../scripts/run_template.go $_TEMPLATE_FILE > $_TEMP_FILE
    ok_or_fatal "Failed to render template $_TEMPLATE_FILE"
    eval $_RESULT_VARNAME="$_TEMP_FILE"
}

###############################################################################
# MAIN
###############################################################################
BASE_FOLDER=$(dirname $0)
PP1_ORIGIN_CONFIG_FILE=combinations/fork12-pessimistic-multi.yml
PP2_ORIGIN_CONFIG_FILE=combinations/fork12-pessimistic-multi-attach-second-cdk.yml
KURTOSIS_ENCLAVE=pp

[ -z  $KURTOSIS_FOLDER ] &&  echo "KURTOSIS_FOLDER is not set" && exit 1
[ ! -d $KURTOSIS_FOLDER ] &&  echo "KURTOSIS_FOLDER is not a directory ($KURTOSIS_FOLDER)" && exit 1


[ ! -f $PP1_ORIGIN_CONFIG_FILE ] && echo "File $PP1_ORIGIN_CONFIG_FILE does not exist" && exit 1
[ ! -f $PP2_ORIGIN_CONFIG_FILE ] && echo "File $PP2_ORIGIN_CONFIG_FILE does not exist" && exit 1

build_docker_if_required
resolve_template $PP1_ORIGIN_CONFIG_FILE PP1_RENDERED_CONFIG_FILE
resolve_template $PP2_ORIGIN_CONFIG_FILE PP2_RENDERED_CONFIG_FILE

kurtosis clean --all
kurtosis run --enclave $KURTOSIS_ENCLAVE --args-file "$PP1_RENDERED_CONFIG_FILE" --image-download always $KURTOSIS_FOLDER
ok_or_fatal "Failed to run kurtosis pp1"

kurtosis run --enclave $KURTOSIS_ENCLAVE --args-file "$PP2_RENDERED_CONFIG_FILE" --image-download always $KURTOSIS_FOLDER
ok_or_fatal "Failed to run kurtosis attached second cdk"