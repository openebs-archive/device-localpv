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
	"k8s.io/apimachinery/pkg/types"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=storagepool

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced
type StoragePool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Spec is the spec for a StoragePool resource
	Spec StoragePoolSpec `json:"spec"`
	// Status is for handling status of StoragePool resource
	Status StoragePoolStatus `json:"status,omitempty"`
}

// StoragePoolSpec is the spec for a StoragePool resource
type StoragePoolSpec struct {
	// StorageCohortName of the cohort the pool is a part of
	StorageCohortName string `json:"storageCohortName,omitempty"`

	// Type of the storage pool, viz lvm, spdk blobstor
	Type StoragePoolType `json:"type,omitempty"`

	// PoolCofiguration based on the above type
	PoolCofiguration StoragePoolCofiguration `json:"typeSpecificConfiguration,omitempty"`

	// DeviceSpec for device related filtering, configuration
	DeviceSpec StoragePoolDeviceSpec `json:"deviceSpec,omitempty"`

	// Capabilities of the Storage pool and the supported configurations
	Capabilities StoragePoolCapabilities `json:"capabilities,omitempty"`

	// RequestedCapacity for the pool creation
	RequestedCapacity resource.Quantity `json:"requestedCapacity,omitempty"`
}

// StoragePoolStatus is for handling status of StoragePool resource
type StoragePoolStatus struct {
	// ReferenceResource
	ReferenceResource StoragePoolReferenceResource `json:"referenceResource,omitempty"`

	// StorageCapacity of the pool,viz total, used and available capacity
	StorageCapacity StoragePoolStorageCapacity `json:"storageCapacity,omitempty"`

	// StorageIOPs of the pool,viz total, provisioned, used and available capacity
	StorageIOPs StoragePoolStorageIOPs `json:"storageIops,omitempty"`

	// DeviceConfiguration for list of Device UIDs and raid configurations currently active
	DeviceConfiguration StoragePoolDeviceConfiguration `json:"deviceConfiguration,omitempty"`

	//VolumeSizeMaxLimit for maximum volume size allowed
	VolumeSizeMaxLimit resource.Quantity `json:"volumeSizeMaxLimit,omitempty"`

	// State of the pool based on various scenarios
	State StoragePoolState `json:"state,omitempty"`
}

// StoragePoolType of the storage pool, viz lvm, spdk blobstor
type StoragePoolType string

const (
	// LVM2StoragePool for lvm pools
	LVM2StoragePool StoragePoolType = "lvm2"

	// SpdkBlobstorStoragePool for spdk-blobstor pools
	SpdkBlobstorStoragePool StoragePoolType = "spdk-blobstor"
)

// StoragePoolCofiguration for the defined type of the pool
type StoragePoolCofiguration struct {
	// LVM pool related configuration
	// +nullable
	LVM *LVMPoolConfiguration

	// SpdkBlobStor pool related configuration
	// +nullable
	SpdkBlobStor *SpdkBlobStorPoolConfiguration
}

// LVMPoolConfiguration for LVM pool related configuration
type LVMPoolConfiguration struct {
	//TODO: Decide the sub fields
}

// SpdkBlobStorPoolConfiguration for SpdkBlobStor pool related configuration
type SpdkBlobStorPoolConfiguration struct {
	//TODO: Decide the sub fields
}

// StoragePoolDeviceSpec for the configuration and filtering related specs for devices
type StoragePoolDeviceSpec struct {
	// MaxDeviceCount for maximum allowed devices
	MaxDeviceCount uint64 `json:"maxDeviceCount,omitempty"`

	// DeviceTypeIdentifier for the device viz.  ssd-4k, ssd-16k, nvme-64k
	DeviceTypeIdentifier StoragePoolDeviceTypeIdentifier `json:"deviceTypeIdentifier,omitempty"`

	// DeviceSelector for labels to match with devices incase they are labeled
	DeviceSelector map[string]string `json:"deviceSelector,omitempty"`

	// DeviceSelector for filtering devices based on FiltersOptions
	DeviceFilters map[string][]FilterOptions `json:"deviceFilters,omitempty"`

	// Devices for the UIDs of the free devices
	Devices []types.UID `json:"devices,omitempty"`
}

type StoragePoolCapabilities struct {
	// DataStorage for the storage pool
	DataStorage StoragePoolDataStorageCapabilities `json:"dataStorage,omitempty"`

	// DataSecurity for the storage pool
	DataSecurity StoragePoolDataSecurityCapabilities `json:"dataSecurity,omitempty"`

	// IOConnectivity for the storage pool
	IOConnectivity StoragePoolIOConnectivityCapabilities `json:"ioConnectivity,omitempty"`

	// IOPerformance for the storage pool
	IOPerformance StoragePoolIOPerformanceCapabilities `json:"ioPerformance,omitempty"`

	// DataProtection for the storage pool
	DataProtection StoragePoolDataProtectionCapabilities `json:"dataProtection,omitempty"`
}

