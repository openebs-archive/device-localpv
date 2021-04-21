# OpenEBS Local Device CSI Driver

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
kubectl apply -f https://raw.githubusercontent.com/openebs/device-localpv/master/deploy/device-operator.yaml
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

If the device with meta partition is available on certain nodes only, then make use of topology to tell the list of nodes where we have the devices available.
As shown in the below storage class, we can use allowedTopologies to describe device availability on nodes.

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

The above storage class tells that device with meta partition "test-device" is available on nodes device-node1 and device-node2 only. The Device CSI driver will create volumes on those nodes only.


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

