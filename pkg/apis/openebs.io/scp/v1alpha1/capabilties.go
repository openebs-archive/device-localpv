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

// Capabilities to be supported by the pool or volume
type Capabilities struct {
	// DataStorage capabilities
	DataStorage *DataStorageCapabilities `json:"dataStorage"`

	// Data security capabilities
	// +nullable
	DataSecurity *DataSecurityCapabilities `json:"dataSecurity,omitempty"`

	// IO connectivity capabilities
	// +nullable
	IOConnectivity *IOConnectivityCapabilities `json:"IOConnectivity,omitempty"`

	// IO performance capabilities
	IOPerformance *IOPerformanceCapabilities `json:"IOPerformance"`

	// DataProtection capabilities
	// +nullable
	DataProtection *DataProtectionCapabilities `json:"dataProtection,omitempty"`
}

// DataStorageCapabilities defines data storage capabilities for a pool or volume
type DataStorageCapabilities struct {
	//AccessModes contains the actual access modes
	AccessModes []AccessMode `json:"accessModes"`

	// ProvisioningPolicy defines provisioning policy type
	ProvisioningPolicy []string `json:"provisioningPolicy"`

	// MultiPathing to be supported or not.
	// +nullable
	MultiPathing []string `json:"multiPathing,omitempty"`

	// Compression to be supported or not, if yes the algorithm
	//	+nullable
	Compression []string `json:"compression,omitempty"`

	// Deduplication to be supported or not.
	// +nullable
	Deduplication []string `json:"deduplication,omitempty"`
}

// AccessMode of the pool or volume
type AccessMode string

const (
	// ReadWriteOnce can be mounted in read/write mode to exactly 1 host
	ReadWriteOnce AccessMode = "ReadWriteOnce"

	// ReadOnlyMany can be mounted in read-only mode to many hosts
	ReadOnlyMany AccessMode = "ReadOnlyMany"

	// ReadWriteMany can be mounted in read/write mode to many hosts
	ReadWriteMany AccessMode = "ReadWriteMany"
)


// DataSecurityCapabilities defines media encryption algorithms and data sanitization policy.
type DataSecurityCapabilities struct {

	// MediaEncryption to be supported or not, if yes the algorithms
	// +nullable
	MediaEncryption []string `json:"mediaEncryption,omitempty"`

	// DataSanitizationPolicy to be supported, viz Clear, CryptographicErase
	// +nullable
	DataSanitizationPolicy []string `json:"dataSanitizationPolicy,omitempty"`
}

// IOConnectivityCapabilities defines access protocols to be supported by the pool or volume
type IOConnectivityCapabilities struct {
	// AccessProtocols to be supported, viz NVMe, NVMeOverFabrics
	// +nullable
	AccessProtocols []string `json:"accessProtocols,omitempty"`
}

// IOPerformanceCapabilities defines IO performance capabilities for the pool or volume.
type IOPerformanceCapabilities struct {
	// AverageIOOperationLatencyMicroseconds to be supported or not.
	// +nullable
	AverageIOOperationLatencyMicroseconds *uint64 `json:"averageIOOperationLatencyMicroseconds,omitempty"`

	// maxIOOperationsPerSecondPerTerabyte to be supported or not.
	// +nullable
	maxIOOperationsPerSecondPerTerabyte *uint64 `json:"maxIOOperationsPerSecondPerTerabyte,omitempty"`

	// storage tier to be used for pool or volume creation. eg(Platinum, Gold, Silver)
	StorageTier StorageTier `json:"storageTier"`
}

type DataProtectionCapabilities struct {
	// TODO: Decide the Fields
}

type StorageTier string

const (
	// StorageTierPlatinum indicates that the tier to be used is platinum
	StorageTierPlatinum StorageTier = "P1"

	// StorageTierGold indicates that the tier to be used is gold
	StorageTierGold StorageTier = "P2"

	// StorageTierSilver indicates that the tier to be used is silver
	StorageTierSilver StorageTier = "P3"
)
