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
	Spec              StoragePoolSpec   `json:"spec"`
	Status            StoragePoolStatus `json:"status,omitempty"`
}

// StoragePoolSpec is the spec for a StoragePool resource
type StoragePoolSpec struct {
	StorageCohortID           string                               `json:"storageCohortID,omitempty"`
	Type                      StoragePoolType                      `json:"type,omitempty"`
	TypeSpecificConfiguration StoragePoolTypeSpecificConfiguration `json:"typeSpecificConfiguration,omitempty"`
	DeviceSpec                StoragePoolDeviceSpec                `json:"deviceSpec,omitempty"`
	Capabilities              StoragePoolCapabilities              `json:"capabilities,omitempty"`
	RequestedCapacity         resource.Quantity                    `json:"requestedCapacity,omitempty"`
}

// StoragePoolStatus is the handing status of StoragePool resource
type StoragePoolStatus struct {
	ReferenceResource   StoragePoolReferenceResource   `json:"referenceResource,omitempty"`
	StorageCapacity     StoragePoolStorageCapacity     `json:"storageCapacity,omitempty"`
	StorageIOPs         StoragePoolStorageIOPs         `json:"storageIops,omitempty"`
	DeviceConfiguration StoragePoolDeviceConfiguration `json:"deviceConfiguration,omitempty"`
	VolumeSizeMaxLimit  resource.Quantity              `json:"volumeSizeMaxLimit,omitempty"`
	State               StoragePoolState               `json:"state,omitempty"`
}

type StoragePoolType string

const (
	// TODO: ADD FOR CURRENT USE CASES
	LVM2StoragePoolType         StoragePoolType = "lvm2"
	ZFSStoragePoolType          StoragePoolType = "zfs"
	MayaStorStoragePoolType     StoragePoolType = "mayastor"
	SPDKBlobstorStoragePoolType StoragePoolType = "spdk blobstor"
)

type StoragePoolTypeSpecificConfiguration struct {
	PoolType            StoragePoolType            `json:"poolType,omitempty"`
	Mode                StoragePoolMode            `json:"mode,omitempty"`
	SupportedRaidTypes  map[RaidType]int           `json:"supportedRaidTypes,omitempty"`
	PoolSpecificOptions StoragePoolSpecificOptions `json:"poolSpecificOptions,omitempty"`
}

type StoragePoolDeviceSpec struct {
	MaxDeviceCount       uint64                          `json:"maxDeviceCount,omitempty"`
	DeviceTypeIdentifier StoragePoolDeviceTypeIdentifier `json:"deviceTypeIdentifier,omitempty"`
	DeviceSelector       map[string]string               `json:"deviceSelector,omitempty"`
	DeviceFilters        map[string][]FilterOptions      `json:"deviceFilters,omitempty"`
	Devices              []types.UID                     `json:"devices,omitempty"`
}

type StoragePoolCapabilities struct {
	DataStorageCapabilities    StoragePoolDataStorageCapabilities    `json:"dataStorageCapabilities,omitempty"`
	DataSecurityCapabilities   StoragePoolDataSecurityCapabilities   `json:"dataSecurityCapabilities,omitempty"`
	IOConnectivityCapabilities StoragePoolIOConnectivityCapabilities `json:"ioConnectivityCapabilities,omitempty"`
	IOPerformanceCapabilities  StoragePoolIOPerformanceCapabilities  `json:"ioPerformanceCapabilities,omitempty"`
	DataProtectionCapabilities StoragePoolDataProtectionCapabilities `json:"dataProtectionCapabilities,omitempty"`
}

type StoragePoolReferenceResource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
}

type StoragePoolStorageCapacity struct {
	TotalCapacity     resource.Quantity `json:"totalCapacity,omitempty"`
	UsedCapacity      resource.Quantity `json:"usedCapacity,omitempty"`
	AvailableCapacity resource.Quantity `json:"availableCapacity,omitempty"`
}

type StoragePoolStorageIOPs struct {
	TotalIOPs       uint64 `json:"totalIops,omitempty"`
	AvailableIOPs   uint64 `json:"availableIops,omitempty"`
	ProvisionedIOPs uint64 `json:"provisionedIops,omitempty"`
	UsedIOPs        uint64 `json:"usedIops,omitempty"`
}

type StoragePoolDeviceConfiguration struct {
	Devices           []types.UID      `json:"devices,omitempty"`
	RaidConfiguration map[RaidType]int `json:"raidConfiguration,omitempty"`
}

type StoragePoolState string

const (
	// StoragePoolStatusEmpty ensures the create operation is to be done, if import fails.
	StoragePoolStatusEmpty StoragePoolState = ""
	// StoragePoolStatusOnline signifies that the pool is online.
	StoragePoolStatusOnline StoragePoolState = "ONLINE"
	// StoragePoolStatusOffline signifies that the pool is offline.
	StoragePoolStatusOffline StoragePoolState = "OFFLINE"
	// StoragePoolStatusDegraded signifies that the pool is degraded.
	StoragePoolStatusDegraded StoragePoolState = "DEGRADED"
	// StoragePoolStatusFaulted signifies that the pool is faulted.
	StoragePoolStatusFaulted StoragePoolState = "FAULTED"
	// StoragePoolStatusRemoved signifies that the pool is removed.
	StoragePoolStatusRemoved StoragePoolState = "REMOVED"
	// StoragePoolStatusUnavail signifies that the pool is not available.
	StoragePoolStatusUnavail StoragePoolState = "UNAVAIL"
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
	// StoragePoolStatusInit is initial state of CSP, before pool creation.
	StoragePoolStatusInit StoragePoolState = "Init"
	// StoragePoolStatusCreateFailed is state when pool creation failed
	StoragePoolStatusCreateFailed StoragePoolState = "PoolCreationFailed"
)

