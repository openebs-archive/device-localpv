## Parameters

### StorageClass With Custom Node Labels

There can be a use case where we have certain kinds of device present on certain nodes only, and we want a particular type of application to use that device. We can create a storage class with `allowedTopologies` and mention all the nodes there where that device type is present:

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

Here we can have device with meta partition name “test-device” created on the nvme disks and want to use this high performing devices for the applications that need higher IOPS. We can use the above SorageClass to create the PVC and deploy the application using that.

The problem with the above StorageClass is that it works fine if the number of nodes is less, but if the number of nodes is huge, it is cumbersome to list all the nodes like this. In that case, what we can do is, we can label all the similar nodes using the same key value and use that label to create the StorageClass.

``` 
user@k8s-master:~ $ kubectl label node k8s-node-2 openebs.io/devname=nvme
node/k8s-node-2 labeled
user@k8s-master:~ $ kubectl label node k8s-node-1 openebs.io/devname=nvme
node/k8s-node-1 labeled
```

Now, restart the Device-LocalPV Driver (if already deployed, otherwise please ignore) so that it can pick the new node label as the supported topology. Check [faq](./faq.md#1-how-to-add-custom-topology-key) for more details.

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
