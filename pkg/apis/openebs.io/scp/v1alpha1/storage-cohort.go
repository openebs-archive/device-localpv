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
	// +optional
	Spec StorageCohortSpec `json:"spec,omitempty"`

	// Most recently observed status of the cohort.
	// Populated by the cohort operator or cohort manager.
	// +optional
	Status StorageCohortStatus `json:"status,omitempty"`
}

// StorageCohortSpec describes the attributes that a cohort is created with.
type StorageCohortSpec struct {
	// NodeSelector is used to specify the cohort to be considered
	// +optional
	NodeSelector *metav1.LabelSelector `json:"nodeSelector,omitempty"`

	// CohortManager stores the details of the cohort manager responsible for
	// managing the cohort
	// +optional
	CohortManager CohortManagerDetails `json:"cohortManager,omitempty"`

	// StorageProvisioner contains list of all provisioners responsible for
	// the provisioning tasks for different storage solutions in the cohort
	// +optional
	StorageProvisioner []StorageProvisionerDetails `json:"storageProvisioner,omitempty"`
}

// StorageCohortStorageProvisionerDetails stores the different storage provisioners information
// which takes the job pf provisioning pools and volumes
type StorageProvisionerDetails struct {
	// StorageType represents the type of storage solution which the provisioner is responsible for
	// implementing
	// For example: lvm or device-localpv
	// +optional
	StorageType string `json:"storageType,omitempty"`

	// provisioner is the driver expected to handle this cohort.
	// This is an optionally-prefixed name, like a label key.
	// For example: "openebs.io/scp-lvm-provisioner" or "openebs.io/scp-device-provisioner".
	// +optional
	Provisioner string `json:"provisioner,omitempty"`
}

// CohortManagerDetails is information about the cohort manager managing
// the cohort.
type CohortManagerDetails struct {
	// ApiUrl is the cohort manager endpoint used for communicating with the cohort
	// +optional
	ApiUrl []string `json:"apiUrl,omitempty"`

	// HealthCheckUrl describes a health check endpoint of cohort manager against which
	// an api call is to be performed to determine whether it is alive or ready to receive traffic.
	// +optional
	HealthCheckUrl string `json:"healthCheckUrl,omitempty"`

	// TODO add authentication & security spec
}

// StorageCohortStatus is information about the current status of a storage cohort.
type StorageCohortStatus struct {
	// Conditions is an array of current observed cohort component's conditions.
	// +optional
	Conditions []ComponentCondition `json:"conditions,omitempty"`

	// Capabilities represent capabilities that a cohort consists of
	// +optional
	Capabilities Capabilities `json:"capabilities,omitempty"`
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

type ComponentConditionType string

// These are valid conditions of cohort.
// In the future, we can add more. The current set of conditions are:
// CohortComponentReady, CohortComponentSchedulable.
const (
	// CohortComponentReady means cohort component is healthy and ready to perform its task.
	CohortComponentReady ComponentConditionType = "Ready"
	// CohortComponentSchedulable means the cohort component is healthy and schedulable.
	CohortComponentSchedulable ComponentConditionType = "Schedulable"
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

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=storagecohorts
type StorageCohortList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []StorageVolume `json:"items"`
}
