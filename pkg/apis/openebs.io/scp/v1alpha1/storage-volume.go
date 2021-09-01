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
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Capacity resource.Quantity `json:"capacity"`

	// ProtocolConfiguration represents protocol specific configurations for NVME
	// +nullable
	ProtocolConfiguration ProtocolConfiguration `json:"protocolConfiguration,omitempty"`

	// VolumeType stores configuration that is passed through storage class
	VolumeType VolumeType `json:"volumeType"`

	// Affinity is a group of affinity scheduling rules
	// +nullable
	Affinity Affinity `json:"affinity,omitempty"`

	// StorageCohortID represents the cohort in which volume has to be provisioned
	StorageCohortName string `json:"storageCohortName"`

	// StoragePoolID represents the pool in which volume has to be created
	StoragePoolName string `json:"storagePoolName"`
}

// ProtocolConfiguration represents protocol specific configurations for NVME
type ProtocolConfiguration struct {
	// AllowedHosts represents the allowed hosts
	AllowedHosts []string `json:"allowedHosts,omitempty"`
}

// VolumeType stores configuration of volumes that is passed through storage class
type VolumeType struct {
	// Different Capabilities for a volume
	Capabilities *Capabilities `json:"capabilities"`
}

// Capabilities of the volume
type Capabilities struct {
	// DataStorage capabilities
	DataStorage *DataStorageCapabilities `json:"dataStorage"`

	// Data security capabilities
	DataSecurity *DataSecurityCapabilities `json:"dataSecurity,omitempty"`

	// IO connectivity capabilities for the volume
	IOConnectivity *IOConnectivityCapabilities `json:"IOConnectivity,omitempty"`

	// IO performance capabilities
	IOPerformance *IOPerformanceCapabilities `json:"IOPerformance,omitempty"`

	MaxIOOperationsPerSecondPerTerabyte string `json:"maxIOOperationsPerSecondPerTerabyte,omitempty"`

	// storage tier to be used for volume creations. eg(Platinum, Gold, Silver)
	StorageTier StorageTier `json:"storageTier"`

	DataProtection *DataProtectionCapabilities `json:"dataProtection,omitempty"`
}

type StorageTier string

const (
	// StorageTierPlatinum indicates that the tier to be used for volume is platinum
	StorageTierPlatinum StorageTier = "P1"

	// StorageTierGold indicates that the tier to be used for volume is gold
	StorageTierGold StorageTier = "P2"

	// StorageTierSilver indicates that the tier to be used for volume is silver
	StorageTierSilver StorageTier = "P3"
)

// DataStorageCapabilities defines data storage capabilities for a volume
type DataStorageCapabilities struct {
	//AccessModes contains the actual access modes for the volume
	AccessModes StorageVolumeAccessMode `json:"accessModes"`

	// ProvisioningPolicy defines provisioning policy type
	ProvisioningPolicy ProvisioningPolicy `json:"provisioningPolicy"`

	// RecoveryTimeObjectives to be supported or not
	RecoveryTimeObjectives RecoveryTimeObjectives `json:"recoveryTimeObjectives"`

	// Compression to be supported or not, if yes the algorithm
	//	+nullable
	Compression *CompressionAlgorithm `json:"compression,omitempty"`

	// Deduplication to be supported or not
	// +nullable
	Deduplication *bool `json:"deduplication,omitempty"`
}

type StorageVolumeAccessMode string

const (
	// can be mounted in read/write mode to exactly 1 host
	ReadWriteOnce StorageVolumeAccessMode = "ReadWriteOnce"
	// can be mounted in read-only mode to many hosts
	ReadOnlyMany StorageVolumeAccessMode = "ReadOnlyMany"
	// can be mounted in read/write mode to many hosts
	ReadWriteMany StorageVolumeAccessMode = "ReadWriteMany"
)


type RecoveryTimeObjectives string

const (
	// OnlineActive indicates that the recovery time objective for a volume is online active
	OnlineActive RecoveryTimeObjectives = "OnlineActive"

	// OnlinePassive indicates that the recovery time objective for a volume is online active
	OnlinePassive RecoveryTimeObjectives = "OnlinePassive"
)

type ProvisioningPolicy string

