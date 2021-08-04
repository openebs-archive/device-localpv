## About this experiment

This experiment upgrades the device-localpv driver components from any previous version to the latest desired stable version or to the master branch ci images. 

## Supported platforms:

K8S : 1.18+

OS : Ubuntu

## Entry-Criteria

- K8s nodes should be ready.
- Do not provision/deprovision any volumes during the upgrade, if we can not control it, then we can scale down the openebs-device-controller stateful set to zero replica which will pause all the provisioning/deprovisioning request. And once upgrade is done, the upgraded Driver will continue the provisioning/deprovisioning process.

## Exit-Criteria

- device-driver should be upgraded to desired version.
- All the components related to device-localpv driver including device-controller and csi node-agents should be running and upraded to desired version as well.
- All the device volumes should be healthy and data prior to the upgrade should not be impacted.
- After upgrade we should be able to provision the volume and other related task with no regressions.

## How to run

- This experiment accepts the parameters in form of kubernetes job environmental variables.
- For running this experiment of upgrading device-localpv driver, clone openens/device-localpv[https://github.com/openebs/device-localpv] repo and then first apply rbac and crds for e2e-framework.
```
kubectl apply -f device-localpv/e2e-tests/hack/rbac.yaml
kubectl apply -f device-localpv/e2e-tests/hack/crds.yaml
```
then update the needed test specific values in run_e2e_test.yml file and create the kubernetes job.
```
kubectl create -f run_e2e_test.yml
```
All the env variables description is provided with the comments in the same file.
After creating kubernetes job, when the job’s pod is instantiated, we can see the logs of that pod which is executing the test-case.

```
kubectl get pods -n e2e
kubectl logs -f <upgrade-device-localpv-xxxxx-xxxxx> -n e2e
```
To get the test-case result, get the corresponding e2e custom-resource `e2eresult` (short name: e2er ) and check its phase (Running or Completed) and result (Pass or Fail).

```
kubectl get e2er
kubectl get e2er upgrade-device-localpv -n e2e --no-headers -o custom-columns=:.spec.testStatus.phase
kubectl get e2er upgrade-device-localpv -n e2e --no-headers -o custom-columns=:.spec.testStatus.result
```