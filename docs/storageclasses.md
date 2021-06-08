## Parameters

### devname (*must* parameter)

devname specifies the name of the device where the volume has been created. The *devname* is the must argument. It is 
the name of the meta partition that is created on the disk.

```console
user@host:~$ parted /dev/sdd print
Model: VMware Virtual disk (scsi)
Disk /dev/sdd: 17.2GB
Sector size (logical/physical): 512B/512B
Partition Table: gpt
Disk Flags: 

Number  Start   End     Size    File system  Name         Flags
 1      1049kB  10.5MB  9437kB               test-device
```

The name of first partition is the devname.

```
devname: "test-device"
```



### StorageClass With k8s Scheduler

The Device-LocalPV Driver has two types of its own scheduling logic, VolumeWeighted and CapacityWeighted. To choose any 
one of the scheduler add scheduler parameter in storage class and give its value accordingly.
```
parameters:
 scheduler: "VolumeWeighted"
 devname: "test-device"
```
CapacityWeighted is the default scheduler in device-localpv driver, so even if we don't use scheduler parameter in 
storage-class, driver will pick the node where total provisioned volumes have occupied less capacity on the given 
device. On the other hand for using VolumeWeighted scheduler, we have to specify it under scheduler parameter in 
storage-class. Then driver will pick the node to create volume where device is less loaded with the volumes/partitions. 
Here, it just checks the volume count and creates the volume where less volume is configured in a given device. It does 
not account for other factors like available CPU or memory while making scheduling decisions.

In case where you want to use node selector/affinity rules on the application pod or have CPU/Memory constraints, 
the Kubernetes scheduler should be used. To make use of Kubernetes scheduler, we can set the volumeBindingMode as 
WaitForFirstConsumer in the storage class:

```yaml
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

Here, in this case, the Kubernetes scheduler will select a node for the application pod and then ask the Device-LocalPV 
driver to create the volume on the selected node. The driver will create the volume where the pod has been scheduled.


### StorageClass With Custom Node Labels

There can be a use case where we have certain kinds of device present on certain nodes only, and we want a particular 
type of application to use that device. We can create a storage class with `allowedTopologies` and mention all the nodes
there where that device type is present:

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
 name: device-sc
allowVolumeExpansion: true
parameters:
  devname: "test-device"
provisioner: device.csi.openebs.io
allowedTopologies:
- matchLabelExpressions:
  - key: openebs.io/nodename
    values:
    - node-1
    - node-2
```

Here we can have device with meta partition name “test-device” created on the nvme disks and want to use this high 
performing devices for the applications that need higher IOPS. We can use the above SorageClass to create the PVC and 
deploy the application using that.

The problem with the above StorageClass is that it works fine if the number of nodes is less, but if the number of nodes
is huge, it is cumbersome to list all the nodes like this. In that case, what we can do is, we can label all the similar
nodes using the same key value and use that label to create the StorageClass.

``` 
user@k8s-master:~ $ kubectl label node k8s-node-2 openebs.io/devname=nvme
node/k8s-node-2 labeled
user@k8s-master:~ $ kubectl label node k8s-node-1 openebs.io/devname=nvme
node/k8s-node-1 labeled
```

Now, restart the Device-LocalPV Driver (if already deployed, otherwise please ignore) so that it can pick the new node 
label as the supported topology. Check [faq](./faq.md#1-how-to-add-custom-topology-key) for more details.

```
$ kubectl delete po -n kube-system -l role=openebs-device
```

Now, we can create the StorageClass like this:

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
 name: nvme-device-sc
allowVolumeExpansion: true
parameters:
 devname: "test-device"
provisioner: device.csi.openebs.io
allowedTopologies:
- matchLabelExpressions:
  - key: openebs.io/devname
    values:
    - nvme
```

Here, the volumes will be provisioned on the nodes which has label “openebs.io/devname” set as “nvme”.
