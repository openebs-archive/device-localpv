/*
 Copyright © 2021 The OpenEBS Authors

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

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=devicenode

// DeviceNode records information about all devices available
// in a node. In general, the openebs node-agent creates the DeviceNode
// object & periodically synchronizing the devices available in the node.
// DeviceNode has an owner reference pointing to the corresponding node object.
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced
type DeviceNode struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Devices []Device `json:"devices"`
}

// Device specifies attributes of a given device that exists on node.
type Device struct {
	// Name of the device(from the meta partition)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// UUID denotes a unique identity of a device.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	UUID string `json:"uuid"`

	// Size specifies the total size of the device.
	// +kubebuilder:validation:Required
	Size resource.Quantity `json:"size"`
	// Free specifies the available capacity of the device.
	// +kubebuilder:validation:Required
	Free resource.Quantity `json:"free"`
}

// DeviceNodeList is a collection of DeviceNode resources
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=devicenodes
type DeviceNodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []DeviceNode `json:"items"`
}