// StoragePoolReferenceResource
type StoragePoolReferenceResource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
}

// StoragePoolStorageCapacity for the pool,viz total, used and available capacity
type StoragePoolStorageCapacity struct {
	// Total Capacity of the pool
	Total resource.Quantity `json:"total,omitempty"`

	// Used Capacity of the pool
	Used resource.Quantity `json:"used,omitempty"`

	// Available Capacity of the pool
	Available resource.Quantity `json:"available,omitempty"`
}

// StoragePoolStorageIOPs for the pool,viz total, provisioned, used and available capacity
type StoragePoolStorageIOPs struct {
	// Total IOPs of the pool
	Total uint64 `json:"total,omitempty"`

	// Available IOPs of the pool
	Available uint64 `json:"available,omitempty"`

	// Provisioned IOPs of the pool
	Provisioned uint64 `json:"provisioned,omitempty"`

	// Used IOPs of the pool
	Used uint64 `json:"used,omitempty"`
}

type StoragePoolDeviceConfiguration struct {
	// Devices for the list of UIDs of the devices
	Devices []types.UID `json:"devices,omitempty"`

	// RaidConfiguration for the raid configurations currently being used
	RaidConfiguration map[RaidType]int `json:"raidConfiguration,omitempty"`
}

// StoragePoolState for the state of the pool for various scenarios
type StoragePoolState string

const (
	// StoragePoolStatusEmpty ensures the create operation is to be done, if import fails.
	StoragePoolStatusEmpty StoragePoolState = ""

	// StoragePoolStatusOnline signifies that the pool is online.
	StoragePoolStatusOnline StoragePoolState = "Online"

	// StoragePoolStatusOffline signifies that the pool is offline.
	StoragePoolStatusOffline StoragePoolState = "Offline"

	// StoragePoolStatusDegraded signifies that the pool is degraded.
	StoragePoolStatusDegraded StoragePoolState = "Degraded"

	// StoragePoolStatusFaulted signifies that the pool is faulted.
	StoragePoolStatusFaulted StoragePoolState = "Faulted"

	// StoragePoolStatusRemoved signifies that the pool is removed.
	StoragePoolStatusRemoved StoragePoolState = "Removed"

	// StoragePoolStatusUnavailable signifies that the pool is not available.
	StoragePoolStatusUnavailable StoragePoolState = "Unavailable"

	// StoragePoolStatusError signifies that the pool status could not be fetched.
	StoragePoolStatusError StoragePoolState = "Error"

	// StoragePoolStatusDeletionFailed ensures the resource deletion has failed.
	StoragePoolStatusDeletionFailed StoragePoolState = "DeletionFailed"

	// StoragePoolStatusInvalid ensures invalid resource.
	StoragePoolStatusInvalid StoragePoolState = "Invalid"

	// StoragePoolStatusErrorDuplicate ensures error due to duplicate resource.
	StoragePoolStatusErrorDuplicate StoragePoolState = "ErrorDuplicate"

	// StoragePoolStatusPending ensures pending task for StoragePool.
	StoragePoolStatusPending StoragePoolState = "Pending"

	// StoragePoolStatusInit is initial state of StoragePool, before pool creation.
	StoragePoolStatusInit StoragePoolState = "Init"

	// StoragePoolStatusCreateFailed is state when pool creation failed
	StoragePoolStatusCreateFailed StoragePoolState = "PoolCreationFailed"
)

// StoragePoolMode for the storage pool
type StoragePoolMode string

const (
	// SharedStoragePoolMode if the pool is to be shared
	SharedStoragePoolMode StoragePoolMode = "Shared"

	// ExclusiveStoragePoolMode if the pool is exclusive
	ExclusiveStoragePoolMode StoragePoolMode = "Exclusive"
)

// RaidType for the pool
type RaidType string

const (
	LinearRaidType RaidType = "linear"
	Raid0RaidType  RaidType = "raid0"
	Raid1RaidType  RaidType = "raid1"
	Raid5RaidType  RaidType = "raid5"
	Raid6RaidType  RaidType = "raid6"
	Raid10RaidType RaidType = "raid10"
)

// StoragePoolDeviceTypeIdentifier for the device type constraint
type StoragePoolDeviceTypeIdentifier struct {
	// Model of the device
	Model string `json:"model,omitempty"`

	// PartNumber of the device
	PartNumber string `json:"partNumber,omitempty"`
}

// FilterOptions for device filters
type FilterOptions struct {
	Key     string `json:"key,omitempty"`
	Name    string `json:"name,omitempty"`
	State   string `json:"state,omitempty"`
	Exclude string `json:"exclude,omitempty"`
	Include string `json:"include,omitempty"`
}

