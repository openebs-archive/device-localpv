/*
Copyright 2017 The Kubernetes Authors.

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

package device

import (
	apis "github.com/openebs/device-localpv/pkg/apis/openebs.io/device/v1alpha1"
)

// zfs related constants
const (
	DevPath = "/dev/"
)

// TODO @praveengt
func CreateVolume(vol *apis.DeviceVolume) error {
	return nil
}

// TODO @praveengt
func DestroyVolume(vol *apis.DeviceVolume) error {

}

// TODO @praveengt
func CheckVolumeExists(vol *apis.DeviceVolume) (bool, error) {

}

// GetVolumeDevPath return the dev path for volume
func GetVolumeDevPath(vol *apis.DeviceVolume) (string, error) {
	devicePath, err := getDevPathFromMetaPartition(vol)

	return devicePath, err
}

// TODO @praveengt
// reads the information from partition and get the device path
func getDevPathFromMetaPartition(vol *apis.DeviceVolume) (string, error) {
	return "", nil
}
