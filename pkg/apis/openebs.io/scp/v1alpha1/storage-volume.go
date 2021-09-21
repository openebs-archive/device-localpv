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
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// StorageVolumeProtectionFinalizer makes sure to clean up external resources before
	// deletion of storage volume
	StorageVolumeProtectionFinalizer = "openebs.io/volume-source-protection"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=storagevolume

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced,shortName=sv
type StorageVolume struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StorageVolumeSpec   `json:"spec"`
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
	Parameters interface{} `json:"parameters,omitempty"`

	// Affinity is a group of affinity scheduling rules
	// +optional
	Affinity Affinity `json:"affinity,omitempty"`

	// StorageCohortReference points to the pre-existing StorageCohort resource
	StorageCohortReference v1.ObjectReference `json:"storageCohortReference"`

	// StoragePoolReference points to the pre-existing StoragePool resource
	StoragePoolReference v1.ObjectReference `json:"storagePoolReference"`
}

// Affinity is a group of affinity scheduling rules
type Affinity struct {

	// Describes volume affinity scheduling rules
	// +optional
	VolumeAffinity []VolumeAffinityTerm `json:"volumeAffinity,omitempty"`

	// Describes volume anti-affinity scheduling rules
	// +optional
	VolumeAntiAffinity []VolumeAffinityTerm `json:"volumeAntiAffinity,omitempty"`

	// Describes cohort affinity scheduling rules
	// +optional
	CohortAffinity []metav1.LabelSelector `json:"cohortAffinity,omitempty"`

	// Describes cohort anti-affinity scheduling rules
	// +optional
	CohortAntiAffinity []metav1.LabelSelector `json:"cohortAntiAffinity,omitempty"`

	// TopologySpreadConstraint specifies how to spread matching volumes among the given topology.
	// +optional
	TopologySpreadConstraint []v1.TopologySpreadConstraint `json:"topologySpreadConstraint,omitempty"`
}

// VolumeAffinityTerm specifies affinity requirements for a StorageVolume
type VolumeAffinityTerm struct {
	// TopologyKey is the key of cohort labels. StorageCohort that have a label with this key
	// and identical values are considered to be in the same topology.
	// +optional
	TopologyKey string `json:"topologyKey,omitempty"`

	// A label query over a set of resources, in this case storage volume.
	// +optional
	LabelSelector *metav1.LabelSelector `json:"labelSelector,omitempty"`
}

// StorageVolumeStatus defines the observed state of StorageVolume
type StorageVolumeStatus struct {
	// Phase defines the state of the volume
	Condition []StorageVolumeCondition `json:"condition,omitempty"`

	// Capacity stores total, used and available size
	// +optional
	Capacity Capacity `json:"capacity,omitempty"`

	// TargetInfo defines the nvme spec i.e nql, targetIP and status
	// +optional
	TargetInfo []TargetInfo `json:"targetInfo,omitempty"`
}

// StorageVolumeCondition contains condition information for a StorageVolume.
type StorageVolumeCondition struct {
	// Type is the type of the condition.
	Type StorageVolumeConditionType `json:"type"`

	// Status is the status of the condition.
	// Can be True, False, Unknown.
	Status v1.ConditionStatus `json:"status"`

	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`

	// Last time we got an update on a given condition.
	// +optional
	LastHeartbeatTime metav1.Time `json:"lastHeartbeatTime,omitempty"`

	// Unique, one-word, CamelCase reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty"`

	// Human-readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty"`
}

// StorageVolumeConditionType is a valid value for StorageVolumeCondition.Type
type StorageVolumeConditionType string

const (
	// StorageVolumeConditionTypePending indicates pending status of the StorageVolume object
	StorageVolumeConditionTypePending StorageVolumeConditionType = "Pending"

	// StorageVolumeConditionTypeScheduled represents scheduled StorageVolume object
	StorageVolumeConditionTypeScheduled StorageVolumeConditionType = "Scheduled"

	// StorageVolumeConditionTypeReady represents StorageVolume object is in ready state
	StorageVolumeConditionTypeReady StorageVolumeConditionType = "Ready"
)

// Capacity defines total, used and available size
type Capacity struct {
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

// TargetInfo represents nvme spec for the volume
type TargetInfo struct {

	// Nqn represents target nvme Qualified Name.combination of nodeBase
	// +optional
	Nqn string `json:"nqn,omitempty"`

	// TargetIP IP of the NVME target
	// +optional
	TargetIP string `json:"targetIP,omitempty"`

	// Status of the nvme target
	// +optional
	Status string `json:"status,omitempty"`
}
