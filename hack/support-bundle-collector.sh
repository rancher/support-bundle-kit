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
cd ${BUNDLE_DIR}
# get some host information
cp ${HOST_PATH}/etc/hostname .

# collect logs
mkdir -p logs
cd logs
dmesg &> dmesg.log

# k3s logs don't rorate well and can be huge
tail -c 10m ${HOST_PATH}/var/log/k3s-service.log > k3s-service.log
tail -c 10m ${HOST_PATH}/var/log/k3s-restarter.log > k3s-restarter.log

cp ${HOST_PATH}/var/log/qemu-ga.log* .
cp ${HOST_PATH}/var/log/messages* .
cp ${HOST_PATH}/var/log/console.log .

cd ${OUTPUT_DIR}
zip -r node_bundle.zip $(basename ${BUNDLE_DIR})
rm -rf bundle

set -o errexit
curl -v -i -H "Content-Type: application/zip" --data-binary @node_bundle.zip "${SUPPORT_BUNDLE_MANAGER_URL}/nodes/${NODE_NAME}"

sleep infinity
