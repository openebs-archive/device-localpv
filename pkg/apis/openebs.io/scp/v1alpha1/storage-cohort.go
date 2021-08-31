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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=storagecohort

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced
type StorageCohort struct {
	metav1.TypeMeta `json:",inline"`

	// Standard object's metadata.
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the behavior of a cohort.
	Spec StorageCohortSpec `json:"spec,omitempty"`

	// Most recently observed status of the cohort.
	// Populated by the cohort operator or cohort manager.
	Status StorageCohortStatus `json:"status,omitempty"`
}

// StorageCohortSpec describes the attributes that a cohort is created with.
type StorageCohortSpec struct {
	// NodeSelector is used to specify the cohort to be considered
	NodeSelector *metav1.LabelSelector `json:"nodeSelector,omitempty"`

	// CohortManager stores the details of the cohort manager responsible for
	// managing the cohort
	CohortManager CohortManagerDetails `json:"cohortManager,omitempty"`

	// StorageProvisioner contains list of all provisioners responsible for
	// the provisioning tasks for different storage solutions in the cohort
	StorageProvisioner []StorageCohortStorageProvisionerDetails `json:"storageProvisioner,omitempty"`
}

// StorageCohortStorageProvisionerDetails stores the different storage provisioners information
// which takes the job pf provisioning pools and volumes
type StorageCohortStorageProvisionerDetails struct {
	// StorageType represents the type of storage solution which the provisioner is responsible for
	// implementing
	// For example: lvm or device-localpv
	StorageType string `json:"storageType,omitempty"`

	// provisioner is the driver expected to handle this cohort.
	// This is an optionally-prefixed name, like a label key.
	// For example: "openebs.io/scp-lvm-provisioner" or "openebs.io/scp-device-provisioner".
	Provisioner string `json:"provisioner,omitempty"`
}

// CohortManagerDetails is information about the cohort manager managing
// the cohort.
type CohortManagerDetails struct {
	// ApiUrl is the cohort manager endpoint used for communicating with the cohort
	ApiUrl string `json:"apiUrl,omitempty"`

	// HealthCheckUrl describes a health check endpoint of cohort manager against which
	// an api call is to be performed to determine whether it is alive or ready to receive traffic.
	HealthCheckUrl string `json:"healthCheckUrl,omitempty"`
	// TODO add authentication & security spec
}

// StorageCohortStatus is information about the current status of a storage cohort.
type StorageCohortStatus struct {
	// StorageCohortPhase is the recently observed lifecycle phase of the cohort.
	StorageCohortPhase StorageCohortPhase `json:"phase,omitempty"`

	// Conditions is an array of current observed cohort component's conditions.
	Conditions []ComponentCondition `json:"conditions,omitempty"`

	// Capabilities represent capabilities that a cohort consists of
	Capabilities StorageCohortCapabilities `json:"capabilities,omitempty"`
}

// StorageCohortCapabilities lists a set of capabilities present within the cohort.
type StorageCohortCapabilities struct {
	// DataStorageCapabilities list storage capabilities supported by the cohort
	DataStorageCapabilities StorageCohortDataStorageCapabilities `json:"dataStorageCapabilities,omitempty"`

	// DataSecurityCapabilities holds data security configuration that will be followed by the cohort.
	DataSecurityCapabilities StorageCohortDataSecurityCapabilities `json:"dataSecurityCapabilities,omitempty"`

	// IOConnectivityCapabilities is the I/O capabilities of the cohort
	IOConnectivityCapabilities StorageCohortIOConnectivityCapabilities `json:"ioConnectivityCapabilities,omitempty"`

	// IOPerformanceCapabilities enlist I/O performance capabilities supported by the cohort
	IOPerformanceCapabilities StorageCohortIOPerformanceCapabilities `json:"ioPerformanceCapabilities,omitempty"`

	// DataProtectionCapabilities list data protection capabilities in case of failures
	DataProtectionCapabilities StorageCohortDataProtectionCapabilities `json:"dataProtectionCapabilities,omitempty"`
}

// StorageCohortDataStorageCapabilities
type StorageCohortDataStorageCapabilities struct {
	// Contains the types of access modes supported by the cohort
	AccessModes []StorageCohortAccessMode `json:"accessModes,omitempty"`

	// ProvisioningPolicies defines what type of pool/volume can be created inside the cohort.
	ProvisioningPolicies []ProvisioningPolicy `json:"provisioning_policies,omitempty"`

	// Multipathing describes the different configurations of accessing
	// storage through different nodes within a cohort
	Multipathing StorageCohortMultipathingOption `json:"multipathing,omitempty"`

	// If specified, the compression algorithm supported by the cohort
	// +nullable
	Compression *[]StorageCohortCompressionAlgorithm `json:"compression,omitempty"`

	// Deduplication, if set to true, ensures that the data is de-duplicated within the cohort
	// +nullable
	Deduplication *bool `json:"deduplication,omitempty"`
}

// StorageCohortDataSecurityCapabilities
type StorageCohortDataSecurityCapabilities struct {
	// MediaEncryption either specifies whether encryption is supported or specifies the encryption algorithm itself
	// +nullable
	MediaEncryption *[]StorageCohortEncryptionAlgorithm `json:"mediaEncryption,omitempty"`

	// DataSanitizationPolicy specifies the supported data sanity checks
	DataSanitizationPolicy StorageCohortDataSanitizationPolicy `json:"dataSanitizationPolicy,omitempty"`
}

// StorageCohortIOConnectivityCapabilities
type StorageCohortIOConnectivityCapabilities struct {
	// Contains the list of access protocols supported by the cohort
	AccessProtocols []StorageCohortAccessProtocol `json:"accessProtocols,omitempty"`
}

