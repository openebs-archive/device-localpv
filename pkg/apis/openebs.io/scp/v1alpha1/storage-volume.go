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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=storagevolume

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced,shortName=sv
type StorageVolume struct {
	metav1.TypeMeta `json:",inline"`

	// Standard object's metadata.
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines a specification of a storage volume.
	Spec StorageVolumeSpec `json:"spec"`

	// Status represents the current information/status for the storage volume.
	// +optional
	Status StorageVolumeStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=storagevolumes
type StorageVolumeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []StorageVolume `json:"items"`
}

// StorageVolumeSpec defines the desired spec of StorageVolume
// +k8s:openapi-gen=true
type StorageVolumeSpec struct {
	// Capacity of the volume
	Capacity resource.Quantity `json:"capacity"`

	// Capabilities stores the capabilities of a volume that is passed through storage class
	Capabilities *Capabilities `json:"capabilities"`

	// Parameters holds the configuration that is required at the time of volume creation
	// Can be nvme or iscsi specific configurations
	// +optional
	Parameters map[string]string `json:"parameters,omitempty"`

	// Affinity is a group of affinity scheduling rules
	// +optional
	Affinity Affinity `json:"affinity,omitempty"`

	// StorageCohortReference points to the pre-existing StorageCohort resource
	// +optional
	StorageCohortReference *corev1.ObjectReference `json:"storageCohortReference,omitempty"`

	// StoragePoolReference points to the pre-existing StoragePool resource
	// +optional
	StoragePoolReference *corev1.ObjectReference `json:"storagePoolReference,omitempty"`

	// StorageProvisioner specifies the provisioner name that is responsible for volume provisioning tasks.
	// +optional
	StorageProvisioner string `json:"storageProvisioner,omitempty"`
}

// StorageVolumeStatus defines the observed state of StorageVolume
type StorageVolumeStatus struct {
	// Phase defines the state of the volume
	Condition []StorageVolumeCondition `json:"condition,omitempty"`

	// Capacity stores total, used and available size
	// +optional
	Capacity VolumeCapacity `json:"capacity,omitempty"`

	// TargetInfo defines the target information i.e nqn, targetIP and status
	// +optional
	TargetInfo []map[string]string `json:"targetInfo,omitempty"`
}

// StorageVolumeCondition contains condition information for a StorageVolume.
type StorageVolumeCondition struct {
	// Type is the type of the condition.
	Type StorageVolumeConditionType `json:"type"`

	// Condition is the current observed condition of a storage volume
	Condition `json:",inline"`
}

// StorageVolumeConditionType is a valid value for StorageVolumeCondition.Type
type StorageVolumeConditionType string

const (
	// StorageVolumeConditionTypeScheduled represents scheduled StorageVolume object
	StorageVolumeConditionTypeScheduled StorageVolumeConditionType = "Scheduled"

	// StorageVolumeConditionTypeCreationPending represents volume creation pending on the storage nodes
	StorageVolumeConditionTypeCreationPending StorageVolumeConditionType = "CreationPending"

	// StorageVolumeConditionTypeReady represents StorageVolume object is in ready state
	StorageVolumeConditionTypeReady StorageVolumeConditionType = "Ready"
)

// VolumeCapacity defines total, used and available size
type VolumeCapacity struct {
	// Total represents total size of volume
	// +optional
	Total resource.Quantity `json:"total,omitempty"`

	// Used represents used size of the volume
	// +optional
	Used resource.Quantity `json:"used,omitempty"`

	// Available represents available size of the volume
	// +optional
	Available resource.Quantity `json:"available,omitempty"`
}
