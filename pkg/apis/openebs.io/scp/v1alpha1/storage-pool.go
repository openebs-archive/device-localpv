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
// +resource:path=storagepool

// StoragePool describes a Storage Pool resource.
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced,shortName=sp
type StoragePool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Spec is the specifications and configurations to be consumed by the provisioner for a StoragePool resource
	// +required
	Spec StoragePoolSpec `json:"spec"`
	// Status is for handling status of StoragePool resource
	// +optional
	Status StoragePoolStatus `json:"status,omitempty"`
}

// StoragePoolSpec is the spec for a StoragePool resource
type StoragePoolSpec struct {
	// StorageCohortReference points to the StorageCohort resource, the pool is to be created upon
	// +required
	StorageCohortReference corev1.ObjectReference `json:"storageCohortReference"`

	// Provisioner refers to the pool provisioner which will be responsible for creating
	// and managing the pool
	// +required
	Provisioner string `json:"Provisioner"`

	// Configuration points to the configuration, a custom resource, map of parameters or configmap
	// that can be used to specify the pool and its device related configuration.
	// +required
	Configuration interface{} `json:"configuration"`

	// Capabilities to be supported by the StoragePool
	// +optional
	Capabilities Capabilities `json:"capabilities,omitempty"`

	// RequestedCapacity needed for the pool creation
	// +required
	RequestedCapacity resource.Quantity `json:"requestedCapacity"`
}

// StoragePoolStatus is for handling status of StoragePool resource
type StoragePoolStatus struct {
	// ReferenceResource points to the pre-existing pool resource or the created pool resource after creation of pool.
	// +optional
	ReferenceResource corev1.ObjectReference `json:"referenceResource,omitempty"`

	// Capacity of the pool,viz total, used and available capacity
	// +optional
	Capacity StorageCapacity `json:"capacity,omitempty"`

	// IOPs of the pool,viz total, provisioned, used and available IOPs
	// +optional
	IOPs StorageIOPs `json:"IOPs,omitempty"`

	// VolumeSizeMaxLimit for maximum volume size allowed
	// +optional
	VolumeSizeMaxLimit resource.Quantity `json:"volumeSizeMaxLimit,omitempty"`

	// Conditions enlist the conditions of the pool based on various types.
	// +optional
	Conditions []StoragePoolCondition `json:"state,omitempty"`
}

// StorageCapacity for the pool,viz total, used and available capacity
type StorageCapacity struct {
	// Total Capacity of the pool
	// +optional
	Total resource.Quantity `json:"total,omitempty"`

	// Used Capacity of the pool
	// +optional
	Used resource.Quantity `json:"used,omitempty"`

	// Available Capacity of the pool
	// +optional
	Available resource.Quantity `json:"available,omitempty"`

	// Provisioned Capacity of the pool
	// +optional
	Provisioned resource.Quantity `json:"provisioned,omitempty"`
}

// StorageIOPs for the pool,viz total, provisioned, used and available capacity
type StorageIOPs struct {
	// Total IOPs of the pool
	// +optional
	Total uint64 `json:"total,omitempty"`

	// Available IOPs of the pool
	// +optional
	Available uint64 `json:"available,omitempty"`

	// Provisioned IOPs of the pool
	// +optional
	Provisioned uint64 `json:"provisioned,omitempty"`

	// Used IOPs of the pool
	// +optional
	Used uint64 `json:"used,omitempty"`
}

// StoragePoolCondition describes the condition of a StoragePool at a certain point based on a specific type
type StoragePoolCondition struct {
	// Type of StoragePool condition.
	// +required
	Type StoragePoolConditionType `json:"type"`

	// Status of the condition, one of True, False, Unknown.
	// +required
	Status corev1.ConditionStatus `json:"status"`

	// The last time this condition was updated.
	// +optional
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`

	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`

	// The reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty"`

	// A human readable message indicating details about the transition.
	// +optional
	Message string `json:"message,omitempty"`
}

// StoragePoolConditionType specifies the particular state that the conditions are based on.
type StoragePoolConditionType string

const (
	// StoragePoolConditionTypePoolExpansion condition will be available when user triggers
	// pool expansion by adding disks to the pool
	StoragePoolConditionTypePoolExpansion StoragePoolConditionType = "PoolExpansion"

	// StoragePoolConditionTypeDiskReplacement condition will be available when user triggers
	// disk replacement.
	StoragePoolConditionTypeDiskReplacement StoragePoolConditionType = "DiskReplacement"

	// StoragePoolConditionTypeDiskUnavailable condition will be available when one (or) more
	// disks were unavailable
	StoragePoolConditionTypeDiskUnavailable StoragePoolConditionType = "DiskUnavailable"

	// StoragePoolConditionTypelPoolLost condition will be available when the underlying specific pool
	// is not in usable.
	StoragePoolConditionTypelPoolLost StoragePoolConditionType = "PoolOffline"

	// StoragePoolConditionTypePoolHealthy condition will be available when the underlying specific pool
	// is usable.
	StoragePoolConditionTypePoolHealthy StoragePoolConditionType = "PoolHealthy"
)

// StoragePoolList is a list of StoragePool resources
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=storagepools
type StoragePoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []StoragePool `json:"items"`
}