// StorageCohortIOPerformanceCapabilities
type StorageCohortIOPerformanceCapabilities struct {
	// Average I/O operation latency(in microseconds) allowed within the cohort
	AverageIOOperationLatencyMicroseconds uint64 `json:"averageIoOperationLatencyMicroseconds,omitempty"`

	// Max I/O operation(per second/TB) allowed within the cohort
	MaxIOOperationsPerSecondPerTerabyte uint64 `json:"maxIoOperationsPerSecondPerTerabyte,omitempty"`

	// List of storage tiers supported by the cohort
	StorageTier []StorageCohortStorageTier `json:"storageTier,omitempty"`
}

// StorageCohortDataProtectionCapabilities
type StorageCohortDataProtectionCapabilities struct {
	// Whether replication is supported by the cohort
	// +optional
	Replication *bool `json:"replication,omitempty"`

	// Backup controls data can be backed up by the cohort
	// +optional
	Backup *bool `json:"backup,omitempty"`

	// whether cohort will support snapshots
	// +optional
	Snapshots *bool `json:"snapshots,omitempty"`
}

// ComponentCondition contains condition information for a storage cohort.
type ComponentCondition struct {
	// Name of the component
	// For example: "cohort-manager"
	Name string `json:"name"`
	// Type of component condition.
	Type ComponentConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status ComponentConditionStatus `json:"status"`
	// Last time we got an update on a given condition.
	LastHeartbeatTime metav1.Time `json:"lastHeartbeatTime,omitempty"`
	// Last time the condition transit from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// (brief) reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`
	// Human readable message indicating details about last transition.
	Message string `json:"message,omitempty"`
}

type StorageCohortEncryptionAlgorithm string

// These are valid encryption algorithms.
const (
	// TODO: verify the algorithms required
	AES StorageCohortEncryptionAlgorithm = "AES"
	DES StorageCohortEncryptionAlgorithm = "DES"
	ECC StorageCohortEncryptionAlgorithm = "ECC"
	RSA StorageCohortEncryptionAlgorithm = "RSA"
)

type StorageCohortDataSanitizationPolicy string

const (
	NoneDataSanitizationPolicy               StorageCohortDataSanitizationPolicy = "None"
	ClearDataSanitizationPolicy              StorageCohortDataSanitizationPolicy = "Clear"
	CryptographicEraseDataSanitizationPolicy StorageCohortDataSanitizationPolicy = "CryptographicErase"
)

type ComponentConditionType string

// These are valid conditions of cohort.
// In the future, we can add more. The current set of conditions are:
// CohortComponentReady, CohortComponentNotReady.
const (
	// CohortComponentReady means cohort component is healthy and ready to perform its task.
	CohortComponentReady ComponentConditionType = "Ready"
	// CohortComponentNotReady means the cohort component is unhealthy and not able to perform its task.
	CohortComponentNotReady ComponentConditionType = "NotReady"
	// TODO add more types if necessary
)

type ComponentConditionStatus string

// These are valid condition statuses. "ConditionTrue" means a component is in the condition.
// "ConditionFalse" means a component is not in the condition. "ConditionUnknown" means
// cohort-operator/cohort-manager can't decide if a cohort component is in the condition or not.
const (
	ConditionTrue    ComponentConditionStatus = "True"
	ConditionFalse   ComponentConditionStatus = "False"
	ConditionUnknown ComponentConditionStatus = "Unknown"
)

// StorageCohortAccessMode defines various access modes for cohort.
type StorageCohortAccessMode string

const (
	// ReadWriteOnce can be mounted in read/write mode to exactly 1 host
	ReadWriteOnce StorageCohortAccessMode = "ReadWriteOnce"
	// ReadOnlyMany can be mounted in read-only mode to many hosts
	ReadOnlyMany StorageCohortAccessMode = "ReadOnlyMany"
	// ReadWriteMany can be mounted in read/write mode to many hosts
	ReadWriteMany StorageCohortAccessMode = "ReadWriteMany"
)

type ProvisioningPolicy string

const (
	ThickProvisioning ProvisioningPolicy = "thick"
	ThinProvisioning  ProvisioningPolicy = "thin"
)

type StorageCohortPhase string

// These are the valid phases of cohort.
const (
	// StorageCohortReady means the cohort has been created/added by the system, and has all its components running.
	StorageCohortReady StorageCohortPhase = "Ready"
	// StorageCohortNotReady means the cohort has been configured but not reachable.
	StorageCohortNotReady StorageCohortPhase = "NotReady"
)

type StorageCohortMultipathingOption string

const (
	// OnlineActiveMultipathingOption represent active/active configuration
	OnlineActiveMultipathingOption StorageCohortMultipathingOption = "OnlineActive"
	// OnlinePassiveMultipathingOption represent active/passive configuration
	OnlinePassiveMultipathingOption StorageCohortMultipathingOption = "OnlinePassive"
	// NoneMultipathingOption represent no configuration
	NoneMultipathingOption StorageCohortMultipathingOption = "None"
)

type StorageCohortCompressionAlgorithm string

// Compression algorithms
const (
	HuffmanCompression StorageCohortCompressionAlgorithm = "huffman"
)

type StorageCohortAccessProtocol string

const (
	NVMeAccessProtocol            StorageCohortAccessProtocol = "NVMe"
	NVMeOverFabricsAccessProtocol StorageCohortAccessProtocol = "NVMeOverFabrics"
)

type StorageCohortStorageTier string

const (
	Platinum StorageCohortStorageTier = "P1"
	Gold     StorageCohortStorageTier = "P2"
	Silver   StorageCohortStorageTier = "P3"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=storagecohorts
type StorageCohortList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []StorageVolume `json:"items"`
}
