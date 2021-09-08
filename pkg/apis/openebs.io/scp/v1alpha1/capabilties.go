/*
Copyright © 2019 The OpenEBS Authors

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

// Capabilities defines a set of attributes, which describe properties that the
// storage pool or volume may support.
type Capabilities struct {
	// DataStorage specifies accessModes, provisioning, multiPathing, compression
	// and deduplication characteristics supported by a pool or a volume.
	DataStorage *DataStorageCapabilities `json:"dataStorage"`

	// DataSecurity specifies security standards that is supported.
	// +optional
	DataSecurity *DataSecurityCapabilities `json:"dataSecurity,omitempty"`

	// IOConnectivity describes capabilities to support various IO Connectivity options i.e accessProtocols
	// +optional
	IOConnectivity *IOConnectivityCapabilities `json:"IOConnectivity,omitempty"`

	// IOPerformance describe the capabilities to support various IO performance options.
	IOPerformance *IOPerformanceCapabilities `json:"IOPerformance"`

	// DataProtection describes data protection capabilities.
	// +optional
	DataProtection *DataProtectionCapabilities `json:"dataProtection,omitempty"`
}

// DataStorageCapabilities defines data storage capabilities for a pool or volume
type DataStorageCapabilities struct {
	//AccessModes contains the actual access modes, viz ReadWriteOnce, ReadOnlyMany, ReadWriteMany
	AccessModes []AccessMode `json:"accessModes"`

	// ProvisioningPolicy defines provisioning policy type, viz thick, thin
	ProvisioningPolicy []ProvisioningPolicy `json:"provisioningPolicy"`

	// MultiPathing to be supported or not.
	// +optional
	MultiPathing []string `json:"multiPathing,omitempty"`

	// Compression to be supported or not, if yes the algorithm
	// +optional
	Compression []string `json:"compression,omitempty"`

	// Deduplication to be supported or not.
	// +optional
	Deduplication []string `json:"deduplication,omitempty"`
}

// AccessMode of the pool or volume, viz ReadWriteOnce, ReadOnlyMany and ReadWriteMany
type AccessMode string

const (
	// AccessModeReadWriteOnce can be mounted in read/write mode to exactly 1 host
	AccessModeReadWriteOnce AccessMode = "ReadWriteOnce"

	// AccessModeReadOnlyMany can be mounted in read-only mode to many hosts
	AccessModeReadOnlyMany AccessMode = "ReadOnlyMany"

	// AccessModeReadWriteMany can be mounted in read/write mode to many hosts
	AccessModeReadWriteMany AccessMode = "ReadWriteMany"
)

// ProvisioningPolicy specifies provisioning type, viz thick, thin
type ProvisioningPolicy string

const (
	// ProvisioningPolicyThin specifies thin provisioning
	ProvisioningPolicyThin ProvisioningPolicy = "thin"

	// ProvisioningPolicyThick specifies thick provisioning
	ProvisioningPolicyThick ProvisioningPolicy = "thick"
)

// DataSecurityCapabilities defines media encryption algorithms and data sanitization policy.
type DataSecurityCapabilities struct {

	// MediaEncryption to be supported or not, if yes the algorithms
	// +optional
	MediaEncryption []string `json:"mediaEncryption,omitempty"`

	// DataSanitizationPolicy to be supported, viz Clear, CryptographicErase
	// +optional
	DataSanitizationPolicy []DataSanitizationPolicy `json:"dataSanitizationPolicy,omitempty"`
}

// DataSanitizationPolicy specify the data sanitization policy, viz None, Clear, CryptographicErase
type DataSanitizationPolicy string

const (
	// DataSanitizationPolicyNone specifies no sanitization policy
	DataSanitizationPolicyNone DataSanitizationPolicy = "None"

	// DataSanitizationPolicyClear sanitize data in all user-addressable storage
	// locations for protection against simple non-invasive data recovery techniques.
	DataSanitizationPolicyClear DataSanitizationPolicy = "Clear"

	// DataSanitizationPolicyCryptographicErase specifies to leverages the encryption of target data by
	// enabling sanitization of the target data’s encryption key.
	DataSanitizationPolicyCryptographicErase DataSanitizationPolicy = "CryptographicErase"
)

// IOConnectivityCapabilities defines access protocols to be supported by the pool or volume
type IOConnectivityCapabilities struct {
	// AccessProtocols to be supported, viz NVMe, NVMeOverFabrics, iSCSI
	// +optional
	AccessProtocols []AccessProtocols `json:"accessProtocols,omitempty"`
}

// AccessProtocols supported, viz NVMe, NVMeOverFabrics, iSCSI
type AccessProtocols string

const (
	// AccessProtocolNVMe specifies NVMe protocol
	AccessProtocolNVMe AccessProtocols = "NVMe"

	// AccessProtocolNVMeOverFabrics specifies NVMeOverFabrics protocol
	AccessProtocolNVMeOverFabrics AccessProtocols = "NVMeOverFabrics"

	// AccessProtocolISCSI specifies NVMeOverFabrics protocol
	AccessProtocolISCSI AccessProtocols = "iSCSI"
)

// IOPerformanceCapabilities defines IO performance capabilities for the pool or volume.
type IOPerformanceCapabilities struct {
	// AverageIOOperationLatencyMicroseconds to be supported or not.
	// +optional
	AverageIOOperationLatencyMicroseconds *int64 `json:"averageIOOperationLatencyMicroseconds,omitempty"`

	// maxIOOperationsPerSecondPerTerabyte to be supported or not.
	// +optional
	MaxIOPSPerTB *int64 `json:"maxIOPSPerTB,omitempty"`

	// StorageTier is a classification of the service based on several factors
	// like performance, redundancy, availability, etc. that can be used for pool or volume creation, viz Platinum, Gold, Silver)
	StorageTier string `json:"storageTier"`
}

// DataProtectionCapabilities defines data protection capabilties.
type DataProtectionCapabilities struct {
	// TODO: Decide the Fields
}