type StoragePoolMode string

const (
	// TODO: ADD FOR CURRENT USE CASES
	SharedStoragePoolMode    StoragePoolMode = "Shared"
	ExclusiveStoragePoolMode StoragePoolMode = "Exclusive"
)

type RaidType string

const (
	// TODO: ADD FOR CURRENT USE CASES
	LinearRaidType RaidType = "linear"
	Raid0RaidType  RaidType = "raid0"
	Raid1RaidType  RaidType = "raid1"
	Raid5RaidType  RaidType = "raid5"
	Raid6RaidType  RaidType = "raid6"
	Raid10RaidType RaidType = "raid10"
)

type StoragePoolSpecificOptions struct {
	//TODO: Decide the sub fields
}

type StoragePoolDeviceTypeIdentifier struct {
	Model      string `json:"model,omitempty"`
	PartNumber string `json:"partNumber,omitempty"`
}

type FilterOptions struct {
	Key     string `json:"key,omitempty"`
	Name    string `json:"name,omitempty"`
	State   string `json:"state,omitempty"`
	Exclude string `json:"exclude,omitempty"`
	Include string `json:"include,omitempty"`
}

type StoragePoolDataStorageCapabilities struct {
	AccessModes          []StoragePoolAccessMode            `json:"accessModes,omitempty"`
	ProvisioningPolicies []ProvisioningPolicy               `json:"provisioning_policies,omitempty"`
	Multipathing         StoragePoolMultipathingOption      `json:"multipathing,omitempty"`
	Compression          *[]StoragePoolCompressionAlgorithm `json:"compression,omitempty"`
	Deduplication        *bool                              `json:"deduplication,omitempty"`
}

type StoragePoolDataSecurityCapabilities struct {
	MediaEncryption        *[]StoragePoolMediaEncryptionAlgorithm `json:"mediaEncryption,omitempty"`
	DataSanitizationPolicy StoragePoolDataSanitizationPolicy  `json:"dataSanitizationPolicy,omitempty"`
}

type StoragePoolIOConnectivityCapabilities struct {
	AccessProtocols []StoragePoolAccessProtocol `json:"accessProtocols,omitempty"`
}

type StoragePoolIOPerformanceCapabilities struct {
	AverageIOOperationLatencyMicroseconds uint64                 `json:"averageIoOperationLatencyMicroseconds,omitempty"`
	MaxIOOperationsPerSecondPerTerabyte   uint64                 `json:"maxIoOperationsPerSecondPerTerabyte,omitempty"`
	StorageTier                           StoragePoolStorageTier `json:"storageTier,omitempty"`
}

type StoragePoolDataProtectionCapabilities struct {
	Replication bool `json:"replication,omitempty"`
	Backup      bool `json:"backup,omitempty"`
	Snapshots   bool `json:"snapshots,omitempty"`
}

type StoragePoolAccessMode string

const (
	// TODO: ADD FOR CURRENT USE CASES
	// ReadWriteOnce can be mounted in read/write mode to exactly 1 host
	ReadWriteOnce StoragePoolAccessMode = "ReadWriteOnce"
	// ReadOnlyMany can be mounted in read-only mode to many hosts
	ReadOnlyMany StoragePoolAccessMode = "ReadOnlyMany"
	// ReadWriteMany can be mounted in read/write mode to many hosts
	ReadWriteMany StoragePoolAccessMode = "ReadWriteMany"
)

type ProvisioningPolicy string

const (
	// TODO: ADD FOR CURRENT USE CASES
	ThickProvisioning ProvisioningPolicy = "thick"
	ThinProvisioning  ProvisioningPolicy = "thin"
)

type StoragePoolMultipathingOption string

const (
	// TODO: ADD FOR CURRENT USE CASES
	OnlineActiveMultipathingOption  StoragePoolMultipathingOption = "OnlineActive"
	OnlinePassiveMultipathingOption StoragePoolMultipathingOption = "OnlinePassive"
	NoneMultipathingOption          StoragePoolMultipathingOption = "OnlineActive"
)

type StoragePoolCompressionAlgorithm string

const (
	// TODO: ADD FOR CURRENT USE CASES
	HuffmanCompression StoragePoolCompressionAlgorithm = "huffman"
)

type StoragePoolMediaEncryptionAlgorithm string


type StoragePoolDataSanitizationPolicy string

const (
	// TODO: ADD FOR CURRENT USE CASES
	NoneDataSanitizationPolicy               StoragePoolDataSanitizationPolicy = "None"
	ClearDataSanitizationPolicy              StoragePoolDataSanitizationPolicy = "Clear"
	CryptographicEraseDataSanitizationPolicy StoragePoolDataSanitizationPolicy = "CryptographicErase"
)

type StoragePoolAccessProtocol string

const (
	// TODO: ADD FOR CURRENT USE CASES
	NVMeAccessProtocol           StoragePoolAccessProtocol = "NVMe"
	NVMeOverFabicsAccessProtocol StoragePoolAccessProtocol = "NVMeOverFabics"
)

type StoragePoolStorageTier string

const (
	// TODO: ADD FOR CURRENT USE CASES
	Platinum StoragePoolStorageTier = "P1"
	Gold     StoragePoolStorageTier = "P2"
	Silver   StoragePoolStorageTier = "P3"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=storagepools
type StoragePoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []StoragePool `json:"items"`
}
