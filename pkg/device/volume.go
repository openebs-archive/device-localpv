// Copyright © 2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package device

import (
	"context"
	"os"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"

	apis "github.com/openebs/device-localpv/pkg/apis/openebs.io/device/v1alpha1"
	"github.com/openebs/device-localpv/pkg/builder/volbuilder"
)

const (
	// DeviceNamespaceKey is the environment variable to get openebs namespace
	//
	// This environment variable is set via kubernetes downward API
	DeviceNamespaceKey string = "DEVICE_DRIVER_NAMESPACE"
	// GoogleAnalyticsKey This environment variable is set via env
	GoogleAnalyticsKey string = "OPENEBS_IO_ENABLE_ANALYTICS"
	// DeviceFinalizer for the DeviceVolume CR
	DeviceFinalizer string = "device.openebs.io/finalizer"
	// DeviceNameKey is key for Device group name
	DeviceNameKey string = "openebs.io/devicename"
	// DeviceNodeKey will be used to insert Label in DeviceVolume CR
	DeviceNodeKey string = "kubernetes.io/nodename"
	// DeviceTopologyKey is supported topology key for the device driver
	DeviceTopologyKey string = "openebs.io/nodename"
	// DeviceStatusPending shows object has not handled yet
	DeviceStatusPending string = "Pending"
	// DeviceStatusFailed shows object operation has failed
	DeviceStatusFailed string = "Failed"
	// DeviceStatusReady shows object has been processed
	DeviceStatusReady string = "Ready"
)

var (
	// DeviceNamespace is openebs system namespace
	DeviceNamespace string

	// NodeID is the NodeID of the node on which the pod is present
	NodeID string

	// GoogleAnalyticsEnabled should send google analytics or not
	GoogleAnalyticsEnabled string
)

func init() {

	DeviceNamespace = os.Getenv(DeviceNamespaceKey)
	if DeviceNamespace == "" && os.Getenv("OPENEBS_NODE_DRIVER") != "" {
		klog.Fatalf("Device_NAMESPACE environment variable not set")
	}
	NodeID = os.Getenv("OPENEBS_NODE_ID")
	if NodeID == "" && os.Getenv("OPENEBS_NODE_DRIVER") != "" {
		klog.Fatalf("NodeID environment variable not set")
	}

	GoogleAnalyticsEnabled = os.Getenv(GoogleAnalyticsKey)
}

// ProvisionVolume creates a DeviceVolume CR,
// watcher for volume is present in CSI agent
func ProvisionVolume(
	vol *apis.DeviceVolume,
) (*apis.DeviceVolume, error) {

	createdVolume, err := volbuilder.NewKubeclient().WithNamespace(DeviceNamespace).Create(vol)
	if err == nil {
		klog.Infof("provisioned volume %s", vol.Name)
	}

	return createdVolume, err
}

// DeleteVolume deletes the corresponding Device Volume CR
func DeleteVolume(volumeID string) (err error) {
	err = volbuilder.NewKubeclient().WithNamespace(DeviceNamespace).Delete(volumeID)
	if err == nil {
		klog.Infof("deprovisioned volume %s", volumeID)
	}

	return
}

// GetDeviceVolume fetches the given DeviceVolume
func GetDeviceVolume(volumeID string) (*apis.DeviceVolume, error) {
	getOptions := metav1.GetOptions{}
	vol, err := volbuilder.NewKubeclient().
		WithNamespace(DeviceNamespace).Get(volumeID, getOptions)
	return vol, err
}

// GetDeviceVolumeState returns DeviceVolume OwnerNode and State for
// the given volume. CreateVolume request may call it again and
// again until volume is "Ready".
func GetDeviceVolumeState(volID string) (string, string, error) {
	getOptions := metav1.GetOptions{}
	vol, err := volbuilder.NewKubeclient().
		WithNamespace(DeviceNamespace).Get(volID, getOptions)

	if err != nil {
		return "", "", err
	}

	return vol.Spec.OwnerNodeID, vol.Status.State, nil
}

// UpdateVolInfo updates DeviceVolume CR with node id and finalizer
func UpdateVolInfo(vol *apis.DeviceVolume) error {
	finalizers := []string{DeviceFinalizer}
	labels := map[string]string{DeviceNodeKey: NodeID}

	if vol.Finalizers != nil {
		return nil
	}

	newVol, err := volbuilder.BuildFrom(vol).
		WithFinalizer(finalizers).
		WithVolumeStatus(DeviceStatusReady).
		WithLabels(labels).Build()

	if err != nil {
		return err
	}

	_, err = volbuilder.NewKubeclient().WithNamespace(DeviceNamespace).Update(newVol)
	return err
}

// RemoveVolFinalizer adds finalizer to DeviceVolume CR
func RemoveVolFinalizer(vol *apis.DeviceVolume) error {
	vol.Finalizers = nil

	_, err := volbuilder.NewKubeclient().WithNamespace(DeviceNamespace).Update(vol)
	return err
}

// WaitForDeviceVolumeProcessed waits till the lvm volume becomes
// ready or failed (i.e reaches to terminal state).
func WaitForDeviceVolumeProcessed(ctx context.Context, volumeID string) (*apis.DeviceVolume, error) {
	timer := time.NewTimer(0)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil, status.FromContextError(ctx.Err())
		case <-timer.C:
		}
		vol, err := GetDeviceVolume(volumeID)
		if err != nil {
			return nil, status.Errorf(codes.Aborted,
				"device: wait failed, not able to get the volume %s %s", volumeID, err.Error())
		}
		if vol.Status.State == DeviceStatusReady ||
			vol.Status.State == DeviceStatusFailed {
			return vol, nil
		}
		timer.Reset(1 * time.Second)
	}
}

// WaitForDeviceVolumeDestroy waits till the lvm volume gets deleted.
func WaitForDeviceVolumeDestroy(ctx context.Context, volumeID string) error {
	timer := time.NewTimer(0)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return status.FromContextError(ctx.Err())
		case <-timer.C:
		}
		_, err := GetDeviceVolume(volumeID)
		if err != nil {
			if k8serror.IsNotFound(err) {
				return nil
			}
			return status.Errorf(codes.Aborted,
				"lvm: destroy wait failed, not able to get the volume %s %s", volumeID, err.Error())
		}
		timer.Reset(1 * time.Second)
	}
}