// StoragePoolDataStorageCapabilities for the pool
type StoragePoolDataStorageCapabilities struct {
	// AccessModes to be supported by the pool
	AccessModes []AccessMode `json:"accessModes,omitempty"`

	// ProvisioningPolicies to be supported by the pool
	ProvisioningPolicies []StoragePoolProvisioningPolicy `json:"provisioning_policies,omitempty"`

	// Multipathing scenario to be supported by pool viz, OnlineActive, OnlinePassive
	Multipathing StoragePoolMultipathingOption `json:"multipathing,omitempty"`

	// Compression to be supported or not, if yes the algorithms
	// +nullable
	Compression *[]StoragePoolCompressionAlgorithm `json:"compression,omitempty"`

	// Deduplication to be supported or not
	// +nullable
	Deduplication *bool `json:"deduplication,omitempty"`
}

// StoragePoolDataSecurityCapabilities for the pool
type StoragePoolDataSecurityCapabilities struct {
	// MediaEncryption to be supported or not, if yes the algorithms
	// +nullable
	MediaEncryption *[]StoragePoolMediaEncryptionAlgorithm `json:"mediaEncryption,omitempty"`

	// DataSanitizationPolicy to be supported, viz Clear, CryptographicErase
	DataSanitizationPolicy StoragePoolDataSanitizationPolicy `json:"dataSanitizationPolicy,omitempty"`
}

// StoragePoolIOConnectivityCapabilities for the pool
type StoragePoolIOConnectivityCapabilities struct {
	// AccessProtocols to be supported by the pool, viz NVMe, NVMeOverFabrics
	AccessProtocols []StoragePoolAccessProtocol `json:"accessProtocols,omitempty"`
}

// StoragePoolIOPerformanceCapabilities for the pool
type StoragePoolIOPerformanceCapabilities struct {
	// AverageIOOperationLatencyMicroseconds to be supported by the pool
	AverageIOOperationLatencyMicroseconds uint64 `json:"averageIoOperationLatencyMicroseconds,omitempty"`

	// MaxIOOperationsPerSecondPerTerabyte to be supported by the pool
	MaxIOOperationsPerSecondPerTerabyte uint64 `json:"maxIoOperationsPerSecondPerTerabyte,omitempty"`

	// StorageTier to which the pool belongs to, viz Platinum, Gold, Silver
	StorageTier StoragePoolTier `json:"storageTier,omitempty"`
}

// StoragePoolDataProtectionCapabilities for the pool
type StoragePoolDataProtectionCapabilities struct {
	// Replication to be supported or not
	// +nullable
	Replication *bool `json:"replication,omitempty"`

	// Backup to be supported or not
	// +nullable
	Backup *bool `json:"backup,omitempty"`

	// Snapshots to be supported or not
	// +nullable
	Snapshots *bool `json:"snapshots,omitempty"`
}

// AccessMode of the volumes created on pool
type AccessMode string

const (
	// ReadWriteOnce can be mounted in read/write mode to exactly 1 host
	ReadWriteOnce AccessMode = "ReadWriteOnce"

	// ReadOnlyMany can be mounted in read-only mode to many hosts
	ReadOnlyMany AccessMode = "ReadOnlyMany"

	// ReadWriteMany can be mounted in read/write mode to many hosts
	ReadWriteMany AccessMode = "ReadWriteMany"
)

// StoragePoolProvisioningPolicy for the pool
type StoragePoolProvisioningPolicy string

const (
	// ThickProvisioning for thick provisioning
	ThickProvisioning StoragePoolProvisioningPolicy = "thick"

	// ThinProvisioning for thin provisioning
	ThinProvisioning StoragePoolProvisioningPolicy = "thin"
)

// StoragePoolMultipathingOption supported by the pool
type StoragePoolMultipathingOption string

const (
	OnlineActiveMultipathingOption  StoragePoolMultipathingOption = "OnlineActive"
	OnlinePassiveMultipathingOption StoragePoolMultipathingOption = "OnlinePassive"
	NoneMultipathingOption          StoragePoolMultipathingOption = "None"
)

// StoragePoolCompressionAlgorithm supported by the pool
type StoragePoolCompressionAlgorithm string

// StoragePoolMediaEncryptionAlgorithm supported by the pool
type StoragePoolMediaEncryptionAlgorithm string

// StoragePoolDataSanitizationPolicy supported by the pool
type StoragePoolDataSanitizationPolicy string

const (
	NoneDataSanitizationPolicy               StoragePoolDataSanitizationPolicy = "None"
	ClearDataSanitizationPolicy              StoragePoolDataSanitizationPolicy = "Clear"
	CryptographicEraseDataSanitizationPolicy StoragePoolDataSanitizationPolicy = "CryptographicErase"
)

// StoragePoolAccessProtocol supported by the pool
type StoragePoolAccessProtocol string

const (
	NVMeAccessProtocol           StoragePoolAccessProtocol = "NVMe"
	NVMeOverFabicsAccessProtocol StoragePoolAccessProtocol = "NVMeOverFabics"
)

// StoragePoolTier for the pool
type StoragePoolTier string

const (
	PlatinumTier StoragePoolTier = "P1"
	GoldTier     StoragePoolTier = "P2"
	SilverTier   StoragePoolTier = "P3"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=storagepools
type StoragePoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []StoragePool `json:"items"`
}