const (
	// ProvisioningPolicyThick indicates that the provisioning policy is thick
	ProvisioningPolicyThick ProvisioningPolicy = "Thick"

	// ProvisioningPolicyThin indicates that the provisioning policy is thin
	ProvisioningPolicyThin ProvisioningPolicy = "Thin"
)

type CompressionAlgorithm string

const(
	HuffmanCompression CompressionAlgorithm = "huffman"
)

type DataSecurityCapabilities struct {
	MediaEncryption *bool `json:"mediaEncryption,omitempty"`
}

type IOConnectivityCapabilities struct {
	AccessProtocols string `json:"accessProtocols,omitempty"`
}

type IOPerformanceCapabilities struct {
	AverageIOOperationLatencyMicroseconds string `json:"averageIOOperationLatencyMicroseconds,omitempty"`
}

type DataProtectionCapabilities struct {
	// TODO: Decide the Fields
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
	TopologySpreadConstraint []corev1.TopologySpreadConstraint `json:"topologySpreadConstraint,omitempty"`
}

type VolumeAffinityTerm struct {
	TopologyKey string `json:"topologyKey,omitempty"`

	LabelSelector *metav1.LabelSelector `json:"labelSelector,omitempty"`
}

// StorageVolumeStatus defines the observed state of StorageVolume
// +k8s:openapi-gen=true
type StorageVolumeStatus struct {
	// Phase defines the state of the volume
	Phase StorageVolumePhase `json:"phase"`

	// Capacity stores total, used and available size
	// +nullable
	Capacity Capacity `json:"capacity,omitempty"`

	// NVMESpec defines the nvme spec i.e nql, targetIP and status
	// +nullable
	NVMESpec []NVMESpec `json:"NVMESpec,omitempty"`

	// Error denotes the error occurred during provisioning a volume.
	// Error field should only be set when State becomes Failed.
	Error *StorageVolumeError `json:"error,omitempty"`
}

// Capacity defines total, used and available size
type Capacity struct {
	// Total represents total size of volume
	Total resource.Quantity `json:"total,omitempty"`

	// Used represents used size of the volume
	Used resource.Quantity `json:"used,omitempty"`

	// Available represents available size of the volume
	Available resource.Quantity `json:"available,omitempty"`

}

// NVMESpec represents nvme spec for the volume
type NVMESpec struct {

	// Nqn represents target nvme Qualified Name.combination of nodeBase
	Nqn string `json:"nqn,omitempty"`

	// TargetIP IP of the NVME target
	TargetIP string `json:"targetIP,omitempty"`

	// Status of the nvme target
	Status string `json:"status,omitempty"`
}

// StorageVolumePhase represents the current phase of StorageVolume.
type StorageVolumePhase string

const (
	// StorageVolumePhasePending indicates that the StorageVolume is still waiting for
	// the StorageVolume to be created and bound
	StorageVolumePhasePending StorageVolumePhase = "Pending"

	// StorageVolumePhaseFailed indicates that the StorageVolume provisioning
	// has failed
	StorageVolumePhaseFailed StorageVolumePhase = "Failed"

	// StorageVolumePhaseDeleting indicates the the StorageVolume is de-provisioned
	StorageVolumePhaseDeleting StorageVolumePhase = "Deleting"

	// StorageVolumePhaseHealthy indicates the the StorageVolume is healthy
	StorageVolumePhaseHealthy StorageVolumePhase = "Healthy"

	// StorageVolumePhaseUnhealthy indicates the the StorageVolume is unhealthy
	StorageVolumePhaseUnhealthy StorageVolumePhase = "Unhealthy"
)

// StorageVolumeError specifies the error occurred during volume provisioning.
type StorageVolumeError struct {
	Code    StorageVolumeErrorCode `json:"code,omitempty"`
	Message string          `json:"message,omitempty"`
}

// StorageVolumeErrorCode represents the error code to represent
// specific class of errors.
type StorageVolumeErrorCode string

const (
	// Internal represents system internal error.
	Internal StorageVolumeErrorCode = "Internal"
	// InsufficientCapacity represent device doesn't
	// have enough capacity to fit the volume request.
	InsufficientCapacity StorageVolumeErrorCode = "InsufficientCapacity"
)

