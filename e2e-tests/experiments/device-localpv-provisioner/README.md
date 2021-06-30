## About this experiment

This experiment deploys the device-localpv provisioner in kube-system namespace which includes device-controller and csi-node agent deamonset. Apart from this, meta partition of device creation and generic use-case storage-classes for dynamic provisioning of the volumes based on values provided of env's gets created in this experiment.

## Supported platforms:

K8S : 1.18+

OS : Ubuntu

## Entry-Criteria

- K8s cluster should be in healthy state including all the nodes in ready state.
- If we don't want to use this experiment to deploy device-localpv provisioner, we can directly apply the device-localpv operator file as mentioned below and make sure you have meta partition of devices are created on desired nodes to provision volumes.

```
kubectl apply -f https://raw.githubusercontent.com/openebs/device-localpv/master/deploy/device-operator.yaml
```

## Exit-Criteria

- device-localpv driver components should be deployed successfully and all the pods including device-controller and csi node-agent daemonset are in running state.

## How to run

- This experiment accepts the parameters in form of kubernetes job environmental variables.
- For running this experiment of deploying device-localpv provisioner, clone openens/device-localpv[https://github.com/openebs/device-localpv] repo and then first apply rbac and crds for e2e-framework.
```
kubectl apply -f device-localpv/e2e-tests/hack/rbac.yaml
kubectl apply -f device-localpv/e2e-tests/hack/crds.yaml
```
then update the needed test specific values in run_e2e_test.yml file and create the kubernetes job.
```
kubectl create -f run_e2e_test.yml
```
All the env variables description is provided with the comments in the same file.