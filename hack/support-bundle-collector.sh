#!/bin/bash -x

HOST_PATH=${SUPPORT_BUNDLE_HOST_PATH:-/}
OUTPUT_DIR=${SUPPORT_BUNDLE_CACHE_PATH:-/tmp/support-bundle}

if [ -z "$SUPPORT_BUNDLE_MANAGER_URL" ]; then
    echo "Environment variable SUPPORT_BUNDLE_MANAGER_URL is not defined"
    exit 1
fi

[ ! -e ${OUTPUT_DIR} ] && mkdir -p $OUTPUT_DIR

NODE_NAME=${SUPPORT_BUNDLE_NODE_NAME:-$(cat ${HOST_PATH}/etc/hostname)}
BUNDLE_DIR="${OUTPUT_DIR}/${NODE_NAME}"

mkdir -p ${BUNDLE_DIR}

if [ -n "$SUPPORT_BUNDLE_COLLECTOR" ]; then
    OS_COLLECTOR="collector-$SUPPORT_BUNDLE_COLLECTOR"
else
    OS_ID=$(bash -c "source ${HOST_PATH}/etc/os-release && echo \$ID")
    if [ -z "$OS_ID" ]; then
        echo "Unable to determine OS ID"
        exit 1
    fi
    
    OS_COLLECTOR="collector-$OS_ID"
fi
echo "OS_COLLECTOR="${OS_COLLECTOR}

if [ -x "$(which $OS_COLLECTOR)" ]; then
    $OS_COLLECTOR $HOST_PATH $BUNDLE_DIR
else
    echo "No OS collector found"
fi

cd ${OUTPUT_DIR}
zip -r node_bundle.zip $(basename ${BUNDLE_DIR})
rm -rf bundle

set -o errexit
curl -v -i -H "Content-Type: application/zip" --data-binary @node_bundle.zip "${SUPPORT_BUNDLE_MANAGER_URL}/nodes/${NODE_NAME}"

sleep infinity
