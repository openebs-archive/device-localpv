#!/usr/bin/env bash
# Copyright 2021 The OpenEBS Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


set -e

# setup a block device and meta partition on the disk
truncate -s 1024G /tmp/disk.img
disk=`sudo losetup -f /tmp/disk.img --show`
sudo parted "$disk" mklabel gpt
sudo parted "$disk" mkpart test-device 1MiB 10MiB

DEVICE_OPERATOR=deploy/device-operator.yaml

export DEVICE_DRIVER_NAMESPACE="openebs"
export TEST_DIR="tests"
export NAMESPACE="kube-system"
export KUBECONFIG=$HOME/.kube/config

# Prepare env for running BDD tests
# Minikube is already running
kubectl apply -f $DEVICE_OPERATOR

dumpAgentLogs() {
  NR=$1
  AgentPOD=$(kubectl get pods -l app=openebs-device-node -o jsonpath='{.items[0].metadata.name}' -n "$NAMESPACE")
  kubectl describe po "$AgentPOD" -n "$NAMESPACE"
  printf "\n\n"
  kubectl logs --tail="${NR}" "$AgentPOD" -n "$NAMESPACE" -c openebs-device-plugin
  printf "\n\n"
}

dumpControllerLogs() {
  NR=$1
  ControllerPOD=$(kubectl get pods -l app=openebs-device-controller -o jsonpath='{.items[0].metadata.name}' -n "$NAMESPACE")
  kubectl describe po "$ControllerPOD" -n "$NAMESPACE"
  printf "\n\n"
  kubectl logs --tail="${NR}" "$ControllerPOD" -n "$NAMESPACE" -c openebs-device-plugin
  printf "\n\n"
}

isPodReady(){
  [ "$(kubectl get po "$1" -o 'jsonpath={.status.conditions[?(@.type=="Ready")].status}' -n "$NAMESPACE")" = 'True' ]
}

isDriverReady(){
  for pod in $deviceDriver;do
    isPodReady "$pod" || return 1
  done
}

waitForDeviceDriver() {
  period=120
  interval=1

  i=0
  while [ "$i" -le "$period" ]; do
    deviceDriver="$(kubectl get pods -l role=openebs-device -o 'jsonpath={.items[*].metadata.name}' -n "$NAMESPACE")"
    if isDriverReady "$deviceDriver"; then
      return 0
    fi

    i=$(( i + interval ))
    echo "Waiting for device-driver to be ready..."
    sleep "$interval"
  done

  echo "Waited for $period seconds, but all pods are not ready yet."
  return 1
}

# wait for device-driver to be up
waitForDeviceDriver

cd $TEST_DIR

kubectl get po -n "$NAMESPACE"

set +e

echo "running ginkgo test case"

ginkgo -v

if [ $? -ne 0 ]; then

lsblk -b
sudo fdisk -l
sudo udevadm info ${disk}p1
sudo parted -l

echo "******************** Device Controller logs***************************** "
dumpControllerLogs 1000

echo "********************* Device Agent logs *********************************"
dumpAgentLogs 1000

echo "get all the pods"
kubectl get pods -owide --all-namespaces

echo "get pvc and pv details"
kubectl get pvc,pv -oyaml --all-namespaces

echo "get sc details"
kubectl get sc --all-namespaces -oyaml

echo "get device volume details"
kubectl get devicevolumes.local.openebs.io -n openebs -oyaml

exit 1
fi

printf "\n\n######### All test cases passed #########\n\n"
