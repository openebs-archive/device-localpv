/*
Copyright 2021 The OpenEBS Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tests

import (
	"time"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"

	"github.com/openebs/device-localpv/pkg/builder/volbuilder"
	"github.com/openebs/device-localpv/pkg/device"
	"github.com/openebs/device-localpv/tests/container"
	"github.com/openebs/device-localpv/tests/deploy"
	"github.com/openebs/device-localpv/tests/k8svolume"
	"github.com/openebs/device-localpv/tests/pod"
	"github.com/openebs/device-localpv/tests/pts"
	"github.com/openebs/device-localpv/tests/pvc"
	"github.com/openebs/device-localpv/tests/sc"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IsPVCBoundEventually checks if the pvc is bound or not eventually
func IsPVCBoundEventually(pvcName string) bool {
	return gomega.Eventually(func() bool {
		volume, err := PVCClient.
			Get(pvcName, metav1.GetOptions{})
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
		return pvc.NewForAPIObject(volume).IsBound()
	},
		60, 5).
		Should(gomega.BeTrue())
}

// IsPodRunningEventually return true if the pod comes to running state
func IsPodRunningEventually(namespace, podName string) bool {
	return gomega.Eventually(func() bool {
		p, err := PodClient.
			WithNamespace(namespace).
			Get(podName, metav1.GetOptions{})
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
		return pod.NewForAPIObject(p).
			IsRunning()
	},
		60, 5).
		Should(gomega.BeTrue())
}

// IsPVCDeletedEventually tries to get the deleted pvc
// and returns true if pvc is not found
// else returns false
func IsPVCDeletedEventually(pvcName string) bool {
	return gomega.Eventually(func() bool {
		_, err := PVCClient.
			Get(pvcName, metav1.GetOptions{})
		return k8serrors.IsNotFound(err)
	},
		120, 10).
		Should(gomega.BeTrue())
}

func createFstypeStorageClass(ftype string) {
	var (
		err error
	)

	parameters := map[string]string{
		"devname": DeviceMetaName,
		"fstype":  ftype,
	}

	ginkgo.By("building a " + ftype + " storage class")
	scObj, err = sc.NewBuilder().
		WithGenerateName(scName).
		WithVolumeExpansion(true).
		WithParametersNew(parameters).
		WithProvisioner(LocalProvisioner).Build()
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred(),
		"while building ext4 storageclass obj with prefix {%s}", scName)

	scObj, err = SCClient.Create(scObj)
	gomega.Expect(err).To(gomega.BeNil(), "while creating a ext4 storageclass {%s}", scName)
}

func createStorageClass() {
	var (
		err error
	)

	parameters := map[string]string{
		"devname": DeviceMetaName,
	}

	ginkgo.By("building a default storage class")
	scObj, err = sc.NewBuilder().
		WithGenerateName(scName).
		WithParametersNew(parameters).
		WithProvisioner(LocalProvisioner).Build()
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred(),
		"while building default storageclass obj with prefix {%s}", scName)

	scObj, err = SCClient.Create(scObj)
	gomega.Expect(err).To(gomega.BeNil(), "while creating a default storageclass {%s}", scName)
}

// VerifyDeviceVolume verify the properties of a device-volume
func VerifyDeviceVolume() {
	ginkgo.By("fetching device volume")
	vol, err := DeviceClient.WithNamespace(OpenEBSNamespace).
		Get(pvcObj.Spec.VolumeName, metav1.GetOptions{})
	gomega.Expect(err).To(gomega.BeNil(), "while fetching the device volume {%s}", pvcObj.Spec.VolumeName)

	ginkgo.By("verifying device volume")
	gomega.Expect(vol.Spec.DevName).To(gomega.Equal(scObj.Parameters["devname"]),
		"while checking volume group of device volume", pvcObj.Spec.VolumeName)

	gomega.Expect(vol.Finalizers[0]).To(gomega.Equal(device.DeviceFinalizer), "while checking finializer to be set {%s}", pvcObj.Spec.VolumeName)
}

func deleteStorageClass() {
	err := SCClient.Delete(scObj.Name, &metav1.DeleteOptions{})
	gomega.Expect(err).To(gomega.BeNil(),
		"while deleting device storageclass {%s}", scObj.Name)
}

func createAndVerifyPVC() {
	var (
		err     error
		pvcName = "devicepv-pvc"
	)
	ginkgo.By("building a pvc")
	pvcObj, err = pvc.NewBuilder().
		WithName(pvcName).
		WithNamespace(OpenEBSNamespace).
		WithStorageClass(scObj.Name).
		WithAccessModes(accessModes).
		WithCapacity(capacity).Build()
	gomega.Expect(err).ShouldNot(
		gomega.HaveOccurred(),
		"while building pvc {%s} in namespace {%s}",
		pvcName,
		OpenEBSNamespace,
	)

	ginkgo.By("creating above pvc")
	pvcObj, err = PVCClient.WithNamespace(OpenEBSNamespace).Create(pvcObj)
	gomega.Expect(err).To(
		gomega.BeNil(),
		"while creating pvc {%s} in namespace {%s}",
		pvcName,
		OpenEBSNamespace,
	)

	ginkgo.By("verifying pvc status as bound")

	status := IsPVCBoundEventually(pvcName)
	gomega.Expect(status).To(gomega.Equal(true),
		"while checking status equal to bound")

	pvcObj, err = PVCClient.WithNamespace(OpenEBSNamespace).Get(pvcObj.Name, metav1.GetOptions{})
	gomega.Expect(err).To(
		gomega.BeNil(),
		"while retrieving pvc {%s} in namespace {%s}",
		pvcName,
		OpenEBSNamespace,
	)
}

func createAndVerifyBlockPVC() {
	var (
		err     error
		pvcName = "devicepv-pvc"
	)

	volmode := corev1.PersistentVolumeBlock

	ginkgo.By("building a pvc")
	pvcObj, err = pvc.NewBuilder().
		WithName(pvcName).
		WithNamespace(OpenEBSNamespace).
		WithStorageClass(scObj.Name).
		WithAccessModes(accessModes).
		WithVolumeMode(&volmode).
		WithCapacity(capacity).Build()
	gomega.Expect(err).ShouldNot(
		gomega.HaveOccurred(),
		"while building pvc {%s} in namespace {%s}",
		pvcName,
		OpenEBSNamespace,
	)

	ginkgo.By("creating above pvc")
	pvcObj, err = PVCClient.WithNamespace(OpenEBSNamespace).Create(pvcObj)
	gomega.Expect(err).To(
		gomega.BeNil(),
		"while creating pvc {%s} in namespace {%s}",
		pvcName,
		OpenEBSNamespace,
	)

	ginkgo.By("verifying pvc status as bound")

	status := IsPVCBoundEventually(pvcName)
	gomega.Expect(status).To(gomega.Equal(true),
		"while checking status equal to bound")

	pvcObj, err = PVCClient.WithNamespace(OpenEBSNamespace).Get(pvcObj.Name, metav1.GetOptions{})
	gomega.Expect(err).To(
		gomega.BeNil(),
		"while retrieving pvc {%s} in namespace {%s}",
		pvcName,
		OpenEBSNamespace,
	)
}

func createDeployVerifyApp() {
	ginkgo.By("creating and deploying app pod")
	createAndDeployAppPod(appName)
	time.Sleep(30 * time.Second)
	ginkgo.By("verifying app pod is running", verifyAppPodRunning)
}

func createAndDeployAppPod(appname string) {
	var err error
	ginkgo.By("building a busybox app pod deployment using above device volume")
	deployObj, err = deploy.NewBuilder().
		WithName(appname).
		WithNamespace(OpenEBSNamespace).
		WithLabelsNew(
			map[string]string{
				"app": "busybox",
			},
		).
		WithSelectorMatchLabelsNew(
			map[string]string{
				"app": "busybox",
			},
		).
		WithPodTemplateSpecBuilder(
			pts.NewBuilder().
				WithLabelsNew(
					map[string]string{
						"app": "busybox",
					},
				).
				WithContainerBuilders(
					container.NewBuilder().
						WithImage("busybox").
						WithName("busybox").
						WithImagePullPolicy(corev1.PullIfNotPresent).
						WithCommandNew(
							[]string{
								"sh",
								"-c",
								"date > /mnt/datadir/date.txt; sync; sleep 5; sync; tail -f /dev/null;",
							},
						).
						WithVolumeMountsNew(
							[]corev1.VolumeMount{
								corev1.VolumeMount{
									Name:      "datavol1",
									MountPath: "/mnt/datadir",
								},
							},
						),
				).
				WithVolumeBuilders(
					k8svolume.NewBuilder().
						WithName("datavol1").
						WithPVCSource(pvcObj.Name),
				),
		).
		Build()

	gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "while building app deployement {%s}", appName)

	deployObj, err = DeployClient.WithNamespace(OpenEBSNamespace).Create(deployObj)
	gomega.Expect(err).ShouldNot(
		gomega.HaveOccurred(),
		"while creating pod {%s} in namespace {%s}",
		appName,
		OpenEBSNamespace,
	)
}

func createAndDeployBlockAppPod() {
	var err error
	ginkgo.By("building a busybox app pod deployment using above device volume")
	deployObj, err = deploy.NewBuilder().
		WithName(appName).
		WithNamespace(OpenEBSNamespace).
		WithLabelsNew(
			map[string]string{
				"app": "busybox",
			},
		).
		WithSelectorMatchLabelsNew(
			map[string]string{
				"app": "busybox",
			},
		).
		WithPodTemplateSpecBuilder(
			pts.NewBuilder().
				WithLabelsNew(
					map[string]string{
						"app": "busybox",
					},
				).
				WithContainerBuilders(
					container.NewBuilder().
						WithImage("busybox").
						WithName("busybox").
						WithImagePullPolicy(corev1.PullIfNotPresent).
						WithCommandNew(
							[]string{
								"sh",
								"-c",
								"date > /mnt/datadir/date.txt; sync; sleep 5; sync; tail -f /dev/null;",
							},
						).
						WithVolumeDevicesNew(
							[]corev1.VolumeDevice{
								corev1.VolumeDevice{
									Name:       "datavol1",
									DevicePath: "/dev/xvda",
								},
							},
						),
				).
				WithVolumeBuilders(
					k8svolume.NewBuilder().
						WithName("datavol1").
						WithPVCSource(pvcObj.Name),
				),
		).
		Build()

	gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "while building app deployement {%s}", appName)

	deployObj, err = DeployClient.WithNamespace(OpenEBSNamespace).Create(deployObj)
	gomega.Expect(err).ShouldNot(
		gomega.HaveOccurred(),
		"while creating pod {%s} in namespace {%s}",
		appName,
		OpenEBSNamespace,
	)
}

func createDeployVerifyBlockApp() {
	ginkgo.By("creating and deploying app pod", createAndDeployBlockAppPod)
	time.Sleep(30 * time.Second)
	ginkgo.By("verifying app pod is running", verifyAppPodRunning)
}

func verifyAppPodRunning() {
	var err error
	appPod, err = PodClient.WithNamespace(OpenEBSNamespace).
		List(metav1.ListOptions{
			LabelSelector: "app=busybox",
		},
		)
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "while verifying application pod")

	status := IsPodRunningEventually(OpenEBSNamespace, appPod.Items[0].Name)
	gomega.Expect(status).To(gomega.Equal(true), "while checking status of pod {%s}", appPod.Items[0].Name)
}

func deleteAppDeployment(appname string) {
	err := DeployClient.WithNamespace(OpenEBSNamespace).
		Delete(appname, &metav1.DeleteOptions{})
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "while deleting application pod")
}

func deletePVC(pvcname string) {
	err := PVCClient.WithNamespace(OpenEBSNamespace).Delete(pvcname, &metav1.DeleteOptions{})
	gomega.Expect(err).To(
		gomega.BeNil(),
		"while deleting pvc {%s} in namespace {%s}",
		pvcname,
		OpenEBSNamespace,
	)
	ginkgo.By("verifying deleted pvc")
	status := IsPVCDeletedEventually(pvcname)
	gomega.Expect(status).To(gomega.Equal(true), "while trying to get deleted pvc")
}

// IsDeviceVolDeletedEventually : After deleted check if resource still exists
func IsDeviceVolDeletedEventually() bool {
	ginkgo.By("Fetching device volume to check if its deleted")
	return gomega.Eventually(func() bool {
		_, err := DeviceClient.WithNamespace(OpenEBSNamespace).Get(DeviceVolName, metav1.GetOptions{})
		return k8serrors.IsNotFound(err)
	},
		120, 5).
		Should(gomega.BeTrue())

}

// CreateDeviceVolume : Create new DeviceVol resource
func CreateDeviceVolume() {

	nodeList, err := NodeClient.List(metav1.ListOptions{})
	gomega.Expect(err).To(gomega.BeNil(), "while retrieving DeviceNode")
	gomega.Expect(len(nodeList.Items) > 0).To(gomega.BeTrue(), "while retrieving DeviceNode")

	ginkgo.By("Building DeviceVol")
	capacity := "109951162777600" // 100Ti
	volObj, err := volbuilder.NewBuilder().
		WithName(DeviceVolName).
		WithCapacity(capacity).
		WithDeviceName(DeviceMetaName).
		WithOwnerNode(nodeList.Items[0].ObjectMeta.Name).
		WithVolumeStatus(device.DeviceStatusPending).Build()

	ginkgo.By("Creating the above DeviceVol")
	volObj, err = DeviceClient.WithNamespace(OpenEBSNamespace).Create(volObj)
	gomega.Expect(err).To(
		gomega.BeNil(),
		"while creating DeviceVol {%s} in namespace {%s}",
		DeviceVolName,
		OpenEBSNamespace,
	)

	_, err = DeviceClient.WithNamespace(OpenEBSNamespace).Get(DeviceVolName, metav1.GetOptions{})
	gomega.Expect(err).To(
		gomega.BeNil(),
		"while retrieving DeviceVol {%s} in namespace {%s}",
		DeviceVolName,
		OpenEBSNamespace,
	)
}

// IsDeviceVolStatusFailedEventually : Checks for the DeviceVol status and waits till it becomes failed
func IsDeviceVolStatusFailedEventually() {
	status := gomega.Eventually(func() bool {
		deviceVol, _ := DeviceClient.WithNamespace(OpenEBSNamespace).Get(DeviceVolName, metav1.GetOptions{})
		return deviceVol.Status.State == device.DeviceStatusFailed
	},
		120, 5).
		Should(gomega.BeTrue())

	gomega.Expect(status).To(gomega.Equal(true), "while checking the status of DeviceVol")
	ginkgo.By("DeviceVol is in failed state")
}

func deleteDeviceVol() {
	err := DeviceClient.WithNamespace(OpenEBSNamespace).Delete(DeviceVolName)
	gomega.Expect(err).To(
		gomega.BeNil(),
		"while deleting deviceVol {%s} in namespace {%s}",
		DeviceVolName,
		OpenEBSNamespace,
	)
	status := IsDeviceVolDeletedEventually()
	gomega.Expect(status).To(gomega.Equal(true), "while trying to get deleted DeviceVol")
}
