# OpenEBS Local Device CSI Driver
[![Build Status](https://github.com/openebs/device-localpv/actions/workflows/build.yml/badge.svg)](https://github.com/openebs/device-localpv/actions/workflows/build.yml)
[![Slack](https://img.shields.io/badge/chat!!!-slack-ff1493.svg?style=flat-square)](https://kubernetes.slack.com/messages/openebs)
[![Community Meetings](https://img.shields.io/badge/Community-Meetings-blue)](https://hackmd.io/yJb407JWRyiwLU-XDndOLA?view)
[![Go Report](https://goreportcard.com/badge/github.com/openebs/device-localpv)](https://goreportcard.com/report/github.com/openebs/device-localpv)
[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/4871/badge)](https://bestpractices.coreinfrastructure.org/projects/4871)

<img width="300" align="right" alt="OpenEBS Logo" src="https://raw.githubusercontent.com/cncf/artwork/master/projects/openebs/stacked/color/openebs-stacked-color.png" xmlns="http://www.w3.org/1999/html">

CSI Driver for using Local Block Devices 

## Project Status

Currently, the Device-LocalPV CSI Driver is in pre-alpha.

## Usage

### Prerequisites

Before installing the device CSI driver please make sure your Kubernetes Cluster
must meet the following prerequisites:

1. Disks are available on the node with a single 10MB partition having partition name used to 
   identify the disk
2. You have access to install RBAC components into kube-system namespace.
   The OpenEBS Device driver components are installed in kube-system namespace
   to allow them to be flagged as system critical components.

### Supported System

K8S : 1.18+

OS : Ubuntu

### Setup

Find the disk which you want to use for the Device LocalPV, for testing a loopback device can be used

```
truncate -s 1024G /tmp/disk.img
sudo losetup -f /tmp/disk.img --show
```

Create the meta partition on the loop device which will be used for provisioning volumes

```
sudo parted /dev/loop9 mklabel gpt
sudo parted /dev/loop9 mkpart test-device 1MiB 10MiB
```

### Installation

Deploy the Operator yaml

```
kubectl apply -f https://raw.githubusercontent.com/openebs/device-localpv/develop/deploy/device-operator.yaml
```

### Deployment


#### 1. Create a Storage class

```
$ cat sc.yaml

apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: openebs-device-sc
allowVolumeExpansion: true
parameters:
  devname: "test-device"
provisioner: device.csi.openebs.io
volumeBindingMode: WaitForFirstConsumer
```

Check the doc on [storageclasses](docs/storageclasses.md) to know all the supported parameters for Device LocalPV

##### Device Availability

If the device with meta partition is available on certain nodes only, then make use of topology to tell the list of 
nodes where we have the devices available. As shown in the below storage class, we can use allowedTopologies to 
describe device availability on nodes.

```
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: openebs-device-sc
allowVolumeExpansion: true
parameters:
  devname: "test-device"
provisioner: device.csi.openebs.io
allowedTopologies:
- matchLabelExpressions:
  - key: kubernetes.io/hostname
    values:
      - device-node1
      - device-node2
```

The above storage class tells that device with meta partition "test-device" is available on nodes device-node1 and 
device-node2 only. The Device CSI driver will create volumes on those nodes only.

##### Scheduler

The OpenEBS Device driver has its own scheduler which will try to distribute the PV across the nodes so that one node 
should not be loaded with all the volumes. Currently, the driver supports two scheduling algorithms: VolumeWeighted and 
CapacityWeighted, in which it will try to find a device which has lesser number of volumes provisioned in it or 
less capacity of volume provisioned out of a device respectively, from all the nodes where the devices are available. 
To know about how to select scheduler via storage-class See [this](https://github.com/openebs/device-localpv/blob/master/docs/storageclasses.md#storageclass-with-k8s-scheduler).
Once it is able to find the node, it will create a PV for that node and also create a DeviceVolume custom resource for 
the volume with the node information. The watcher for this DeviceVolume CR will get all the information for this object 
and creates a partition with the given size on the mentioned node.

The scheduling algorithm currently only accounts for either the number of volumes or total capacity occupied from a 
device and does not account for other factors like available cpu or memory while making scheduling decisions. So if you 
want to use node selector/affinity rules on the application pod, or have cpu/memory constraints, kubernetes scheduler 
should be used. To make use of kubernetes scheduler, you can set the `volumeBindingMode` as `WaitForFirstConsumer` in 
the storage class. This will cause a delayed binding, i.e kubernetes scheduler will schedule the application pod first, 
and then it will ask the Device driver to create the PV. The driver will then create the PV on the node where the pod 
is scheduled.

```
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: openebs-device-sc
allowVolumeExpansion: true
parameters:
  devname: "test-device"
provisioner: device.csi.openebs.io
volumeBindingMode: WaitForFirstConsumer
```

Please note that once a PV is created for a node, application using that PV will always get scheduled to that particular
node only, as PV will be sticky to that node. The scheduling algorithm by Device driver or kubernetes will come into 
picture only during the deployment time. Once the PV is created, the application can not move anywhere as the data is 
there on the node where the PV is.

#### 2. Create the PVC

```
$ cat pvc.yaml

kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: csi-devicepv
spec:
  storageClassName: openebs-device-sc
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 4Gi
```

Create a PVC using the storage class created for the Device driver.

#### 3. Deploy the application

Create the deployment yaml using the pvc backed by Device driver storage.

```
$ cat fio.yaml

apiVersion: v1
kind: Pod
metadata:
  name: fio
spec:
  restartPolicy: Never
  containers:
  - name: perfrunner
    image: openebs/tests-fio
    command: ["/bin/bash"]
    args: ["-c", "while true ;do sleep 50; done"]
    volumeMounts:
       - mountPath: /datadir
         name: fio-vol
    tty: true
  volumes:
  - name: fio-vol
    persistentVolumeClaim:
      claimName: csi-devicepv
```

After the deployment of the application, we can go to the node and see that the partition is created and is being used as a volume
by the application for reading/writting the data.

#### 4. Deprovisioning

for deprovisioning the volume we can delete the application which is using the volume and then we can go ahead and delete the pv, as part of deletion of pv the partition will be wiped and deleted from the device.

```
$ kubectl delete -f fio.yaml
pod "fio" deleted
$ kubectl delete -f pvc.yaml
persistentvolumeclaim "csi-devicepv" deleted
```

Features
---

- [x] Access Modes
   - [x] ReadWriteOnce
   - ~~ReadOnlyMany~~
   - ~~ReadWriteMany~~
- [x] Volume modes
   - [x] `Filesystem` mode
   - [x] `Block` mode
- [x] Supports fsTypes: `ext4`, `xfs`
- [x] Volume metrics
- [x] Topology
- [ ] Snapshot
- [ ] Clone
- [ ] Volume Resize
- [ ] Thin Provision
- [ ] Backup/Restore
- [ ] Ephemeral inline volume

Project Roadmap
---

The project roadmap is defined and tracked using github projects [here]()
