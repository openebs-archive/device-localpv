/*
Copyright 2020 The OpenEBS Authors

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

package driver

import (
	"regexp"
	"strconv"

	"k8s.io/klog/v2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openebs/device-localpv/pkg/builder/volbuilder"
	"github.com/openebs/device-localpv/pkg/device"
)

// scheduling algorithm constants
const (
	// pick the node where less volumes are provisioned for the given device name
	VolumeWeighted = "VolumeWeighted"

	// pick the node where total provisioned volumes have occupied less capacity from the given device name
	// this will be the default scheduler when none provided
	CapacityWeighted = "CapacityWeighted"
)

// getVolumeWeightedMap goes through all the devices on all the nodes
// and creates the node mapping of the volume for all the nodes.
// It returns a map which has nodes as key and volumes present
// on the nodes as corresponding value.
func getVolumeWeightedMap(deviceName string) (map[string]int64, error) {
	nmap := map[string]int64{}

	vollist, err := volbuilder.NewKubeclient().
		WithNamespace(device.DeviceNamespace).
		List(metav1.ListOptions{})

	if err != nil {
		return nmap, err
	}

	// create the map of the volume count
	// for the given deviceName
	for _, vol := range vollist.Items {
		devRegex, err := regexp.Compile(vol.Spec.DevName)
		if err != nil {
			klog.Infof("Disk: Regex compile failure %s, %+v", vol.Spec.DevName, err)
			return nil, err
		}
		if devRegex.MatchString(deviceName) {
			nmap[vol.Spec.OwnerNodeID]++
		}
	}

	return nmap, nil
}

// getCapacityWeightedMap goes through all the devices on all the nodes
// and creates the node mapping of the capacity for all the nodes.
// It returns a map which has nodes as key and capacity provisioned
// on the nodes as corresponding value. The scheduler will use this map
// and picks the node which is less weighted.
func getCapacityWeightedMap(deviceName string) (map[string]int64, error) {
	nmap := map[string]int64{}

	volList, err := volbuilder.NewKubeclient().
		WithNamespace(device.DeviceNamespace).
		List(metav1.ListOptions{})

	if err != nil {
		return nmap, err
	}

	// create the map of the volume capacity
	// for the given device name
	for _, vol := range volList.Items {
		devRegex, err := regexp.Compile(vol.Spec.DevName)
		if err != nil {
			klog.Infof("Disk: Regex compile failure %s, %+v", vol.Spec.DevName, err)
			return nil, err
		}
		if devRegex.MatchString(deviceName) {
			volSize, err := strconv.ParseInt(vol.Spec.Capacity, 10, 64)
			if err == nil {
				nmap[vol.Spec.OwnerNodeID] += volSize
			}
		}
	}

	return nmap, nil
}

// getNodeMap returns the node mapping for the given scheduling algorithm
func getNodeMap(schd string, deviceName string) (map[string]int64, error) {
	switch schd {
	case VolumeWeighted:
		return getVolumeWeightedMap(deviceName)
	case CapacityWeighted:
		return getCapacityWeightedMap(deviceName)
	}
	// return CapacityWeighted(default) if not specified
	return getCapacityWeightedMap(deviceName)
}
