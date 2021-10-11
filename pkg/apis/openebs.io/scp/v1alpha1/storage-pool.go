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
	StorageCohortReference *corev1.ObjectReference `json:"storageCohortReference"`

	// StorageProvisioner refers to the pool provisioner which will be responsible for creating
	// and managing the pool
	// +optional
	StorageProvisioner string `json:"storageProvisioner,omitempty"`

	// Configuration points to the configuration, a custom resource, map of parameters or configmap
	// that can be used to specify the pool and its device related configuration.
	// +required
	Configuration PoolConfiguration `json:"configuration"`

	// Capabilities to be supported by the StoragePool
	// +optional
	Capabilities Capabilities `json:"capabilities,omitempty"`

	// RequestedCapacity needed for the pool creation
	// +optional
	RequestedCapacity resource.Quantity `json:"requestedCapacity"`
}

// PoolConfiguration is the spec for pool to define all of its configuration
type PoolConfiguration struct {
	// Note: Use of both parameters and reference at the same time should be avoided.
	// But if both are used then its upto the implementation to handle such scenario.

	// Parameters are used to define all the properties or configurations of a pool.
	// Note: This should be empty if the below Reference field is used to define the pool configurations.
	// +optional
	Parameters map[string]string `json:"parameters,omitempty"`

	// Reference can point to either a custom resource or configmap containing all the
	// configuration required for the pool.
	// Note: This should be empty if the above Parameters field is used to define the pool configurations.
	// +optional
	Reference *corev1.ObjectReference `json:"reference,omitempty"`
}

// StoragePoolStatus is for handling status of StoragePool resource
type StoragePoolStatus struct {
	// ReferenceResource points to the pre-existing pool resource or the created pool resource after creation of pool.
	// +optional
	ReferenceResource *corev1.ObjectReference `json:"referenceResource,omitempty"`

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
	// +optional
	Type StoragePoolConditionType `json:"type"`

	// Condition represents the storage pool's current observed condition for the above type
	Condition `json:",inline"`
}

// StoragePoolConditionType specifies the particular state that the conditions are based on.
type StoragePoolConditionType string

const (
	// StoragePoolConditionTypePoolHealthy condition will be available when the underlying specific pool
	// is usable.
	StoragePoolConditionTypePoolHealthy StoragePoolConditionType = "PoolHealthy"

	// StoragePoolConditionTypePoolDegraded condition will be available when the pool is
	// partially unavailable
	StoragePoolConditionTypePoolDegraded StoragePoolConditionType = "PoolDegraded"

	// StoragePoolConditionTypeCreationPending is the transition state of a pool when the pool is
	// initially being provisioned
	StoragePoolConditionTypeCreationPending StoragePoolConditionType = "CreationPending"

	// StoragePoolConditionTypeDeletionPending is the transition state of a pool when the pool is
	// getting deleted
	StoragePoolConditionTypeDeletionPending StoragePoolConditionType = "DeletionPending"

	// StoragePoolConditionTypeUpgradePending is the transition state of a pool when the pool is
	// updated
	StoragePoolConditionTypeUpdatePending StoragePoolConditionType = "UpdatePending"
)

// StoragePoolList is a list of StoragePool resources
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=storagepools
type StoragePoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []StoragePool `json:"items"`
}
